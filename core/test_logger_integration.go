package main

import (
	"cloud_disk/core/internal/logger"
	"context"
	"errors"
	"log"
)

func main() {
	log.Println("=== 测试日志系统集成 ===")

	// 1. 初始化日志系统
	if err := logger.InitSimpleLogger("./logs/error.log"); err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer logger.Close()

	// 2. 模拟文件上传错误
	log.Println("\n--- 测试场景 1: 文件上传失败 ---")
	ctx1 := context.Background()
	ctx1 = context.WithValue(ctx1, "method", "POST")
	ctx1 = context.WithValue(ctx1, "path", "/file/upload")
	ctx1 = context.WithValue(ctx1, "user_id", "user-123")

	logger.LogError(ctx1, "文件上传失败", errors.New("database insert failed"), map[string]interface{}{
		"file_name": "document.pdf",
		"file_size": 2048000,
		"file_hash": "abc123def456",
	})

	// 3. 模拟用户登录错误
	log.Println("\n--- 测试场景 2: 用户登录失败 ---")
	ctx2 := context.Background()
	ctx2 = context.WithValue(ctx2, "method", "POST")
	ctx2 = context.WithValue(ctx2, "path", "/user/login")
	ctx2 = context.WithValue(ctx2, "user_id", "unknown")

	logger.LogError(ctx2, "用户登录失败", errors.New("invalid password"), map[string]interface{}{
		"username": "test@example.com",
		"ip":       "192.168.1.100",
	})

	// 4. 模拟数据库连接失败（Fatal 级别）
	log.Println("\n--- 测试场景 3: 数据库连接失败 ---")
	ctx3 := context.Background()
	ctx3 = context.WithValue(ctx3, "method", "SYSTEM")
	ctx3 = context.WithValue(ctx3, "path", "/init")

	logger.LogFatal(ctx3, "数据库连接失败", errors.New("connection refused"), map[string]interface{}{
		"db_host": "localhost",
		"db_port": 3306,
		"db_name": "cloud_disk",
	})

	// 5. 模拟 panic（Panic 级别）
	log.Println("\n--- 测试场景 4: 程序 Panic ---")
	ctx4 := context.Background()
	ctx4 = context.WithValue(ctx4, "method", "DELETE")
	ctx4 = context.WithValue(ctx4, "path", "/file/delete")
	ctx4 = context.WithValue(ctx4, "user_id", "user-456")

	logger.LogPanic(ctx4, "nil pointer dereference", map[string]interface{}{
		"file_id": "file-789",
		"action":  "delete",
	})

	log.Println("\n✅ 测试完成！")
	log.Println("📁 查看日志文件: cat ./logs/error.log")
	log.Println("📊 格式化查看: cat ./logs/error.log | jq .")
}
