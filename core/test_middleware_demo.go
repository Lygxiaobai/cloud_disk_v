package main

import (
	"cloud_disk/core/internal/logger"
	"cloud_disk/core/internal/middleware"
	"fmt"
	"log"
	"net/http"
)

// 模拟一个有 bug 的 handler
func buggyHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("开始处理请求...")

	// 模拟代码 bug：访问 nil 指针
	var user *struct{ Name string }

	// 💥 这里会 panic
	name := user.Name

	w.Write([]byte("Hello " + name))
}

// 正常的 handler
func normalHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("正常响应"))
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
	http.HandleFunc("/buggy", errorRecovery.Handle(buggyHandler))
	http.HandleFunc("/normal", errorRecovery.Handle(normalHandler))

	fmt.Println("=== 错误恢复中间件演示 ===")
	fmt.Println("服务启动在 http://localhost:8080")
	fmt.Println("")
	fmt.Println("测试步骤：")
	fmt.Println("1. 访问 http://localhost:8080/normal  （正常请求）")
	fmt.Println("2. 访问 http://localhost:8080/buggy   （会触发 panic）")
	fmt.Println("3. 再次访问 http://localhost:8080/normal （服务仍然正常）")
	fmt.Println("4. 查看日志: cat ./logs/error.log | tail -1 | jq .")
	fmt.Println("")
	fmt.Println("观察：即使 /buggy 触发 panic，服务也不会崩溃！")
	fmt.Println("")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
