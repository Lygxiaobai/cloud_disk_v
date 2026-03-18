#!/bin/bash

echo "=========================================="
echo "测试 Elasticsearch 日志功能"
echo "=========================================="
echo ""

# 1. 检查 ES 是否运行
echo "1. 检查 Elasticsearch 状态..."
if ! curl -s http://localhost:9200 > /dev/null 2>&1; then
    echo "   ✗ Elasticsearch 未运行"
    echo "   请先运行: bash start_elasticsearch.sh"
    exit 1
fi
echo "   ✓ Elasticsearch 运行中"
echo ""

# 2. 检查索引模板
echo "2. 检查索引模板..."
TEMPLATE_EXISTS=$(curl -s http://localhost:9200/_index_template/logs-template | jq -r '.index_templates | length')
if [ "$TEMPLATE_EXISTS" -gt 0 ]; then
    echo "   ✓ 索引模板已存在"
else
    echo "   ⚠️  索引模板不存在（首次启动 ES worker 时会自动创建）"
fi
echo ""

# 3. 查看现有索引
echo "3. 查看日志索引..."
curl -s "http://localhost:9200/_cat/indices/logs-*?v&h=index,docs.count,store.size"
echo ""

# 4. 查询最新日志
echo "4. 查询最新 5 条日志..."
cd D:/Go_Project/my_cloud_disk/core
if [ -f "./bin/query_es_logs.exe" ]; then
    ./bin/query_es_logs.exe -n 5
else
    echo "   ⚠️  查询工具未编译，请先运行:"
    echo "   go build -o bin/query_es_logs.exe cmd/query_es_logs/main.go"
fi
echo ""

# 5. 统计日志数量
echo "5. 统计日志数量..."
TOTAL=$(curl -s "http://localhost:9200/logs-*/_count" | jq -r '.count')
echo "   总日志数: $TOTAL"
echo ""

echo "=========================================="
echo "测试完成！"
echo "=========================================="
