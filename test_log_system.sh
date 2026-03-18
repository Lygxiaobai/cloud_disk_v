#!/bin/bash

echo "=========================================="
echo "测试日志系统（完整流程）"
echo "=========================================="
echo ""

cd D:/Go_Project/my_cloud_disk/core

# 1. 检查 RabbitMQ
echo "1. 检查 RabbitMQ 状态..."
if ! curl -s http://localhost:15672 > /dev/null 2>&1; then
    echo "   ✗ RabbitMQ 未运行"
    echo "   请先启动: docker start rabbitmq"
    exit 1
fi
echo "   ✓ RabbitMQ 运行中"
echo ""

# 2. 启动本地日志消费者（后台）
echo "2. 启动本地日志消费者..."
./bin/log_worker_local.exe -f etc/core-api.yaml > logs/local_worker.log 2>&1 &
LOCAL_PID=$!
echo "   ✓ 本地日志消费者已启动 (PID: $LOCAL_PID)"
sleep 2
echo ""

# 3. 测试错误处理（触发日志）
echo "3. 测试错误日志记录..."
echo "   启动主服务并触发一个错误..."

# 启动主服务（后台）
./bin/core.exe -f etc/core-api.yaml > logs/core.log 2>&1 &
CORE_PID=$!
echo "   ✓ 主服务已启动 (PID: $CORE_PID)"
sleep 3

# 触发一个登录失败（会记录错误日志）
echo "   触发登录失败（用户名或密码错误）..."
curl -s -X POST http://localhost:8888/user/login \
  -H "Content-Type: application/json" \
  -d '{"name":"test_error","password":"wrong_password"}' > /dev/null 2>&1

sleep 2
echo ""

# 4. 检查日志文件
echo "4. 检查日志文件..."
if [ -f "logs/error.log" ]; then
    echo "   ✓ 日志文件存在"

    # 查看最新的日志
    echo ""
    echo "   最新日志内容："
    echo "   ----------------------------------------"
    tail -1 logs/error.log | jq '.' 2>/dev/null || tail -1 logs/error.log
    echo "   ----------------------------------------"
else
    echo "   ✗ 日志文件不存在"
fi
echo ""

# 5. 检查 RabbitMQ 队列
echo "5. 检查 RabbitMQ 队列状态..."
QUEUE_MESSAGES=$(curl -s -u guest:guest http://localhost:15672/api/queues/%2F/local_log_queue | jq -r '.messages' 2>/dev/null)
if [ "$QUEUE_MESSAGES" != "null" ]; then
    echo "   ✓ 队列消息数: $QUEUE_MESSAGES"
else
    echo "   ⚠️  无法获取队列状态"
fi
echo ""

# 6. 清理
echo "6. 清理测试进程..."
kill $CORE_PID 2>/dev/null
kill $LOCAL_PID 2>/dev/null
sleep 1
echo "   ✓ 测试进程已停止"
echo ""

echo "=========================================="
echo "测试完成！"
echo "=========================================="
echo ""
echo "查看完整日志："
echo "  - 错误日志: tail -f logs/error.log | jq '.'"
echo "  - 主服务日志: tail -f logs/core.log"
echo "  - 消费者日志: tail -f logs/local_worker.log"
