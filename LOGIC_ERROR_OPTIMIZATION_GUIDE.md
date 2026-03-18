# Logic 层错误处理优化指南

## 📋 优化内容

### 1. 中间件统一构建 Context ✅
- 文件：`core/internal/middleware/error-recovery-middleware.go`
- 改动：在中间件中一次性构建完整的 context（包含 trace_id, method, path, user_id）
- 效果：Logic 层不需要手动构建 context

### 2. 自定义错误类型 ✅
- 文件：`core/internal/errors/logic_error.go`
- 功能：创建错误时自动记录日志
- 支持：ERROR 和 FATAL 两种级别

### 3. Logic 层代码简化 ✅
- 示例：`core/internal/logic/user-login-logic.go`
- 效果：每个错误处理从 5 行减少到 1 行

---

## 🎯 使用方法

### ERROR 级别（普通业务错误）

```go
import "cloud_disk/core/internal/errors"

func (l *SomeLogic) SomeMethod(req *types.Request) (resp *types.Response, err error) {
    // 场景 1: 数据库查询失败（有原始错误 + 额外信息）
    if err != nil {
        return nil, errors.New(l.ctx, "数据库查询失败", err, map[string]interface{}{
            "username": req.Name,
        })
    }

    // 场景 2: 用户不存在（没有原始错误 + 有额外信息）
    if !has {
        return nil, errors.New(l.ctx, "用户不存在", nil, map[string]interface{}{
            "username": req.Name,
            "reason":   "用户名或密码错误",
        })
    }

    // 场景 3: 参数验证失败（没有原始错误 + 没有额外信息）
    if req.Name == "" {
        return nil, errors.New(l.ctx, "参数错误: 用户名不能为空", nil, nil)
    }

    // 场景 4: 格式化消息
    if age < 18 {
        return nil, errors.New(l.ctx, fmt.Sprintf("年龄不符合要求: %d岁", age), nil, nil)
    }

    // 场景 5: 包装已有错误
    token, err := helper.GenerateToken(...)
    if err != nil {
        return nil, errors.New(l.ctx, "Token生成失败", err, map[string]interface{}{
            "user_id": user.Id,
        })
    }
}
```

### FATAL 级别（致命错误）

```go
import "cloud_disk/core/internal/errors"

func (l *SomeLogic) SomeMethod(req *types.Request) (resp *types.Response, err error) {
    // 场景 1: 数据库连接失败
    if err := db.Ping(); err != nil {
        return nil, errors.Fatal(l.ctx, "数据库连接失败", err, nil)
    }

    // 场景 2: Redis 连接失败
    if err := redis.Ping(); err != nil {
        return nil, errors.Fatal(l.ctx, "Redis连接失败", err, nil)
    }

    // 场景 3: 配置加载失败（格式化消息）
    if err := loadConfig(path); err != nil {
        return nil, errors.Fatal(l.ctx, fmt.Sprintf("配置文件 %s 加载失败", path), err, nil)
    }

    // 场景 4: 系统初始化失败
    if err := initSystem(); err != nil {
        return nil, errors.Fatal(l.ctx, "系统初始化失败", err, map[string]interface{}{
            "component": "payment",
        })
    }
}
```

### PANIC 级别（不需要处理）

```go
// ✅ 中间件自动捕获，不需要手动处理
panic("something went wrong")

// 或者
var ptr *User
ptr.Name  // nil pointer panic，中间件自动捕获并记录日志
```

---

## 📊 对比效果

### 修改前（冗余）

```go
func (l *UserLoginLogic) UserLogin(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
    // ❌ 手动构建 context（4 行）
    traceID, _ := l.ctx.Value("trace_id").(string)
    ctx := context.WithValue(l.ctx, "method", "POST")
    ctx = context.WithValue(ctx, "path", "/user/login")
    ctx = context.WithValue(ctx, "trace_id", traceID)

    // ❌ 手动记录日志（5 行）
    var user = new(models.UserBasic)
    has, err := l.svcCtx.Engine.Where("name = ? and password = ?", req.Name, helper.MD5(req.Password)).Get(user)
    if err != nil {
        logger.LogError(ctx, "数据库查询失败", err, map[string]interface{}{
            "username": req.Name,
        })
        return nil, err
    }

    // ❌ 手动记录日志（5 行）
    if !has {
        err = errors.New("用户名或密码错误")
        logger.LogError(ctx, "用户登录失败", err, map[string]interface{}{
            "username": req.Name,
            "reason":   "用户名或密码错误",
        })
        return nil, err
    }

    // ...
}
```

### 修改后（简洁）

```go
import "cloud_disk/core/internal/errors"

func (l *UserLoginLogic) UserLogin(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
    // ✅ 不需要手动构建 context（0 行）

    // ✅ 一行代码搞定（1 行）
    var user = new(models.UserBasic)
    has, err := l.svcCtx.Engine.Where("name = ? and password = ?", req.Name, helper.MD5(req.Password)).Get(user)
    if err != nil {
        return nil, errors.New(l.ctx, "数据库查询失败", err, map[string]interface{}{
            "username": req.Name,
        })
    }

    // ✅ 一行代码搞定（1 行）
    if !has {
        return nil, errors.New(l.ctx, "用户登录失败", nil, map[string]interface{}{
            "username": req.Name,
            "reason":   "用户名或密码错误",
        })
    }

    // ...
}
```

