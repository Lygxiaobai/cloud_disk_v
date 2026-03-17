#!/bin/bash

# 测试热榜功能脚本

echo "=========================================="
echo "开始测试 Redis 缓存热榜功能"
echo "=========================================="
echo ""

# 1. 检查 Redis 连接
echo "1. 检查 Redis 连接..."
redis-cli -h 127.0.0.1 -p 6379 ping
if [ $? -eq 0 ]; then
    echo "✓ Redis 连接正常"
else
    echo "✗ Redis 连接失败"
    exit 1
fi
echo ""

# 2. 查看热门分享列表
echo "2. 查看 Redis 中的热门分享列表..."
HOT_COUNT=$(redis-cli -h 127.0.0.1 -p 6379 LLEN share:hot:list)
echo "热门分享数量: $HOT_COUNT"
if [ "$HOT_COUNT" -gt 0 ]; then
    echo "✓ 热门分享列表已生成"
    echo "前10个热门分享:"
    redis-cli -h 127.0.0.1 -p 6379 LRANGE share:hot:list 0 9
else
    echo "✗ 热门分享列表为空"
fi
echo ""

# 3. 查看热门列表的过期时间
echo "3. 查看热门列表的过期时间..."
TTL=$(redis-cli -h 127.0.0.1 -p 6379 TTL share:hot:list)
if [ "$TTL" -gt 0 ]; then
    echo "✓ 过期时间: ${TTL}秒 (约 $((TTL/60)) 分钟)"
else
    echo "✗ 未设置过期时间或已过期"
fi
echo ""

# 4. 查看所有缓存的分享详情
echo "4. 查看缓存的分享详情..."
DETAIL_COUNT=$(redis-cli -h 127.0.0.1 -p 6379 KEYS "share:detail:*" | wc -l)
echo "已缓存的分享详情数量: $DETAIL_COUNT"
if [ "$DETAIL_COUNT" -gt 0 ]; then
    echo "✓ 已有分享详情缓存"
    echo "示例缓存 key:"
    redis-cli -h 127.0.0.1 -p 6379 KEYS "share:detail:*" | head -3
else
    echo "○ 暂无分享详情缓存（需要访问后才会缓存）"
fi
echo ""

# 5. 查看点击计数
echo "5. 查看点击计数缓存..."
CLICK_COUNT=$(redis-cli -h 127.0.0.1 -p 6379 KEYS "share:click:*" | wc -l)
echo "点击计数缓存数量: $CLICK_COUNT"
if [ "$CLICK_COUNT" -gt 0 ]; then
    echo "✓ 已有点击计数缓存"
    echo "示例点击计数:"
    redis-cli -h 127.0.0.1 -p 6379 KEYS "share:click:*" | head -3 | while read key; do
        count=$(redis-cli -h 127.0.0.1 -p 6379 GET "$key")
        echo "  $key: $count"
    done
else
    echo "○ 暂无点击计数缓存"
fi
echo ""

# 6. 测试 API 访问（如果有热门分享）
if [ "$HOT_COUNT" -gt 0 ]; then
    echo "6. 测试访问热门分享 API..."
    FIRST_SHARE=$(redis-cli -h 127.0.0.1 -p 6379 LINDEX share:hot:list 0)
    if [ -n "$FIRST_SHARE" ]; then
        echo "测试分享 identity: $FIRST_SHARE"
        echo "发送请求..."
        RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" "http://localhost:8888/share/file/detail?identity=$FIRST_SHARE")
        HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
        BODY=$(echo "$RESPONSE" | grep -v "HTTP_CODE")

        if [ "$HTTP_CODE" = "200" ]; then
            echo "✓ API 请求成功 (HTTP $HTTP_CODE)"
            echo "响应内容:"
            echo "$BODY" | python -m json.tool 2>/dev/null || echo "$BODY"
        else
            echo "✗ API 请求失败 (HTTP $HTTP_CODE)"
            echo "$BODY"
        fi
    fi
    echo ""

    # 7. 再次检查缓存是否生成
    echo "7. 检查访问后的缓存状态..."
    DETAIL_KEY="share:detail:$FIRST_SHARE"
    EXISTS=$(redis-cli -h 127.0.0.1 -p 6379 EXISTS "$DETAIL_KEY")
    if [ "$EXISTS" = "1" ]; then
        echo "✓ 分享详情已缓存到 Redis"
        TTL=$(redis-cli -h 127.0.0.1 -p 6379 TTL "$DETAIL_KEY")
        echo "  缓存过期时间: ${TTL}秒 (约 $((TTL/60)) 分钟)"
    else
        echo "✗ 分享详情未缓存"
    fi
fi

echo ""
echo "=========================================="
echo "测试完成"
echo "=========================================="
