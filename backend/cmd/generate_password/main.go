package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// 生成密码哈希工具
// 用法: go run cmd/generate_password/main.go <password>
func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run cmd/generate_password/main.go <password>")
		fmt.Println("示例: go run cmd/generate_password/main.go admin123456")
		os.Exit(1)
	}

	password := os.Args[1]

	// 生成bcrypt哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("❌ 生成密码哈希失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 密码哈希生成成功\n")
	fmt.Printf("原始密码: %s\n", password)
	fmt.Printf("bcrypt哈希: %s\n", string(hash))
	fmt.Printf("\n📝 使用此哈希值更新数据库:\n")
	fmt.Printf("UPDATE vp_users SET password_hash = '%s' WHERE email = 'admin@voicepaper.com';\n", string(hash))
}

