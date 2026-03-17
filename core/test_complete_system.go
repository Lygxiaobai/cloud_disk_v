package main

import (
	"cloud_disk/core/internal/logger"
	"cloud_disk/core/internal/middleware"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

// 模拟 Logic 层
func simulateUserLogin(ctx context.Context, username string) error {
	// 从 context 中获取 TraceID
	traceID, _ := ctx.Value("trace_id").(string)

	// 构建日志上下文
	logCtx := context.WithValue(ctx, "method", "POST")
	logCtx = context.WithValue(logCtx, "path", "/user/login")
	logCtx = context.WithValue(logCtx, "trace_id", traceID)

	// 模拟数据库查询失败
	err := errors.New("database connection timeout")
	logger.LogError(logCtx, "数据库查询失败", err, map[string]interface{}{
		"username": username,
	})

	// 模拟登录失败
	err = errors.New("用户名或密码错误")
	logger.LogError(logCtx, "用户登录失败", err, map[string]interface{}{
		"username": username,
		"reason":   "密码错误",
	})

	return err
}

// 模拟会 panic 的 handler
func panicHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("处理请求...")

	// 模拟一些正常操作
	ctx := r.Context()
	simulateUserLogin(ctx, "test@example.com")

	// 模拟 panic
	var user *struct{ Name string }
	_ = user.Name // 💥 nil pointer panic
}

// 正常的 handler
func normalHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceID, _ := ctx.Value("trace_id").(string)

	log.Printf("处理正常请求，TraceID: %s", traceID)

	// 模拟业务逻辑
	simulateUserLogin(ctx, "normal@example.com")

	w.Write([]byte(fmt.Sprintf("请求成功，TraceID: %s", traceID)))
}

func main() {
	// 初始化日志系统
	if err := logger.InitSimpleLogger("./logs/error.log"); err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}
	defer logger.Close()

	// 创建错误恢复中间件
	errorRecovery := middleware.NewErrorRecoveryMiddleware()

	// 注册路由
	http.HandleFunc("/panic", errorRecovery.Handle(panicHandler))
	http.HandleFunc("/normal", errorRecovery.Handle(normalHandler))

	fmt.Println("=== 日志系统完整测试 ===")
	fmt.Println("")
	fmt.Println("服务启动在 http://localhost:8080")
	fmt.Println("")
	fmt.Println("测试步骤：")
	fmt.Println("1. 打开新终端，运行测试命令")
	fmt.Println("2. 查看日志文件")
	fmt.Println("")
	fmt.Println("测试命令：")
	fmt.Println("  curl http://localhost:8080/normal")
	fmt.Println("  curl http://localhost:8080/panic")
	fmt.Println("")
	fmt.Println("查看日志：")
	fmt.Println("  tail -f ./logs/error.log")
	fmt.Println("")

	// 启动一个 goroutine 自动测试
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("\n>>> 自动测试开始 <<<\n")

		// 测试正常请求
		fmt.Println("1. 测试正常请求...")
		resp, _ := http.Get("http://localhost:8080/normal")
		if resp != nil {
			resp.Body.Close()
			fmt.Println("   ✅ 正常请求完成")
		}

		time.Sleep(1 * time.Second)

		// 测试 panic 请求
		fmt.Println("2. 测试 panic 请求...")
		resp, _ = http.Get("http://localhost:8080/panic")
		if resp != nil {
			resp.Body.Close()
			fmt.Println("   ✅ panic 被捕获，服务继续运行")
		}

		time.Sleep(1 * time.Second)

		// 再次测试正常请求，验证服务没有崩溃
		fmt.Println("3. 再次测试正常请求（验证服务未崩溃）...")
		resp, _ = http.Get("http://localhost:8080/normal")
		if resp != nil {
			resp.Body.Close()
			fmt.Println("   ✅ 服务正常运行")
		}

		fmt.Println("\n>>> 自动测试完成 <<<")
		fmt.Println("\n查看日志：")
		fmt.Println("  tail -10 ./logs/error.log | python -m json.tool")
		fmt.Println("\n按 Ctrl+C 停止服务")
	}()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
