package main

import (
	"fmt"
	"log"
	"voicepaper/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func main() {
	// 从配置文件加载 OSS 配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	endpoint := cfg.OSS.Endpoint
	accessKeyID := cfg.OSS.AccessKeyID
	accessKeySecret := cfg.OSS.AccessKeySecret
	bucketName := cfg.OSS.BucketName

	// 测试文件路径
	testPath := "articles/original_en/20260130_164000.md"

	// 创建 OSS 客户端
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		log.Fatalf("创建 OSS 客户端失败: %v", err)
	}

	// 获取 Bucket
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		log.Fatalf("获取 Bucket 失败: %v", err)
	}

	// 1. 检查 Bucket ACL
	fmt.Println("=== 检查 Bucket ACL ===")
	bucketACL, err := client.GetBucketACL(bucketName)
	if err != nil {
		log.Printf("❌ 获取 Bucket ACL 失败: %v", err)
	} else {
		fmt.Printf("✅ Bucket ACL: %s\n", bucketACL.ACL)
	}

	// 2. 检查文件是否存在
	fmt.Println("\n=== 检查文件是否存在 ===")
	exists, err := bucket.IsObjectExist(testPath)
	if err != nil {
		log.Printf("❌ 检查文件失败: %v", err)
	} else if exists {
		fmt.Printf("✅ 文件存在: %s\n", testPath)
	} else {
		fmt.Printf("❌ 文件不存在: %s\n", testPath)
		return
	}

	// 3. 检查 Object ACL
	fmt.Println("\n=== 检查 Object ACL ===")
	objectACL, err := bucket.GetObjectACL(testPath)
	if err != nil {
		log.Printf("❌ 获取 Object ACL 失败: %v", err)
	} else {
		fmt.Printf("✅ Object ACL: %s\n", objectACL.ACL)
	}

	// 4. 生成签名 URL
	fmt.Println("\n=== 生成签名 URL ===")
	signedURL, err := bucket.SignURL(testPath, oss.HTTPGet, 3600)
	if err != nil {
		log.Printf("❌ 生成签名 URL 失败: %v", err)
	} else {
		fmt.Printf("✅ 签名 URL: %s\n", signedURL)
	}

	// 5. 尝试直接下载（测试权限）
	fmt.Println("\n=== 测试下载权限 ===")
	_, err = bucket.GetObject(testPath)
	if err != nil {
		log.Printf("❌ 下载失败: %v", err)
		fmt.Println("\n🔴 权限问题！AccessKey 可能没有读取权限")
	} else {
		fmt.Println("✅ 下载成功！权限正常")
	}

	// 6. 建议修复方案
	fmt.Println("\n=== 修复建议 ===")
	if err == nil {
		if bucketACL.ACL == "private" {
			fmt.Println("💡 Bucket 是私有的，需要签名访问")
			fmt.Println("💡 请检查 RAM 用户是否有 oss:GetObject 权限")
			fmt.Println("💡 或者将 Bucket ACL 改为 public-read")
		} else if bucketACL.ACL == "public-read" {
			fmt.Println("✅ Bucket 是公开读的，可以直接访问")
		}
	}
}
