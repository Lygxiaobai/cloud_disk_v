# 日志系统集成进度

## 已完成的 Logic 文件（8/18）

### ✅ 已添加日志记录
1. ✅ file-upload-logic.go - 文件上传
2. ✅ file-upload-multipart-logic.go - 分片上传
3. ✅ share-basic-create-logic.go - 创建分享
4. ✅ user-file-delete-logic.go - 删除文件
5. ✅ user-file-move-logic.go - 移动文件
6. ✅ user-folder-create-logic.go - 创建文件夹
7. ✅ user-login-logic.go - 用户登录
8. ✅ user-register-logic.go - 用户注册

## 待添加的 Logic 文件（10/18）

### ⏳ 需要添加日志记录
1. ⏳ mail-code-send-register-logic.go - 发送验证码
2. ⏳ refresh-token-logic.go - 刷新Token
3. ⏳ share-file-detail-logic.go - 分享文件详情
4. ⏳ share-file-save-logic.go - 保存分享文件
5. ⏳ user-detail-logic.go - 用户详情
6. ⏳ user-file-list-logic.go - 文件列表
7. ⏳ user-file-name-update-logic.go - 重命名文件
8. ⏳ user-folder-children-logic.go - 文件夹子项
9. ⏳ user-folder-path-logic.go - 文件夹路径
10. ⏳ user-repository-save-logic.go - 保存文件到用户空间

## 添加模板

每个 Logic 文件需要添加以下内容：

### 1. 导入 logger 包
```go
import (
    "cloud_disk/core/internal/logger"
    // ... 其他导入
)
```

### 2. 在方法开始处获取 TraceID
```go
func (l *YourLogic) YourMethod(req *types.Request) error {
    // 从 context 中获取 TraceID
    traceID, _ := l.ctx.Value("trace_id").(string)
    ctx := context.WithValue(l.ctx, "method", "POST")
    ctx = context.WithValue(ctx, "path", "/your/path")
    ctx = context.WithValue(ctx, "trace_id", traceID)

    // 业务逻辑...
}
```

### 3. 在所有错误返回前添加日志
```go
if err != nil {
    logger.LogError(ctx, "操作失败", err, map[string]interface{}{
        "param1": req.Param1,
        "param2": req.Param2,
    })
    return nil, err
}
```

## 快速添加命令

```bash
# 检查哪些文件还没有添加日志
for file in core/internal/logic/*-logic.go; do
    if ! grep -q "logger.LogError" "$file"; then
        echo "⏳ $(basename $file)"
    else
        echo "✅ $(basename $file)"
    fi
done
```

## 优先级

### 高优先级（用户直接操作）
1. ✅ user-login-logic.go
2. ✅ user-register-logic.go
3. ✅ file-upload-logic.go
4. ✅ file-upload-multipart-logic.go
5. ✅ user-file-delete-logic.go
6. ✅ user-file-move-logic.go
7. ✅ user-folder-create-logic.go
8. ⏳ user-file-name-update-logic.go

### 中优先级（常用功能）
1. ⏳ user-file-list-logic.go
2. ⏳ user-folder-children-logic.go
3. ⏳ share-basic-create-logic.go
4. ⏳ share-file-save-logic.go

### 低优先级（辅助功能）
1. ⏳ user-detail-logic.go
2. ⏳ refresh-token-logic.go
3. ⏳ mail-code-send-register-logic.go
4. ⏳ share-file-detail-logic.go
5. ⏳ user-folder-path-logic.go
6. ⏳ user-repository-save-logic.go

## 下一步

你可以选择：

### 选项 1：我继续添加剩余的 10 个文件
- 优点：完整覆盖所有 Logic
- 缺点：需要一些时间

### 选项 2：你自己按照模板添加
- 优点：你可以根据业务重要性选择
- 缺点：需要手动操作

### 选项 3：先使用当前版本
- 优点：核心功能已覆盖
- 缺点：部分功能没有日志

## 建议

**建议选择选项 1**，让我继续完成剩余的 10 个文件，这样可以确保：
- ✅ 所有接口都有完整的错误日志
- ✅ 任何错误都能追踪
- ✅ 系统完全可观测

是否继续添加剩余的 10 个文件？
