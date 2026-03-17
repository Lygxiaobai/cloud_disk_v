package main

import (
	"cloud_disk/core/internal/logger"
	"context"
	"errors"
	"log"
)

func main() {
	log.Println("=== V4 版本测试：完整版日志系统 ===")

	// 1. 初始化
	if err := logger.InitSimpleLogger("./logs/error.log"); err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer logger.Close()

	// 2. 测试 ERROR 级别
	log.Println("\n--- 测试 ERROR 级别 ---")
	ctx1 := context.Background()
	ctx1 = context.WithValue(ctx1, "trace_id", "trace-001")
	ctx1 = context.WithValue(ctx1, "user_id", "user-123")
	ctx1 = context.WithValue(ctx1, "method", "POST")
	ctx1 = context.WithValue(ctx1, "path", "/file/upload")

	logger.LogError(ctx1, "文件上传失败", errors.New("connection timeout"), map[string]interface{}{
		"file_name": "test.pdf",
		"file_size": 1024000,
		"retry":     3,
	})

	// 3. 测试 FATAL 级别
	log.Println("\n--- 测试 FATAL 级别 ---")
	ctx2 := context.Background()
	ctx2 = context.WithValue(ctx2, "trace_id", "trace-002")
	ctx2 = context.WithValue(ctx2, "user_id", "user-456")
	ctx2 = context.WithValue(ctx2, "method", "GET")
	ctx2 = context.WithValue(ctx2, "path", "/database/connect")

	logger.LogFatal(ctx2, "数据库连接失败", errors.New("connection refused"), map[string]interface{}{
		"db_host": "localhost",
		"db_port": 3306,
	})

	// 4. 测试 PANIC 级别
	log.Println("\n--- 测试 PANIC 级别 ---")
	ctx3 := context.Background()
	ctx3 = context.WithValue(ctx3, "trace_id", "trace-003")
	ctx3 = context.WithValue(ctx3, "user_id", "user-789")
	ctx3 = context.WithValue(ctx3, "method", "DELETE")
	ctx3 = context.WithValue(ctx3, "path", "/file/delete")

	logger.LogPanic(ctx3, "nil pointer dereference", map[string]interface{}{
		"file_id": "12345",
		"action":  "delete",
	})

	log.Println("\n✅ 测试完成！查看日志文件:")
	log.Println("   cat ./logs/error.log")
	log.Println("\n💡 提示：现在日志包含完整信息：")
	log.Println("   - 时间戳")
	log.Println("   - 日志级别（ERROR/FATAL/PANIC）")
	log.Println("   - TraceID 和 UserID")
	log.Println("   - 堆栈信息")
	log.Println("   - 额外字段")
}
