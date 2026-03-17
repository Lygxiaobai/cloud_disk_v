#!/bin/bash

echo "=== 日志系统集成验证 ==="
echo ""

# 检查日志目录
if [ -d "./logs" ]; then
    echo "✅ 日志目录存在"
else
    echo "❌ 日志目录不存在"
fi

# 检查关键文件
echo ""
echo "检查关键文件："

files=(
    "core/core.go"
    "core/internal/logger/simple_logger.go"
    "core/internal/middleware/error-recovery-middleware.go"
    "core/internal/svc/service-context.go"
    "core/internal/handler/routes.go"
    "core/internal/logic/user-login-logic.go"
    "core/internal/logic/user-register-logic.go"
    "core/internal/logic/file-upload-logic.go"
    "core/internal/logic/user-file-delete-logic.go"
    "core/internal/logic/share-basic-create-logic.go"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "  ✅ $file"
    else
        echo "  ❌ $file"
    fi
done

# 检查是否导入了 logger
echo ""
echo "检查日志导入："

if grep -q "cloud_disk/core/internal/logger" core/internal/logic/user-login-logic.go; then
    echo "  ✅ user-login-logic.go 已导入 logger"
else
    echo "  ❌ user-login-logic.go 未导入 logger"
fi

if grep -q "cloud_disk/core/internal/logger" core/internal/logic/user-register-logic.go; then
    echo "  ✅ user-register-logic.go 已导入 logger"
else
    echo "  ❌ user-register-logic.go 未导入 logger"
fi

# 检查中间件注册
echo ""
echo "检查中间件注册："

if grep -q "ErrorRecovery" core/internal/svc/service-context.go; then
    echo "  ✅ ServiceContext 中已添加 ErrorRecovery"
else
    echo "  ❌ ServiceContext 中未添加 ErrorRecovery"
fi

if grep -q "serverCtx.ErrorRecovery" core/internal/handler/routes.go; then
    echo "  ✅ routes.go 中已注册 ErrorRecovery 中间件"
else
    echo "  ❌ routes.go 中未注册 ErrorRecovery 中间件"
fi

# 检查日志初始化
echo ""
echo "检查日志初始化："

if grep -q "logger.InitSimpleLogger" core/core.go; then
    echo "  ✅ main 函数中已初始化日志系统"
else
    echo "  ❌ main 函数中未初始化日志系统"
fi

echo ""
echo "=== 验证完成 ==="
echo ""
echo "下一步："
echo "1. 启动服务: cd core && go run core.go"
echo "2. 触发一些错误（如错误的登录）"
echo "3. 查看日志: cat ./logs/error.log | jq ."
