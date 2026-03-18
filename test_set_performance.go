package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	defer rdb.Close()

	ctx := context.Background()

	fmt.Println("==========================================")
	fmt.Println("测试 SET 优化后的 IsHotShare 性能")
	fmt.Println("==========================================")
	fmt.Println()

	// 测试热门分享
	hotIdentity := "5ddf2b53-085f-43b5-a82d-7154c43ee6de"
	// 测试非热门分享
	coldIdentity := "non-existent-identity"

	iterations := 10000

	// ========== 测试热门分享判断 ==========
	fmt.Println("【测试1：判断热门分享】")
	start := time.Now()

	for i := 0; i < iterations; i++ {
		rdb.SIsMember(ctx, "share:hot:set", hotIdentity).Result()
	}

	hotDuration := time.Since(start)
	hotAvg := hotDuration / time.Duration(iterations)

	fmt.Printf("执行次数: %d\n", iterations)
	fmt.Printf("总耗时: %v\n", hotDuration)
	fmt.Printf("平均耗时: %v\n", hotAvg)
	fmt.Println()

	// ========== 测试非热门分享判断 ==========
	fmt.Println("【测试2：判断非热门分享】")
	start = time.Now()

	for i := 0; i < iterations; i++ {
		rdb.SIsMember(ctx, "share:hot:set", coldIdentity).Result()
	}

	coldDuration := time.Since(start)
	coldAvg := coldDuration / time.Duration(iterations)

	fmt.Printf("执行次数: %d\n", iterations)
	fmt.Printf("总耗时: %v\n", coldDuration)
	fmt.Printf("平均耗时: %v\n", coldAvg)
	fmt.Println()

	// ========== 验证功能 ==========
	fmt.Println("【功能验证】")
	isHot, _ := rdb.SIsMember(ctx, "share:hot:set", hotIdentity).Result()
	isCold, _ := rdb.SIsMember(ctx, "share:hot:set", coldIdentity).Result()

	fmt.Printf("热门分享判断: %v ✓\n", isHot)
	fmt.Printf("非热门分享判断: %v ✓\n", !isCold)
	fmt.Println()

	// ========== 查看热榜集合 ==========
	fmt.Println("【热榜集合内容】")
	members, _ := rdb.SMembers(ctx, "share:hot:set").Result()
	fmt.Printf("热榜数量: %d\n", len(members))
	fmt.Println("热榜成员:")
	for i, member := range members {
		fmt.Printf("  %d. %s\n", i+1, member)
	}
	fmt.Println()

	fmt.Println("==========================================")
	fmt.Println("测试完成")
	fmt.Println("==========================================")
}
