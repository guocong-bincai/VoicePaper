package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	// 测试登录流程
	// 从环境变量或命令行参数获取
	email := os.Getenv("TEST_EMAIL")
	if email == "" {
		fmt.Println("❌ 请设置 TEST_EMAIL 环境变量或提供邮箱参数")
		fmt.Println("用法: TEST_EMAIL=xxx TEST_CODE=xxx go run main.go")
		os.Exit(1)
	}

	code := os.Getenv("TEST_CODE")
	if code == "" {
		fmt.Println("❌ 请设置 TEST_CODE 环境变量")
		os.Exit(1)
	}

	fmt.Printf("🔐 测试登录流程...\n")
	fmt.Printf("邮箱: %s\n", email)
	fmt.Printf("验证码: %s\n", code)
	fmt.Println()

	// 发送登录请求
	url := "http://localhost:8080/api/v1/auth/email/login"
	data := map[string]string{
		"email": email,
		"code":  code,
	}

	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		if token, ok := result["token"].(string); ok {
			fmt.Printf("\n✅ 登录成功！\n")
			fmt.Printf("Token: %s\n", token[:50]+"...")
		}
	} else {
		fmt.Printf("\n❌ 登录失败\n")
	}
}

