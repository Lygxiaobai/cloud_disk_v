#!/bin/bash

echo "=========================================="
echo "启动 Elasticsearch (Docker)"
echo "=========================================="
echo ""

# 检查是否已有 elasticsearch 容器
if docker ps -a | grep -q elasticsearch; then
    echo "发现已存在的 Elasticsearch 容器"

    # 检查是否正在运行
    if docker ps | grep -q elasticsearch; then
        echo "✓ Elasticsearch 已经在运行"
        echo ""
        echo "访问地址: http://localhost:9200"
        exit 0
    else
        echo "正在启动 Elasticsearch..."
        docker start elasticsearch
        echo "✓ Elasticsearch 已启动"
        echo ""
        echo "访问地址: http://localhost:9200"
        exit 0
    fi
fi

# 创建新的 Elasticsearch 容器
echo "创建新的 Elasticsearch 容器..."
echo ""

docker run -d \
  --name elasticsearch \
  -p 9200:9200 \
  -p 9300:9300 \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
  elasticsearch:8.11.0

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Elasticsearch 容器创建成功"
    echo ""
    echo "等待 Elasticsearch 启动..."
    sleep 10

    # 测试连接
    if curl -s http://localhost:9200 > /dev/null 2>&1; then
        echo "✓ Elasticsearch 启动成功"
        echo ""
        echo "访问地址: http://localhost:9200"
        echo ""
        echo "查看集群信息:"
        curl -s http://localhost:9200 | jq '.'
    else
        echo "⚠️  Elasticsearch 可能还在启动中，请稍后再试"
        echo "   运行 'docker logs elasticsearch' 查看日志"
    fi
else
    echo "✗ Elasticsearch 容器创建失败"
    exit 1
fi
