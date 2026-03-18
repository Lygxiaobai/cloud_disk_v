#!/bin/bash

echo "=========================================="
echo "批量优化 Logic 层错误处理"
echo "=========================================="
echo ""

cd D:/Go_Project/my_cloud_disk/core/internal/logic

# 需要修改的文件列表
files=(
    "file-upload-logic.go"
    "file-upload-multipart-logic.go"
    "mail-code-send-register-logic.go"
    "refresh-token-logic.go"
    "share-basic-create-logic.go"
    "share-file-detail-logic.go"
    "share-file-save-logic.go"
    "user-detail-logic.go"
    "user-file-delete-logic.go"
    "user-file-list-logic.go"
    "user-file-move-logic.go"
    "user-file-name-update-logic.go"
    "user-folder-children-logic.go"
    "user-folder-create-logic.go"
    "user-folder-path-logic.go"
    "user-register-logic.go"
    "user-repository-save-logic.go"
)

echo "需要修改的文件数: ${#files[@]}"
echo ""

for file in "${files[@]}"; do
    echo "处理: $file"

    # 1. 替换 import
    sed -i 's|"cloud_disk/core/internal/logger"|"cloud_disk/core/internal/errors"|g' "$file"

    # 2. 删除标准库 errors 的 import（如果存在）
    sed -i '/^[[:space:]]*"errors"$/d' "$file"

    echo "   ✓ 完成"
done

echo ""
echo "=========================================="
echo "批量修改完成！"
echo "=========================================="
echo ""
echo "下一步："
echo "1. 手动检查每个文件，替换 logger.LogError 为 errors.New"
echo "2. 删除手动构建 context 的代码"
echo "3. 编译测试: go build -o bin/core.exe core.go"
