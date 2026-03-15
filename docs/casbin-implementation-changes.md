# Casbin RBAC 实施更改记录

本文档记录了在 cloud_disk 项目中实施 Casbin RBAC 权限管理的所有代码更改。

## 更改概述

本次实施引入了基于 Casbin 的 RBAC（基于角色的访问控制）系统，实现了：
- 用户角色管理（admin、user、readonly）
- JWT Token 中包含角色信息
- 基于角色的接口访问权限控制
- 灵活的权限策略配置

## 1. 数据模型更改

### 1.1 用户基础模型 (core/internal/models/user_basic.go)

**更改内容：**
- 在 `UserBasic` 结构体中添加了 `Role` 字段

```go
type UserBasic struct {
    Id       int
    Identity string
    Name     string
    Password string
    Email    string
    Role     string  // 新增：用户角色字段
}
```

**说明：**
- Role 字段用于存储用户的角色信息（如 admin、user、readonly）
- 需要在数据库中执行相应的 ALTER TABLE 语句添加该列

**数据库迁移 SQL：**
```sql
ALTER TABLE user_basic ADD COLUMN role VARCHAR(32) DEFAULT 'user';
```

## 2. JWT 认证更改

### 2.1 JWT Claims 定义 (core/internal/define/define.go)

**更改内容：**
- 在 `UserClaim` 结构体中添加了 `Role` 字段

```go
type UserClaim struct {
    ID       int
    Identity string
    Name     string
    Role     string  // 新增：角色字段
    jwt.StandardClaims
}
```

**说明：**
- 将用户角色信息编码到 JWT Token 中
- 使得每次请求都能携带用户的角色信息

### 2.2 Token 生成函数 (core/internal/helper/helper.go)

**更改内容：**
- `GenerateToken` 函数添加了 `role` 参数

```go
func GenerateToken(id int, identity string, name string, role string, expireTime int) (string, error) {
    uc := define.UserClaim{
        ID:       id,
        Identity: identity,
        Name:     name,
        Role:     role,  // 新增：将角色写入 JWT
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

**说明：**
- 生成 Token 时将用户角色信息一并编码
- `AnalyzeToken` 函数无需修改，会自动解析出 Role 字段

## 3. 业务逻辑更改

### 3.1 用户登录逻辑 (core/internal/logic/user-login-logic.go)

**更改内容：**
- 在生成 Token 和 RefreshToken 时传入用户角色

```go
// 生成Token
token, err := helper.GenerateToken(user.Id, user.Identity, user.Name, user.Role, define.TokenExpireTime)
if err != nil {
    return nil, err
}
// 生成refreshToken
refreshToken, err := helper.GenerateToken(user.Id, user.Identity, user.Name, user.Role, define.RefreshTokenExpireTime)
```

**说明：**
- 用户登录成功后，从数据库查询的用户信息中获取角色
- 将角色信息写入 Token 和 RefreshToken

### 3.2 Token 刷新逻辑 (core/internal/logic/refresh-token-logic.go)

**更改内容：**
- 在刷新 Token 时保持角色信息

```go
// 根据 refreshToken 生成新的一组 token
token, err := helper.GenerateToken(uc.ID, uc.Identity, uc.Name, uc.Role, define.TokenExpireTime)
if err != nil {
    return nil, err
}
refreshToken, err := helper.GenerateToken(uc.ID, uc.Identity, uc.Name, uc.Role, define.RefreshTokenExpireTime)
if err != nil {
    return nil, err
}
```

**说明：**
- 从旧的 RefreshToken 中解析出用户信息（包括角色）
- 生成新的 Token 时保持角色信息不变

## 4. 中间件更改

### 4.1 认证中间件 (core/internal/middleware/auth-middleware.go)

**更改内容：**
- 在请求头中添加 `UserRole` 字段

```go
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // ... JWT 验证逻辑 ...

        // 解析成功后，在 Header 中设置用户信息
        r.Header.Set("UserId", string(uc.Id))
        r.Header.Set("UserIdentity", uc.Identity)
        r.Header.Set("UserName", uc.Name)
        r.Header.Set("UserRole", uc.Role)  // 新增：传递角色信息

        next(w, r)
    }
}
```

**说明：**
- JWT 验证通过后，将用户角色信息写入请求头
- 后续的 Casbin 中间件可以从请求头中获取角色信息

### 4.2 Casbin 授权中间件 (core/internal/middleware/casbin-middleware.go)

**新增文件：**
- 创建了全新的 Casbin 中间件

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

        ok, err := m.enforcer.Enforce(role, r.URL.Path, r.Method)
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

**说明：**
- 从请求头中获取用户角色
- 使用 Casbin Enforcer 进行权限校验
- 校验维度：角色 (role) + 请求路径 (r.URL.Path) + 请求方法 (r.Method)
- 权限不足时返回 403 Forbidden

## 5. 配置更改

### 5.1 配置结构 (core/internal/config/config.go)

**更改内容：**
- 添加了 Casbin 配置结构

```go
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
    Casbin struct {  // 新增：Casbin 配置
        ModelPath  string
        PolicyPath string
    }
}
```

**说明：**
- ModelPath：Casbin 模型文件路径
- PolicyPath：Casbin 策略文件路径

### 5.2 配置文件 (core/etc/core-api.yaml)

**更改内容：**
- 添加了 Casbin 配置项

```yaml
Casbin:
  ModelPath: internal/authorization/model.conf
  PolicyPath: internal/authorization/policy.csv
