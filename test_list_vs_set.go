package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	defer rdb.Close()

	ctx := context.Background()

	// 准备测试数据：100个 identity
	identities := make([]string, 100)
	for i := 0; i < 100; i++ {
		identities[i] = fmt.Sprintf("identity-%d", i)
	}

	// 清空旧数据
	rdb.Del(ctx, "test:list", "test:set")

	// 写入 List
	for _, id := range identities {
		rdb.RPush(ctx, "test:list", id)
	}

	// 写入 Set
	rdb.SAdd(ctx, "test:set", identities)

	fmt.Println("==========================================")
	fmt.Println("List vs Set 性能对比测试")
	fmt.Println("==========================================")
	fmt.Println()

	// 测试目标：查找第50个元素（中间位置）
	targetIdentity := "identity-50"

	// ========== 测试 List ==========
	fmt.Println("【方案1：List + LRANGE + 遍历】")

	start := time.Now()
	iterations := 10000

	for i := 0; i < iterations; i++ {
		// 模拟当前实现
		result, _ := rdb.LRange(ctx, "test:list", 0, -1).Result()
		found := false
		for _, id := range result {
			if id == targetIdentity {
				found = true
				break
			}
		}
		_ = found
	}

	listDuration := time.Since(start)
	listAvg := listDuration / time.Duration(iterations)

	fmt.Printf("执行次数: %d\n", iterations)
	fmt.Printf("总耗时: %v\n", listDuration)
	fmt.Printf("平均耗时: %v\n", listAvg)
	fmt.Println()

	// ========== 测试 Set ==========
	fmt.Println("【方案2：Set + SISMEMBER】")

	start = time.Now()

	for i := 0; i < iterations; i++ {
		// 优化后的实现
		rdb.SIsMember(ctx, "test:set", targetIdentity).Result()
	}

	setDuration := time.Since(start)
	setAvg := setDuration / time.Duration(iterations)

	fmt.Printf("执行次数: %d\n", iterations)
	fmt.Printf("总耗时: %v\n", setDuration)
	fmt.Printf("平均耗时: %v\n", setAvg)
	fmt.Println()

	// ========== 性能对比 ==========
	fmt.Println("==========================================")
	fmt.Println("性能对比结果")
	fmt.Println("==========================================")

	speedup := float64(listDuration) / float64(setDuration)

	fmt.Printf("List 方案: %v\n", listAvg)
	fmt.Printf("Set 方案:  %v\n", setAvg)
	fmt.Printf("性能提升: %.2fx\n", speedup)
	fmt.Println()

	// ========== 网络传输对比 ==========
	fmt.Println("【网络传输量对比】")

	// List: 传输100个字符串
	listBytes := 0
	for _, id := range identities {
		listBytes += len(id)
	}

	// Set: 只传输1个布尔值
	setBytes := 1

	fmt.Printf("List 方案: 约 %d 字节 (100个字符串)\n", listBytes)
	fmt.Printf("Set 方案:  约 %d 字节 (1个布尔值)\n", setBytes)
	fmt.Printf("传输量减少: %.2fx\n", float64(listBytes)/float64(setBytes))
	fmt.Println()

	// 清理
	rdb.Del(ctx, "test:list", "test:set")

	fmt.Println("==========================================")
	fmt.Println("测试完成")
	fmt.Println("==========================================")
}
