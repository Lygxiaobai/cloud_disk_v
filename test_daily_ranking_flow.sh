#!/bin/bash

echo "=========================================="
echo "日榜功能测试"
echo "=========================================="
echo ""

# 1. 检查 Redis 连接
echo "1. 检查 Redis 连接..."
docker exec my-redis redis-cli PING
echo ""

# 2. 查看当前日榜数据
echo "2. 查看今日日榜 (share:daily:clicks:2026-03-18)..."
docker exec my-redis redis-cli ZREVRANGE share:daily:clicks:2026-03-18 0 10 WITHSCORES
echo ""

# 3. 查看热榜列表
echo "3. 查看热榜列表 (share:hot:list)..."
COUNT=$(docker exec my-redis redis-cli LLEN share:hot:list)
echo "热榜数量: $COUNT"
if [ "$COUNT" -gt 0 ]; then
    echo "前10个热榜分享:"
    docker exec my-redis redis-cli LRANGE share:hot:list 0 9
fi
echo ""

# 4. 查看数据库点击数
echo "4. 查看数据库点击数前5..."
mysql -u root -p123456 -e "USE cloud_disk; SELECT identity, click_num, updated_at FROM share_basic WHERE deleted_at IS NULL ORDER BY click_num DESC LIMIT 5;" 2>/dev/null
echo ""

# 5. 测试访问分享
echo "5. 测试访问分享..."
IDENTITY="5ddf2b53-085f-43b5-a82d-7154c43ee6de"
echo "访问分享: $IDENTITY"
curl -s "http://localhost:8888/share/file/detail/$IDENTITY" | head -c 100
echo ""
echo ""

# 6. 查看访问后的日榜变化
echo "6. 查看访问后的日榜变化..."
SCORE=$(docker exec my-redis redis-cli ZSCORE share:daily:clicks:2026-03-18 $IDENTITY)
echo "分享 $IDENTITY 的今日点击数: $SCORE"
echo ""

# 7. 查看数据库变化
echo "7. 查看数据库点击数变化..."
mysql -u root -p123456 -e "USE cloud_disk; SELECT identity, click_num FROM share_basic WHERE identity='$IDENTITY';" 2>/dev/null
echo ""

echo "=========================================="
echo "测试完成"
echo "=========================================="
