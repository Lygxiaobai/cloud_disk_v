# Casbin RBAC 实施说明

这份文档用于在当前 `cloud_disk` 项目中引入 Casbin，实现基于 RBAC 的权限管理。

目标：

- 用户登录后，除了身份信息，还带有角色信息
- 受保护接口在 JWT 鉴权成功后，再做角色权限判断
- 用 Casbin 统一管理：
  - 哪个角色可以访问哪些接口
  - 哪个角色可以执行哪些 HTTP 方法
- 后续如果要细分成管理员、普通用户、只读用户，都能继续扩展

## 1. 当前项目现状

你当前项目已经有：

- JWT 登录态
- `Authorization` 请求头
- `AuthMiddleware` 负责解析 token
- handler 里通过 `UserIdentity` 取当前登录用户

当前代码位置：

- `core/internal/middleware/auth-middleware.go`
- `core/internal/helper/helper.go`
- `core/internal/define/define.go`
- `core/internal/logic/user-login-logic.go`
- `core/internal/svc/service-context.go`

也就是说，你现在已经有“认证”，但还没有“授权”。

认证解决的是：

- 你是谁

授权解决的是：

- 你能做什么

Casbin 就是用来做第二层“你能做什么”的。

## 2. 推荐实现方案

推荐采用：

- JWT 存用户基础身份 + 角色
- Casbin 负责做角色到接口权限的映射
- 后端新增一层授权中间件

整体链路：

1. 用户登录
2. 后端查出用户角色
3. 生成 token 时把角色写进 JWT
4. `AuthMiddleware` 解析 token
5. 在 request header 或 context 中写入角色
6. `CasbinMiddleware` 根据：
   - 角色
   - 请求路径
   - 请求方法
   做 `Enforce`
7. 通过则放行，失败则返回 `403`

## 3. 你要改哪些文件

建议改这些文件：

- `core/internal/models/user_basic.go`
- `core/internal/define/define.go`
- `core/internal/helper/helper.go`
- `core/internal/logic/user-login-logic.go`
- `core/internal/logic/refresh-token-logic.go`
- `core/internal/middleware/auth-middleware.go`
- `core/internal/svc/service-context.go`
- `core/internal/config/config.go`
- `core/etc/core-api.yaml`
- 新增 `core/internal/middleware/casbin-middleware.go`
- 新增 `core/internal/authorization/model.conf`
- 新增 `core/internal/authorization/policy.csv`

如果你准备做动态策略持久化，也可以再加数据库 adapter。但第一版建议先用本地文件模型和策略，最容易跑通。

## 4. 第一步：给用户补角色字段

文件：

- `core/internal/models/user_basic.go`

当前模型只有：

- `Id`
- `Identity`
- `Name`
- `Password`
- `Email`

建议补一个角色字段：

```go
package models

type UserBasic struct {
    Id       int
    Identity string
    Name     string
    Password string
    Email    string
    Role     string
}

func (u UserBasic) TableName() string {
    return "user_basic"
}
```

### 数据库也要同步加列

例如：

```sql
ALTER TABLE user_basic ADD COLUMN role VARCHAR(32) DEFAULT 'user';
```

你可以先准备几类角色：

- `admin`
- `user`
- `readonly`

## 5. 第二步：JWT 加入角色字段

文件：

- `core/internal/define/define.go`

当前 `UserClaim` 里没有角色字段，建议改成：

```go
package define

import "github.com/golang-jwt/jwt/v4"

type UserClaim struct {
    ID       int
    Identity string
    Name     string
    Role     string
    jwt.StandardClaims
}
```

这样 token 里就能带角色。

## 6. 第三步：生成 token 时带 role

文件：

- `core/internal/helper/helper.go`

当前 `GenerateToken` 只接收：

- `id`
- `identity`
- `name`
- `expireTime`

建议改成：

```go
func GenerateToken(id int, identity string, name string, role string, expireTime int) (string, error) {
    uc := define.UserClaim{
        ID:       id,
        Identity: identity,
        Name:     name,
        Role:     role,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(time.Second * time.Duration(expireTime)).Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, uc)
    tokenString, err := token.SignedString([]byte(define.JwtKey))
    if err != nil {
        return "", err
    }
    return tokenString, nil
}
```

说明：

- 这里只是把角色一起写入 JWT
- `AnalyzeToken()` 逻辑一般不用改，只要 claim 结构变了，它自然能读出 `Role`

## 7. 第四步：登录和刷新 token 逻辑同步带 role

文件：

- `core/internal/logic/user-login-logic.go`
- `core/internal/logic/refresh-token-logic.go`

### 7.1 登录时带 role

在登录成功后：

