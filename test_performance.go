package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	fmt.Println("==========================================")
	fmt.Println("Redis 缓存性能对比测试")
	fmt.Println("==========================================")
	fmt.Println()

	// 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	ctx := context.Background()

	// 获取热门分享列表
	hotList, err := rdb.LRange(ctx, "share:hot:list", 0, 4).Result()
	if err != nil || len(hotList) == 0 {
		fmt.Println("✗ 无法获取热门分享列表")
		return
	}

	fmt.Printf("测试分享数量: %d\n", len(hotList))
	fmt.Println()

	// 测试每个分享
	for i, identity := range hotList {
		fmt.Printf("测试 %d: %s\n", i+1, identity)

		// 清除缓存
		detailKey := "share:detail:" + identity
		rdb.Del(ctx, detailKey)

		// 第一次请求（无缓存）
		url := fmt.Sprintf("http://localhost:8888/share/file/detail/%s", identity)
		start := time.Now()
		resp, err := http.Get(url)
		duration1 := time.Since(start)

		if err != nil {
			fmt.Printf("  ✗ 请求失败: %v\n", err)
			continue
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Printf("  ✗ HTTP %d\n", resp.StatusCode)
			continue
		}

		fmt.Printf("  第一次请求（查数据库）: %v\n", duration1)

		// 等待缓存写入
		time.Sleep(50 * time.Millisecond)

		// 第二次请求（有缓存）
		start = time.Now()
		resp, err = http.Get(url)
		duration2 := time.Since(start)

		if err != nil {
			fmt.Printf("  ✗ 请求失败: %v\n", err)
			continue
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()

		fmt.Printf("  第二次请求（读 Redis）: %v\n", duration2)

		// 计算性能提升
		speedup := float64(duration1) / float64(duration2)
		fmt.Printf("  性能提升: %.1fx\n", speedup)
		fmt.Println()
	}

	fmt.Println("==========================================")
	fmt.Println("测试完成")
	fmt.Println("==========================================")
}
