package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"voicepaper/config"
	"voicepaper/internal/storage"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	// 2. 创建OSS存储实例
	ossStorage, err := storage.NewOSSStorage(cfg)
	if err != nil {
		log.Fatalf("❌ 创建OSS存储失败: %v", err)
	}
	fmt.Println("✅ OSS存储实例创建成功")

	ctx := context.Background()

	// 3. 读取本地文件
	// 尝试多个可能的路径
	possiblePaths := []string{
		"../../../data/output.mp3", // 从 backend/cmd/test_oss/ 到 data/
		"../../data/output.mp3",    // 从 backend/ 到 data/
		"../data/output.mp3",       // 从项目根目录
		"data/output.mp3",          // 当前目录
	}

	var filePath string
	var fileData []byte
	var readErr error

	for _, path := range possiblePaths {
		fileData, readErr = os.ReadFile(path)
		if readErr == nil {
			filePath = path
			break
		}
	}

	if readErr != nil {
		log.Fatalf("❌ 读取文件失败，尝试的路径: %v, 错误: %v", possiblePaths, readErr)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Fatalf("❌ 获取文件信息失败: %v", err)
	}

	fmt.Printf("📁 本地文件信息:\n")
	fmt.Printf("  路径: %s\n", filePath)
	fmt.Printf("  大小: %d 字节 (%.2f MB)\n", fileInfo.Size(), float64(fileInfo.Size())/(1024*1024))

	// 4. 上传到OSS的 paper/ 目录
	ossPath := "paper/output.mp3"
	fmt.Printf("\n📤 开始上传到OSS: %s\n", ossPath)

	url, err := ossStorage.Save(ctx, ossPath, fileData)
	if err != nil {
		log.Fatalf("❌ 上传文件失败: %v", err)
	}

	fmt.Printf("✅ 文件上传成功!\n")
	fmt.Printf("  访问URL: %s\n", url)

	// 5. 验证文件是否存在
	fmt.Printf("\n🔍 验证文件是否存在...\n")
	exists, err := ossStorage.Exists(ctx, ossPath)
	if err != nil {
		log.Fatalf("❌ 检查文件是否存在失败: %v", err)
	}
	if exists {
		fmt.Printf("✅ 文件已存在于OSS\n")
	} else {
		fmt.Printf("❌ 文件不存在于OSS\n")
		return
	}

	// 6. 获取文件详细信息
	fmt.Printf("\n📋 获取文件详细信息...\n")
	ossFileInfo, err := ossStorage.GetFileInfo(ctx, ossPath)
	if err != nil {
		log.Fatalf("❌ 获取文件信息失败: %v", err)
	}

	fmt.Printf("✅ 文件信息获取成功:\n")
	fmt.Printf("  路径: %s\n", ossFileInfo.Path)
	fmt.Printf("  大小: %d 字节 (%.2f MB)\n", ossFileInfo.Size, float64(ossFileInfo.Size)/(1024*1024))
	fmt.Printf("  Content-Type: %s\n", ossFileInfo.ContentType)
	fmt.Printf("  ETag: %s\n", ossFileInfo.ETag)
	fmt.Printf("  最后修改时间: %s\n", ossFileInfo.LastModified.Format("2006-01-02 15:04:05"))
	fmt.Printf("  访问URL: %s\n", ossFileInfo.URL)

	// 7. 测试获取文件大小
	fmt.Printf("\n📏 测试获取文件大小...\n")
	size, err := ossStorage.GetSize(ctx, ossPath)
	if err != nil {
		log.Fatalf("❌ 获取文件大小失败: %v", err)
	}
	fmt.Printf("✅ 文件大小: %d 字节\n", size)

	// 8. 测试获取访问URL
	fmt.Printf("\n🔗 测试获取访问URL...\n")
	accessURL := ossStorage.GetURL(ossPath)
	fmt.Printf("✅ 普通URL: %s\n", accessURL)
	fmt.Printf("   ⚠️  如果Bucket是私有的，此URL可能无法访问\n")

	// 8.1 测试获取签名URL（用于私有Bucket）
	fmt.Printf("\n🔐 测试获取签名URL（有效期1小时）...\n")
	signedURL, err := ossStorage.GetSignedURL(ctx, ossPath, 3600)
	if err != nil {
		fmt.Printf("⚠️  生成签名URL失败: %v\n", err)
	} else {
		fmt.Printf("✅ 签名URL: %s\n", signedURL)
		fmt.Printf("   ✅ 此URL在1小时内有效，可用于访问私有Bucket中的文件\n")
	}

	// 9. 总结
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("🎉 测试完成!\n\n")
	fmt.Printf("📝 总结:\n")
	fmt.Printf("  1. 文件已成功上传到OSS\n")
	fmt.Printf("  2. 文件路径: %s\n", ossPath)
	fmt.Printf("  3. 访问URL: %s\n", accessURL)
	fmt.Printf("\n💡 访问方式:\n")
	fmt.Printf("  浏览器直接访问: %s\n", accessURL)
	fmt.Printf("  或在代码中使用: storage.GetURL(\"%s\")\n", ossPath)
	fmt.Printf(strings.Repeat("=", 60) + "\n")
}