```go
token, err := helper.GenerateToken(user.Id, user.Identity, user.Name, user.Role, define.TokenExpireTime)
refreshToken, err := helper.GenerateToken(user.Id, user.Identity, user.Name, user.Role, define.RefreshTokenExpireTime)
```

### 7.2 刷新 token 时带 role

当前刷新 token 是从 refresh token 解析 claim 再生成新 token，所以这里改成：

```go
token, err := helper.GenerateToken(uc.ID, uc.Identity, uc.Name, uc.Role, define.TokenExpireTime)
refreshToken, err := helper.GenerateToken(uc.ID, uc.Identity, uc.Name, uc.Role, define.RefreshTokenExpireTime)
```

说明：

- 这样 refresh 出来的 token 不会丢角色信息

## 8. 第五步：AuthMiddleware 写入 Role

文件：

- `core/internal/middleware/auth-middleware.go`

当前中间件成功后只写了：

- `UserId`
- `UserIdentity`
- `UserName`

建议再补一行：

```go
r.Header.Set("UserRole", uc.Role)
```

完整意思就是：

- JWT 验证通过后
- 把用户角色也传递给后面的中间件和 handler

## 9. 第六步：新增 Casbin 中间件

文件：

- `core/internal/middleware/casbin-middleware.go`

建议代码：

```go
package middleware

import (
    "net/http"

    "github.com/casbin/casbin/v2"
)

type CasbinMiddleware struct {
    enforcer *casbin.Enforcer
}

func NewCasbinMiddleware(enforcer *casbin.Enforcer) *CasbinMiddleware {
    return &CasbinMiddleware{enforcer: enforcer}
}

func (m *CasbinMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        role := r.Header.Get("UserRole")
        if role == "" {
            w.WriteHeader(http.StatusForbidden)
            w.Write([]byte("forbidden: role missing"))
            return
        }

        obj := r.URL.Path
        act := r.Method

        ok, err := m.enforcer.Enforce(role, obj, act)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            w.Write([]byte(err.Error()))
            return
        }
        if !ok {
            w.WriteHeader(http.StatusForbidden)
            w.Write([]byte("forbidden"))
            return
        }

        next(w, r)
    }
}
```

说明：

- `role` 是主体 subject
- `obj` 是资源 object，一般就是请求路径
- `act` 是动作 action，一般就是 HTTP Method
- 这是典型的 RBAC + path + method 组合

## 10. 第七步：ServiceContext 注入 Casbin Enforcer

文件：

- `core/internal/svc/service-context.go`
- `core/internal/config/config.go`
- `core/etc/core-api.yaml`

### 10.1 先给配置加 Casbin 路径

`core/internal/config/config.go`：

```go
package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
    rest.RestConf
    Mysql struct {
        DataSource string
    }
    Redis struct {
        Addr     string
        Password string
        DB       int
    }
    Casbin struct {
        ModelPath  string
        PolicyPath string
    }
}
```

### 10.2 YAML 加配置

`core/etc/core-api.yaml`：

```yaml
Casbin:
  ModelPath: internal/authorization/model.conf
  PolicyPath: internal/authorization/policy.csv
```

### 10.3 ServiceContext 初始化 Enforcer

建议先加字段：

```go
type ServiceContext struct {
    Config config.Config
    Engine *xorm.Engine
    RDB    *redis.Client
    Auth   rest.Middleware
    Casbin rest.Middleware
}
```

然后初始化：

```go
import (
    "cloud_disk/core/internal/config"
    "cloud_disk/core/internal/middleware"
    "cloud_disk/core/internal/models"
    "github.com/casbin/casbin/v2"
    fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
    "github.com/redis/go-redis/v9"
    "github.com/zeromicro/go-zero/rest"
    "xorm.io/xorm"
)

func NewServiceContext(c config.Config) *ServiceContext {
    adapter := fileadapter.NewAdapter(c.Casbin.PolicyPath)
    enforcer, err := casbin.NewEnforcer(c.Casbin.ModelPath, adapter)
    if err != nil {
        panic(err)
    }

    return &ServiceContext{
        Config: c,
        Engine: models.Init(c.Mysql.DataSource),
        RDB:    models.InitRedis(c.Redis.Addr),
        Auth:   middleware.NewAuthMiddleware().Handle,
        Casbin: middleware.NewCasbinMiddleware(enforcer).Handle,
    }
}
```

说明：

- 这里第一版先用文件策略，不依赖数据库 adapter
- 先把逻辑跑通最重要

## 11. 第八步：Casbin 模型文件

新增文件：

- `core/internal/authorization/model.conf`

