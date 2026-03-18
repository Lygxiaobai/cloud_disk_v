#!/bin/bash

echo "=========================================="
echo "邮件发送性能测试"
echo "=========================================="
echo ""

echo "测试 5 次，记录响应时间..."
echo ""

for i in {1..5}; do
    email="test_$(date +%s%N)@example.com"
    echo "测试 $i: $email"

    start=$(date +%s%N)
    response=$(curl -s -X POST http://localhost:8888/mail/code/send/register \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\"}")
    end=$(date +%s%N)

    duration=$(( (end - start) / 1000000 ))
    echo "  响应时间: ${duration}ms"
    echo "  响应内容: $response"
    echo ""

    sleep 1
done

echo "=========================================="
echo "测试完成"
echo "=========================================="
