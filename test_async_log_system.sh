#!/bin/bash

echo "=========================================="
echo "测试异步日志系统"
echo "=========================================="
echo ""

# 测试接口（触发日志记录）
echo "测试 1: 触发错误日志（邮箱格式错误）"
response=$(curl -s -X POST http://localhost:8888/mail/code/send/register \
    -H "Content-Type: application/json" \
    -d '{"email":"invalid_email"}')
echo "响应: $response"
echo ""

sleep 2

# 检查 RabbitMQ 队列状态
echo "=========================================="
echo "检查 RabbitMQ 队列状态"
echo "=========================================="
echo ""

echo "日志交换机状态:"
curl -s -u guest:guest http://localhost:15672/api/exchanges/%2F/log_exchange | jq '{name, type, durable, message_stats}'
echo ""

echo "本地日志队列状态:"
curl -s -u guest:guest http://localhost:15672/api/queues/%2F/local_log_queue | jq '{name, consumers, messages, message_stats}'
echo ""

echo "ES 日志队列状态:"
curl -s -u guest:guest http://localhost:15672/api/queues/%2F/es_log_queue | jq '{name, consumers, messages, message_stats}'
echo ""

# 检查本地日志文件
echo "=========================================="
echo "检查本地日志文件"
echo "=========================================="
echo ""
echo "最新的 3 条日志:"
tail -3 D:/Go_Project/my_cloud_disk/core/logs/error.log | jq '.'
echo ""

echo "=========================================="
echo "测试完成"
echo "=========================================="
