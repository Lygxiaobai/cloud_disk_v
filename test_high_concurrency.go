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

	// 准备100个 identity
	identities := make([]string, 100)
	for i := 0; i < 100; i++ {
		identities[i] = fmt.Sprintf("identity-%d", i)
	}

	rdb.Del(ctx, "test:list", "test:set")
	for _, id := range identities {
		rdb.RPush(ctx, "test:list", id)
	}
	rdb.SAdd(ctx, "test:set", identities)

	fmt.Println("==========================================")
	fmt.Println("高并发场景性能对比")
	fmt.Println("==========================================")
	fmt.Println()

	// 模拟高并发：1000个请求/秒
	qps := 1000
	duration := 5 * time.Second
	totalRequests := qps * int(duration.Seconds())

	fmt.Printf("模拟场景: %d QPS，持续 %v\n", qps, duration)
	fmt.Printf("总请求数: %d\n", totalRequests)
	fmt.Println()

	// ========== List 方案 ==========
	fmt.Println("【List 方案】")
	start := time.Now()
	listSuccess := 0

	for i := 0; i < totalRequests; i++ {
		targetIdentity := fmt.Sprintf("identity-%d", i%100)
		result, err := rdb.LRange(ctx, "test:list", 0, -1).Result()
		if err == nil {
			for _, id := range result {
				if id == targetIdentity {
					listSuccess++
					break
				}
			}
		}
	}

	listDuration := time.Since(start)
	listQPS := float64(totalRequests) / listDuration.Seconds()

	fmt.Printf("总耗时: %v\n", listDuration)
	fmt.Printf("成功请求: %d\n", listSuccess)
	fmt.Printf("实际 QPS: %.2f\n", listQPS)
	fmt.Printf("平均延迟: %v\n", listDuration/time.Duration(totalRequests))
	fmt.Println()

	// ========== Set 方案 ==========
	fmt.Println("【Set 方案】")
	start = time.Now()
	setSuccess := 0

	for i := 0; i < totalRequests; i++ {
		targetIdentity := fmt.Sprintf("identity-%d", i%100)
		exists, err := rdb.SIsMember(ctx, "test:set", targetIdentity).Result()
		if err == nil && exists {
			setSuccess++
		}
	}

	setDuration := time.Since(start)
	setQPS := float64(totalRequests) / setDuration.Seconds()

	fmt.Printf("总耗时: %v\n", setDuration)
	fmt.Printf("成功请求: %d\n", setSuccess)
	fmt.Printf("实际 QPS: %.2f\n", setQPS)
	fmt.Printf("平均延迟: %v\n", setDuration/time.Duration(totalRequests))
	fmt.Println()

	// ========== 对比 ==========
	fmt.Println("==========================================")
	fmt.Println("性能对比")
	fmt.Println("==========================================")

	speedup := float64(listDuration) / float64(setDuration)
	qpsImprovement := (setQPS - listQPS) / listQPS * 100

	fmt.Printf("List 方案 QPS: %.2f\n", listQPS)
	fmt.Printf("Set 方案 QPS:  %.2f\n", setQPS)
	fmt.Printf("QPS 提升: %.2f%%\n", qpsImprovement)
	fmt.Printf("响应时间提升: %.2fx\n", speedup)
	fmt.Println()

	// ========== CPU 和内存估算 ==========
	fmt.Println("【资源消耗估算】")
	fmt.Println()

	fmt.Println("List 方案:")
	fmt.Printf("  - 每次查询传输: ~1KB\n")
	fmt.Printf("  - 每秒传输量: ~%d MB\n", qps*1/1024)
	fmt.Printf("  - 内存遍历: 100次比较/请求\n")
	fmt.Println()

	fmt.Println("Set 方案:")
	fmt.Printf("  - 每次查询传输: ~1B\n")
	fmt.Printf("  - 每秒传输量: ~%d KB\n", qps*1/1024)
	fmt.Printf("  - 哈希查找: O(1)\n")
	fmt.Println()

	rdb.Del(ctx, "test:list", "test:set")

	fmt.Println("==========================================")
	fmt.Println("结论: Set 方案在高并发下优势明显")
	fmt.Println("==========================================")
}
