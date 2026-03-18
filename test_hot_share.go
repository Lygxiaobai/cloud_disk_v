package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	fmt.Println("==========================================")
	fmt.Println("开始测试 Redis 缓存热榜功能")
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

	// 1. 检查 Redis 连接
	fmt.Println("1. 检查 Redis 连接...")
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("✗ Redis 连接失败: %v\n", err)
		return
	}
	fmt.Printf("✓ Redis 连接正常: %s\n\n", pong)

	// 2. 查看热门分享集合
	fmt.Println("2. 查看 Redis 中的热门分享集合...")
	hotCount, err := rdb.SCard(ctx, "share:hot:set").Result()
	if err != nil {
		fmt.Printf("✗ 获取热门集合失败: %v\n", err)
	} else {
		fmt.Printf("热门分享数量: %d\n", hotCount)
		if hotCount > 0 {
			fmt.Println("✓ 热门分享集合已生成")
			hotList, _ := rdb.SMembers(ctx, "share:hot:set").Result()
			fmt.Println("前10个热门分享:")
			for i, identity := range hotList {
				if i >= 10 {
					break
				}
				fmt.Printf("  %d. %s\n", i+1, identity)
			}
		} else {
			fmt.Println("✗ 热门分享集合为空")
		}
	}
	fmt.Println()

	// 3. 查看热门集合的过期时间
	fmt.Println("3. 查看热门集合的过期时间...")
	ttl, err := rdb.TTL(ctx, "share:hot:set").Result()
	if err != nil {
		fmt.Printf("✗ 获取过期时间失败: %v\n", err)
	} else if ttl > 0 {
		fmt.Printf("✓ 过期时间: %v (约 %.1f 分钟)\n", ttl, ttl.Minutes())
	} else {
		fmt.Println("✗ 未设置过期时间或已过期")
	}
	fmt.Println()

	// 4. 查看所有缓存的分享详情
	fmt.Println("4. 查看缓存的分享详情...")
	detailKeys, err := rdb.Keys(ctx, "share:detail:*").Result()
	if err != nil {
		fmt.Printf("✗ 获取详情缓存失败: %v\n", err)
	} else {
		fmt.Printf("已缓存的分享详情数量: %d\n", len(detailKeys))
		if len(detailKeys) > 0 {
			fmt.Println("✓ 已有分享详情缓存")
			fmt.Println("示例缓存 key:")
			for i, key := range detailKeys {
				if i >= 3 {
					break
				}
				fmt.Printf("  %s\n", key)
			}
		} else {
			fmt.Println("○ 暂无分享详情缓存（需要访问后才会缓存）")
		}
	}
	fmt.Println()

	// 5. 查看点击计数
	fmt.Println("5. 查看点击计数缓存...")
	clickKeys, err := rdb.Keys(ctx, "share:click:*").Result()
	if err != nil {
		fmt.Printf("✗ 获取点击计数失败: %v\n", err)
	} else {
		fmt.Printf("点击计数缓存数量: %d\n", len(clickKeys))
		if len(clickKeys) > 0 {
			fmt.Println("✓ 已有点击计数缓存")
			fmt.Println("示例点击计数:")
			for i, key := range clickKeys {
				if i >= 3 {
					break
				}
				count, _ := rdb.Get(ctx, key).Int64()
				fmt.Printf("  %s: %d\n", key, count)
			}
		} else {
			fmt.Println("○ 暂无点击计数缓存")
		}
	}
	fmt.Println()

	// 6. 测试 API 访问（如果有热门分享）
	if hotCount > 0 {
		fmt.Println("6. 测试访问热门分享 API...")
		hotList, _ := rdb.SMembers(ctx, "share:hot:set").Result()
		if len(hotList) > 0 {
			firstShare := hotList[0]
			fmt.Printf("测试分享 identity: %s\n", firstShare)
			fmt.Println("发送请求...")

			url := fmt.Sprintf("http://localhost:8888/share/file/detail/%s", firstShare)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("✗ API 请求失败: %v\n", err)
			} else {
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)

				if resp.StatusCode == 200 {
					fmt.Printf("✓ API 请求成功 (HTTP %d)\n", resp.StatusCode)
					fmt.Println("响应内容:")
					var prettyJSON map[string]interface{}
					if json.Unmarshal(body, &prettyJSON) == nil {
						formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
						fmt.Println(string(formatted))
					} else {
						fmt.Println(string(body))
					}
				} else {
					fmt.Printf("✗ API 请求失败 (HTTP %d)\n", resp.StatusCode)
					fmt.Println(string(body))
				}
			}
			fmt.Println()

			// 7. 再次检查缓存是否生成
			fmt.Println("7. 检查访问后的缓存状态...")
			time.Sleep(100 * time.Millisecond) // 等待异步缓存写入
			detailKey := "share:detail:" + firstShare
			exists, err := rdb.Exists(ctx, detailKey).Result()
			if err != nil {
				fmt.Printf("✗ 检查缓存失败: %v\n", err)
			} else if exists == 1 {
				fmt.Println("✓ 分享详情已缓存到 Redis")
				ttl, _ := rdb.TTL(ctx, detailKey).Result()
				fmt.Printf("  缓存过期时间: %v (约 %.1f 分钟)\n", ttl, ttl.Minutes())
			} else {
				fmt.Println("✗ 分享详情未缓存")
			}
		}
	}

	fmt.Println()
	fmt.Println("==========================================")
	fmt.Println("测试完成")
	fmt.Println("==========================================")
}
