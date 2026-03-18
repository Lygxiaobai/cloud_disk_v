#!/bin/bash

echo "=========================================="
echo "启动异步日志系统（RabbitMQ fanout 模式）"
echo "=========================================="
echo ""

# 检查 RabbitMQ 是否运行
echo "1. 检查 RabbitMQ 状态..."
if ! docker ps | grep -q rabbitmq; then
    echo "   ⚠️  RabbitMQ 未运行，正在启动..."
    docker start rabbitmq
    sleep 3
fi
echo "   ✓ RabbitMQ 运行中"
echo ""

# 检查 Elasticsearch 是否运行（可选）
echo "2. 检查 Elasticsearch 状态..."
if curl -s http://localhost:9200 > /dev/null 2>&1; then
    echo "   ✓ Elasticsearch 运行中"
else
    echo "   ⚠️  Elasticsearch 未运行（ES 日志消费者将无法工作）"
    echo "   提示: 运行 'docker start elasticsearch' 启动 ES"
fi
echo ""

# 启动本地日志消费者
echo "3. 启动本地日志消费者..."
cd D:/Go_Project/my_cloud_disk/core
./bin/log_worker_local.exe -f etc/core-api.yaml &
LOCAL_PID=$!
echo "   ✓ 本地日志消费者已启动 (PID: $LOCAL_PID)"
echo ""

# 启动 ES 日志消费者
echo "4. 启动 ES 日志消费者..."
./bin/log_worker_es.exe -f etc/core-api.yaml &
ES_PID=$!
echo "   ✓ ES 日志消费者已启动 (PID: $ES_PID)"
echo ""

# 等待消费者初始化
sleep 2

# 启动主服务
echo "5. 启动主服务..."
./bin/core.exe -f etc/core-api.yaml &
CORE_PID=$!
echo "   ✓ 主服务已启动 (PID: $CORE_PID)"
echo ""

echo "=========================================="
echo "所有服务启动完成！"
echo "=========================================="
echo ""
echo "进程信息："
echo "  - 主服务: PID $CORE_PID"
echo "  - 本地日志消费者: PID $LOCAL_PID"
echo "  - ES 日志消费者: PID $ES_PID"
echo ""
echo "RabbitMQ 管理界面: http://localhost:15672"
echo "用户名: guest, 密码: guest"
echo ""
echo "查询 ES 日志: ./bin/query_es_logs.exe -n 10"
echo "过滤级别: ./bin/query_es_logs.exe -level ERROR"
echo "按 TraceID 查询: ./bin/query_es_logs.exe -trace <trace_id>"
echo ""
echo "按 Ctrl+C 停止所有服务"
echo ""

# 等待用户中断
wait