```

**说明：**
- 指定了 Casbin 模型和策略文件的相对路径

## 6. 服务上下文更改

### 6.1 ServiceContext (core/internal/svc/service-context.go)

**更改内容：**
- 添加了 Casbin 中间件字段
- 初始化 Casbin Enforcer

```go
type ServiceContext struct {
    Config config.Config
    Engine *xorm.Engine
    RDB    *redis.Client
    Auth   rest.Middleware
    Casbin rest.Middleware  // 新增：Casbin 中间件
}

func NewServiceContext(c config.Config) *ServiceContext {
    // 新增：初始化 Casbin
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
        Casbin: middleware.NewCasbinMiddleware(enforcer).Handle,  // 新增
    }
}
```

**说明：**
- 使用文件适配器加载策略文件
- 创建 Casbin Enforcer 实例
- 将 Casbin 中间件注入到服务上下文中

## 7. 路由配置更改

### 7.1 路由注册 (core/internal/handler/routes.go)

**更改内容：**
- 在受保护的路由上同时应用 Auth 和 Casbin 中间件

```go
server.AddRoutes(
    rest.WithMiddlewares(
        []rest.Middleware{serverCtx.Auth, serverCtx.Casbin},  // 新增：叠加 Casbin 中间件
        []rest.Route{
            // 文件上传
            {Method: http.MethodPost, Path: "/file/upload", Handler: FileUploadHandler(serverCtx)},
            {Method: http.MethodPost, Path: "/file/upload/multipart", Handler: FileUploadMultipartHandler(serverCtx)},

            // 分享功能
            {Method: http.MethodPost, Path: "/share/basic/create", Handler: ShareBasicCreateHandler(serverCtx)},
            {Method: http.MethodPost, Path: "/share/file/save", Handler: ShareFileSaveHandler(serverCtx)},

            // 用户文件管理
            {Method: http.MethodDelete, Path: "/user/file/delete", Handler: UserFileDeleteHandler(serverCtx)},
            {Method: http.MethodGet, Path: "/user/file/list", Handler: UserFileListHandler(serverCtx)},
            {Method: http.MethodPut, Path: "/user/file/move", Handler: UserFileMoveHandler(serverCtx)},
            {Method: http.MethodPut, Path: "/user/file/name/update", Handler: UserFileNameUpdateHandler(serverCtx)},

            // 文件夹管理
            {Method: http.MethodGet, Path: "/user/folder/children/:id", Handler: UserFolderChildrenHandler(serverCtx)},
            {Method: http.MethodPut, Path: "/user/folder/create", Handler: UserFolderCreateHandler(serverCtx)},
            {Method: http.MethodGet, Path: "/user/folder/path/:identity", Handler: UserFolderPathHandler(serverCtx)},

            // 仓库保存
            {Method: http.MethodPost, Path: "/user/repository/save", Handler: UserRepositorySaveHandler(serverCtx)},
        }...,
    ),
)
```

**说明：**
- 中间件执行顺序：Auth（认证）→ Casbin（授权）→ Handler（业务逻辑）
- 公开接口（如登录、注册）不应用 Casbin 中间件

## 8. Casbin 配置文件

### 8.1 模型文件 (core/internal/authorization/model.conf)

**新增文件：**

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

**说明：**
- `sub`：主体（角色）
- `obj`：对象（请求路径）
- `act`：动作（HTTP 方法）
- `keyMatch2`：支持路径参数匹配，如 `/user/folder/children/:id`
- `g`：角色继承关系（当前未使用，预留扩展）

### 8.2 策略文件 (core/internal/authorization/policy.csv)

**新增文件：**

```csv
# admin 角色权限（完全访问）
p, admin, /user/file/list, GET
p, admin, /user/folder/children/:id, GET
p, admin, /user/folder/path/:identity, GET
p, admin, /user/file/name/update, PUT
p, admin, /user/file/delete, DELETE
p, admin, /user/file/move, PUT
p, admin, /user/folder/create, PUT
p, admin, /user/repository/save, POST
p, admin, /share/basic/create, POST
p, admin, /share/file/save, POST
p, admin, /file/upload, POST
p, admin, /file/upload/multipart, POST

