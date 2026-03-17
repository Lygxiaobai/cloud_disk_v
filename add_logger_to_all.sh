#!/bin/bash

# 为所有 Logic 文件添加日志记录的辅助脚本

echo "=== 需要添加日志的 Logic 文件 ==="
echo ""

files=(
    "file-upload-multipart-logic.go"
    "mail-code-send-register-logic.go"
    "refresh-token-logic.go"
    "share-file-detail-logic.go"
    "share-file-save-logic.go"
    "user-detail-logic.go"
    "user-file-list-logic.go"
    "user-file-move-logic.go"
    "user-file-name-update-logic.go"
    "user-folder-children-logic.go"
    "user-folder-create-logic.go"
    "user-folder-path-logic.go"
    "user-repository-save-logic.go"
)

for file in "${files[@]}"; do
    echo "- $file"
done

echo ""
echo "总计: ${#files[@]} 个文件需要添加日志"
echo ""
echo "添加步骤："
echo "1. 导入 logger 包"
echo "2. 在方法开始处获取 TraceID"
echo "3. 在所有 return err 之前添加 logger.LogError()"