内容建议：

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act
```

说明：

- `sub`：角色
- `obj`：请求路径
- `act`：请求方法
- `keyMatch2` 支持带路径参数的接口，比如 `/share/file/detail/:identity`

## 12. 第九步：Casbin 策略文件

新增文件：

- `core/internal/authorization/policy.csv`

先给一个起步版本：

```csv
p, admin, /user/file/list, GET
p, admin, /user/folder/children/:id, GET
p, admin, /user/folder/path/:identity, GET
p, admin, /user/file/move, PUT
p, admin, /user/file/delete, DELETE
p, admin, /user/file/name/update, PUT
p, admin, /user/folder/create, PUT
p, admin, /user/repository/save, POST
p, admin, /share/basic/create, POST
p, admin, /share/file/save, POST
p, admin, /file/upload, POST
p, admin, /file/upload/multipart, POST

p, user, /user/file/list, GET
p, user, /user/folder/children/:id, GET
p, user, /user/folder/path/:identity, GET
p, user, /user/file/move, PUT
p, user, /user/file/delete, DELETE
p, user, /user/file/name/update, PUT
p, user, /user/folder/create, PUT
p, user, /user/repository/save, POST
p, user, /share/basic/create, POST
p, user, /share/file/save, POST
p, user, /file/upload, POST
p, user, /file/upload/multipart, POST

p, readonly, /user/file/list, GET
p, readonly, /user/folder/children/:id, GET
p, readonly, /user/folder/path/:identity, GET

# 用户角色继承关系
# 如果后面你想让某些角色继承 admin 或 user，可以继续往下加 g 规则
```

## 13. 第十步：路由如何挂载 Casbin 中间件

你现在鉴权接口已经通过 `Auth` 处理中间件保护。

建议方式：

- 保留 `Auth`
- 再叠加 `Casbin`

也就是说：

```go
rest.WithMiddlewares(
    []rest.Middleware{serverCtx.Auth, serverCtx.Casbin},
    []rest.Route{ ... },
)
```

这样顺序就是：

1. 先认证 JWT
2. 再校验角色权限
3. 最后才到 handler

## 14. 建议你优先保护哪些接口

第一版优先保护这些用户私有接口：

- `/user/file/list`
- `/user/folder/children/:id`
- `/user/folder/path/:identity`
- `/user/file/name/update`
- `/user/file/delete`
- `/user/file/move`
- `/user/folder/create`
- `/user/repository/save`
- `/file/upload`
- `/file/upload/multipart`
- `/share/basic/create`
- `/share/file/save`

公开接口不用走 Casbin：

- `/user/login`
- `/user/register`
- `/mail/code/send/register`
- `/refresh/token`
- `/share/file/detail/:identity`

## 15. 你自己手敲的建议顺序

建议顺序：

1. 给 `user_basic` 表和模型加 `Role`
2. 给 JWT claim 加 `Role`
3. 改登录/刷新 token 逻辑
4. 改 `AuthMiddleware` 把 `UserRole` 放进 header
5. 加 `model.conf` 和 `policy.csv`
6. 写 `CasbinMiddleware`
7. 改 `ServiceContext` 注入 Enforcer
8. 把私有路由从 `Auth` 升级成 `Auth + Casbin`

## 16. 你可以先这样测试

### 16.1 数据库给用户设置角色

```sql
UPDATE user_basic SET role = 'admin' WHERE name = 'zhangsan';
UPDATE user_basic SET role = 'user' WHERE name = 'frontend_demo_user';
```

### 16.2 登录拿 token

```powershell
$body = @{ name = 'frontend_demo_user'; password = 'demo123456' } | ConvertTo-Json
Invoke-RestMethod -Uri 'http://127.0.0.1:8888/user/login' -Method Post -ContentType 'application/json' -Body $body
```

### 16.3 用 token 请求受保护接口

```powershell
$body = @{ name = 'frontend_demo_user'; password = 'demo123456' } | ConvertTo-Json
$login = Invoke-RestMethod -Uri 'http://127.0.0.1:8888/user/login' -Method Post -ContentType 'application/json' -Body $body
$headers = @{ Authorization = $login.token }
Invoke-RestMethod -Uri 'http://127.0.0.1:8888/user/file/list?id=0&page=1&size=20' -Method Get -Headers $headers
```

如果 `role` 没权限，就应该返回：

```text
403 forbidden
```

## 17. 最后的建议

第一版 Casbin 不要一上来做太复杂。

建议先做到：

- 角色写进 token
- 中间件能按 `role + path + method` 判断
- 管理员和普通用户两类角色跑通

等这一版稳定以后，再考虑：

- 后台管理动态编辑权限
- Casbin policy 存数据库
- 按资源所有者做细粒度校验