# user 角色权限（标准用户权限）
p, user, /user/file/list, GET
p, user, /user/folder/children/:id, GET
p, user, /user/folder/path/:identity, GET
p, user, /user/file/name/update, PUT
p, user, /user/file/delete, DELETE
p, user, /user/file/move, PUT
p, user, /user/folder/create, PUT
p, user, /user/repository/save, POST
p, user, /share/basic/create, POST
p, user, /share/file/save, POST
p, user, /file/upload, POST
p, user, /file/upload/multipart, POST

# readonly 角色权限（只读权限）
p, readonly, /user/file/list, GET
p, readonly, /user/folder/children/:id, GET
p, readonly, /user/folder/path/:identity, GET
```

**说明：**
- 定义了三种角色：admin、user、readonly
- admin 和 user 当前拥有相同的权限（可根据需求调整）
- readonly 只能访问查询类接口，不能进行修改操作
- 格式：`p, 角色, 路径, HTTP方法`

## 9. 依赖包更改

**新增依赖：**
```
github.com/casbin/casbin/v2
github.com/casbin/casbin/v2/persist/file-adapter
```

**安装命令：**
```bash
go get github.com/casbin/casbin/v2
```

## 10. 测试建议

### 10.1 数据库准备

```sql
-- 为测试用户设置不同角色
UPDATE user_basic SET role = 'admin' WHERE name = 'admin_user';
UPDATE user_basic SET role = 'user' WHERE name = 'normal_user';
UPDATE user_basic SET role = 'readonly' WHERE name = 'readonly_user';
```

### 10.2 测试流程

1. **登录获取 Token**
```powershell
$body = @{ name = 'normal_user'; password = 'password123' } | ConvertTo-Json
$login = Invoke-RestMethod -Uri 'http://127.0.0.1:8888/user/login' -Method Post -ContentType 'application/json' -Body $body
$token = $login.token
```

2. **测试有权限的接口**
```powershell
$headers = @{ Authorization = $token }
Invoke-RestMethod -Uri 'http://127.0.0.1:8888/user/file/list?id=0&page=1&size=20' -Method Get -Headers $headers
# 应该返回正常数据
```

3. **测试无权限的接口（使用 readonly 用户）**
```powershell
$body = @{ name = 'readonly_user'; password = 'password123' } | ConvertTo-Json
$login = Invoke-RestMethod -Uri 'http://127.0.0.1:8888/user/login' -Method Post -ContentType 'application/json' -Body $body
$headers = @{ Authorization = $login.token }
Invoke-RestMethod -Uri 'http://127.0.0.1:8888/user/file/delete' -Method Delete -Headers $headers
# 应该返回 403 Forbidden
```

## 11. 权限策略说明

### 11.1 当前角色定义

| 角色 | 说明 | 权限范围 |
|------|------|----------|
| admin | 管理员 | 所有接口的完全访问权限 |
| user | 普通用户 | 所有接口的完全访问权限（与 admin 相同） |
| readonly | 只读用户 | 仅能访问查询类接口 |

### 11.2 受保护的接口

**查询类接口：**
- GET /user/file/list
- GET /user/folder/children/:id
- GET /user/folder/path/:identity

**修改类接口：**
- PUT /user/file/name/update
- DELETE /user/file/delete
- PUT /user/file/move
- PUT /user/folder/create
- POST /user/repository/save
- POST /share/basic/create
- POST /share/file/save
- POST /file/upload
- POST /file/upload/multipart

**公开接口（无需权限）：**
- POST /user/login
- POST /user/register
- POST /mail/code/send/register
- PUT /refresh/token
- GET /user/detail
- GET /share/file/detail/:identity

## 12. 后续扩展建议

### 12.1 动态权限管理

当前使用文件存储策略，后续可以：
- 使用数据库适配器（如 gorm-adapter）
- 实现后台管理界面动态编辑权限
- 支持运行时热更新权限策略

### 12.2 细粒度权限控制

可以进一步实现：
- 资源所有者校验（用户只能操作自己的文件）
- 基于用户组的权限管理
- 基于资源属性的访问控制（ABAC）

### 12.3 角色继承

利用 Casbin 的角色继承功能：
```csv
g, admin, user  # admin 继承 user 的所有权限
```

## 13. 总结

本次实施完成了以下目标：

✅ 用户模型支持角色字段
✅ JWT Token 包含角色信息
✅ 认证中间件传递角色信息
✅ Casbin 授权中间件实现权限控制
✅ 配置文件支持 Casbin 配置
✅ 服务上下文注入 Casbin Enforcer
✅ 路由层面应用双重中间件（认证 + 授权）
✅ 定义了三种角色的权限策略
✅ 支持路径参数的权限匹配

整个系统现在具备了完整的 RBAC 权限管理能力，可以灵活地控制不同角色对不同接口的访问权限。