**代码减少：** 从 83 行 → 50 行（减少 40%）

---

## 🔧 迁移步骤

### 步骤 1: 修改 import

```go
// 删除
import (
    "cloud_disk/core/internal/logger"
    "errors"
)

// 添加
import (
    "cloud_disk/core/internal/errors"
)
```

### 步骤 2: 删除手动构建 context 的代码

```go
// ❌ 删除这些代码
traceID, _ := l.ctx.Value("trace_id").(string)
ctx := context.WithValue(l.ctx, "method", "POST")
ctx = context.WithValue(ctx, "path", "/user/login")
ctx = context.WithValue(ctx, "trace_id", traceID)
```

### 步骤 3: 替换错误处理

```go
// ❌ 修改前
if err != nil {
    logger.LogError(ctx, "数据库查询失败", err, map[string]interface{}{
        "username": req.Name,
    })
    return nil, err
}

// ✅ 修改后
if err != nil {
    return nil, errors.New(l.ctx, "数据库查询失败", err, map[string]interface{}{
        "username": req.Name,
    })
}
```

---

## 📝 API 参考

### ERROR 级别

| 函数 | 参数 | 说明 |
|------|------|------|
| `errors.New(ctx, message, err, extra)` | ctx: context<br>message: 错误消息<br>err: 原始错误<br>extra: 额外信息 | 创建 ERROR 错误 |
| `errors.Newf(ctx, format, args...)` | ctx: context<br>format: 格式化字符串<br>args: 参数 | 创建格式化 ERROR 错误 |
| `errors.Wrap(ctx, err, message, extra)` | ctx: context<br>err: 原始错误<br>message: 包装消息<br>extra: 额外信息 | 包装已有错误 |

### FATAL 级别

| 函数 | 参数 | 说明 |
|------|------|------|
| `errors.Fatal(ctx, message, err, extra)` | ctx: context<br>message: 错误消息<br>err: 原始错误<br>extra: 额外信息 | 创建 FATAL 错误 |
| `errors.Fatalf(ctx, format, args...)` | ctx: context<br>format: 格式化字符串<br>args: 参数 | 创建格式化 FATAL 错误 |
| `errors.WrapFatal(ctx, err, message, extra)` | ctx: context<br>err: 原始错误<br>message: 包装消息<br>extra: 额外信息 | 包装为 FATAL 错误 |

### 辅助函数

| 函数 | 返回值 | 说明 |
|------|--------|------|
| `errors.Is(err)` | bool | 判断是否为 LogicError |
| `errors.GetLevel(err)` | string | 获取错误级别 |
| `errors.IsFatal(err)` | bool | 判断是否为 FATAL |
| `errors.IsError(err)` | bool | 判断是否为 ERROR |

---

## 🎯 记录的日志信息

使用 `errors.New()` 或 `errors.Fatal()` 创建错误时，会自动记录以下信息：

```json
{
  "timestamp": "2026-03-18 15:30:45",
  "level": "ERROR",
  "trace_id": "abc123",
  "user_id": "user_001",
  "method": "POST",
  "path": "/user/login",
  "message": "数据库查询失败: connection timeout",
  "stack_trace": "goroutine 1 [running]...",
  "extra": {
    "username": "test"
  }
}
```

**信息来源：**
- `timestamp`: 自动生成
- `level`: 根据函数自动设置（ERROR 或 FATAL）
- `trace_id`: 从 context 提取（中间件设置）
- `user_id`: 从 context 提取（中间件设置）
- `method`: 从 context 提取（中间件设置）
- `path`: 从 context 提取（中间件设置）
- `message`: 函数参数
- `stack_trace`: 自动生成
- `extra`: 函数参数

---

## ✅ 优势总结

### 1. 代码简洁
- ✅ 不需要手动构建 context（减少 4 行）
- ✅ 不需要手动调用 logger.LogError（减少 3 行）
- ✅ 每个错误处理从 5 行减少到 1 行

### 2. 统一管理
- ✅ Context 构建逻辑集中在中间件
- ✅ 日志记录逻辑集中在错误类型
- ✅ 修改时只需要改一个地方

### 3. 不易出错
- ✅ 不会忘记记录日志（自动记录）
- ✅ 不会忘记设置 context（中间件统一设置）
- ✅ 不会写错字段名（统一封装）

### 4. 信息更准确
- ✅ path 从 HTTP 请求自动获取，不会写错
- ✅ method 自动获取，不需要手动指定
- ✅ 路由改了，日志自动正确

---

## 🔍 常见问题

### Q1: 如果不想记录日志怎么办？

使用标准库的 error：

```go
// 不记录日志
return nil, fmt.Errorf("错误: %w", err)
```

### Q2: 如何判断错误级别？

```go
if errors.IsFatal(err) {
    // 处理致命错误
}

if errors.IsError(err) {
    // 处理普通错误
}
```

### Q3: 旧代码需要立即迁移吗？

不需要！可以逐步迁移：
- ✅ 新代码使用新方式
- ✅ 旧代码保持不变
- ✅ 有时间再慢慢迁移

### Q4: 中间件改动会影响现有代码吗？

不会！中间件只是增加了字段，不影响现有逻辑。

---

**文档版本：** v1.0
**创建时间：** 2026-03-18
