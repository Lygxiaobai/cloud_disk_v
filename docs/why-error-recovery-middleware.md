# 为什么需要错误恢复中间件？

## 核心原因：防止程序崩溃

在 Go 语言中，如果代码中发生 `panic` 而没有被 `recover()` 捕获，**整个程序会崩溃**。

## 对比演示

### 场景：代码中有个 bug

```go
func HandleUpload(w http.ResponseWriter, r *http.Request) {
    var user *User  // nil

    // 💥 这里会 panic
    log.Println(user.Name)
}
```

### 没有中间件

```
用户 A 访问 /upload
    ↓
触发 panic
    ↓
💥 整个程序崩溃
    ↓
所有用户都无法访问
    ↓
需要手动重启服务
```

**后果：**
- ❌ 服务完全不可用
- ❌ 所有用户受影响
- ❌ 需要人工介入重启
- ❌ 可能丢失正在处理的请求

### 有中间件

```
用户 A 访问 /upload
    ↓
触发 panic
    ↓
✅ 中间件捕获 panic
    ↓
✅ 记录详细日志
    ↓
✅ 返回 500 错误给用户 A
    ↓
✅ 其他用户不受影响
```

**好处：**
- ✅ 服务继续运行
- ✅ 只影响当前请求
- ✅ 自动记录错误日志
- ✅ 其他用户正常使用

## 实际例子

### 例子 1：空指针引用

```go
// 没有中间件
func (l *Logic) Upload(req *Request) error {
    file := getFile(req.FileID)  // 返回 nil
    size := file.Size  // 💥 panic: nil pointer
    // 程序崩溃！
}

// 有中间件
func (l *Logic) Upload(req *Request) error {
    file := getFile(req.FileID)  // 返回 nil
    size := file.Size  // 💥 panic
    // ✅ 中间件捕获，记录日志，返回 500
    // ✅ 服务继续运行
}
```

### 例子 2：数组越界

```go
// 没有中间件
func (l *Logic) Process(req *Request) error {
    arr := []int{1, 2, 3}
    value := arr[10]  // 💥 panic: index out of range
    // 程序崩溃！
}

// 有中间件
func (l *Logic) Process(req *Request) error {
    arr := []int{1, 2, 3}
    value := arr[10]  // 💥 panic
    // ✅ 中间件捕获，记录日志，返回 500
    // ✅ 服务继续运行
}
```

### 例子 3：类型断言失败

```go
// 没有中间件
func (l *Logic) Handle(data interface{}) error {
    str := data.(string)  // 💥 panic: interface conversion
    // 程序崩溃！
}

// 有中间件
func (l *Logic) Handle(data interface{}) error {
    str := data.(string)  // 💥 panic
    // ✅ 中间件捕获，记录日志，返回 500
    // ✅ 服务继续运行
}
```

## 中间件做了什么？

```go
func (m *ErrorRecoveryMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            // 1. 捕获 panic
            if err := recover(); err != nil {

                // 2. 记录详细日志（包括堆栈信息）
                logger.LogPanic(ctx, err, extra)

                // 3. 返回友好的错误响应
                w.WriteHeader(500)
                w.Write([]byte("服务器内部错误"))

                // 4. 程序继续运行，不会崩溃
            }
        }()

        // 执行业务逻辑
        next(w, r)
    }
}
```

## 真实场景

### 场景：网盘系统运行中

```
时间轴：
09:00 - 服务启动，一切正常
10:00 - 用户 A 上传文件，成功
10:30 - 用户 B 上传文件，成功
11:00 - 用户 C 上传一个特殊文件，触发代码 bug
        💥 没有中间件：服务崩溃，所有用户无法访问
        ✅ 有中间件：只有用户 C 看到错误，其他用户正常
11:01 - 用户 D 上传文件，成功（服务仍在运行）
11:05 - 开发者查看日志，发现 bug，修复代码
```

## 什么时候会触发 panic？

### 常见的 panic 场景

1. **空指针引用**
   ```go
   var user *User
   name := user.Name  // panic
   ```

2. **数组/切片越界**
   ```go
   arr := []int{1, 2, 3}
   value := arr[10]  // panic
   ```

3. **类型断言失败**
   ```go
   var data interface{} = 123
   str := data.(string)  // panic
   ```

4. **向已关闭的 channel 发送数据**
   ```go
   ch := make(chan int)
   close(ch)
   ch <- 1  // panic
   ```

5. **并发访问 map**
   ```go
   m := make(map[string]int)
   go func() { m["a"] = 1 }()
   go func() { m["b"] = 2 }()  // panic: concurrent map writes
   ```

## 为什么不在每个函数里写 recover？

### 方案 1：每个函数都写（❌ 不推荐）

```go
func Upload() {
    defer func() {
        if r := recover(); r != nil {
            log.Println(r)
        }
    }()
    // 业务逻辑
}

func Download() {
    defer func() {
        if r := recover(); r != nil {
            log.Println(r)
        }
    }()
    // 业务逻辑
}

// 每个函数都要写，太麻烦！
```

### 方案 2：用中间件（✅ 推荐）

```go
// 只写一次
middleware.ErrorRecovery()

// 所有请求都被保护
func Upload() { /* 业务逻辑 */ }
func Download() { /* 业务逻辑 */ }
func Delete() { /* 业务逻辑 */ }
```

## 总结

### 错误恢复中间件的作用

1. **防止程序崩溃** - 最重要的作用
2. **记录详细日志** - 包括堆栈信息，方便排查
3. **返回友好错误** - 而不是让用户看到程序崩溃
4. **提高可用性** - 单个请求失败不影响其他请求

### 类比

就像给房子装**保险丝**：
- 没有保险丝：某个房间短路 → 整栋楼停电
- 有保险丝：某个房间短路 → 只有这个房间断电，其他房间正常

### 最佳实践

1. **全局注册中间件** - 保护所有接口
2. **记录详细日志** - 方便排查问题
3. **不要滥用 panic** - 业务错误用 error，不要用 panic
4. **定期查看日志** - 发现 panic 要及时修复代码

### 记住

- panic 是**代码 bug**，不是正常的业务错误
- 中间件是**最后一道防线**，防止程序崩溃
- 有 panic 说明代码有问题，要**修复代码**，而不是依赖中间件
