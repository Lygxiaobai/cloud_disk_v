package main

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

func main() {
	// 测试 Casbin 配置是否正确
	modelPath := "internal/authorization/model.conf"
	policyPath := "internal/authorization/policy.csv"

	adapter := fileadapter.NewAdapter(policyPath)
	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		fmt.Printf("❌ 初始化 Casbin 失败: %v\n", err)
		return
	}

	fmt.Println("✅ Casbin 初始化成功")
	fmt.Println("\n=== 测试权限校验 ===\n")

	// 测试用例
	testCases := []struct {
		role   string
		path   string
		method string
		expect bool
	}{
		// admin 角色测试
		{"admin", "/user/file/list", "GET", true},
		{"admin", "/user/file/delete", "DELETE", true},
		{"admin", "/file/upload", "POST", true},

		// user 角色测试
		{"user", "/user/file/list", "GET", true},
		{"user", "/user/file/delete", "DELETE", true},
		{"user", "/file/upload", "POST", true},

		// readonly 角色测试
		{"readonly", "/user/file/list", "GET", true},
		{"readonly", "/user/file/delete", "DELETE", false},
		{"readonly", "/file/upload", "POST", false},

		// 路径参数测试
		{"user", "/user/folder/children/123", "GET", true},
		{"user", "/user/folder/path/abc-def", "GET", true},
		{"readonly", "/user/folder/children/456", "GET", true},

		// 未授权测试
		{"guest", "/user/file/list", "GET", false},
	}

	passCount := 0
	failCount := 0

	for i, tc := range testCases {
		ok, err := enforcer.Enforce(tc.role, tc.path, tc.method)
		if err != nil {
			fmt.Printf("❌ 测试 %d 出错: %v\n", i+1, err)
			failCount++
			continue
		}

		if ok == tc.expect {
			fmt.Printf("✅ 测试 %d: %s %s %s -> %v (预期: %v)\n",
				i+1, tc.role, tc.method, tc.path, ok, tc.expect)
			passCount++
		} else {
			fmt.Printf("❌ 测试 %d: %s %s %s -> %v (预期: %v)\n",
				i+1, tc.role, tc.method, tc.path, ok, tc.expect)
			failCount++
		}
	}

	fmt.Printf("\n=== 测试结果 ===\n")
	fmt.Printf("通过: %d\n", passCount)
	fmt.Printf("失败: %d\n", failCount)
	fmt.Printf("总计: %d\n", passCount+failCount)

	if failCount == 0 {
		fmt.Println("\n🎉 所有测试通过！Casbin 配置正确！")
	} else {
		fmt.Println("\n⚠️  部分测试失败，请检查配置")
	}
}
