package storage

import (
	"context"
	"fmt"
	"log"
	"voicepaper/config"
)

// ExampleOSSFileInfo 示例：如何获取OSS上传文件的信息
// 这个函数展示了如何使用OSS存储获取文件详细信息
func ExampleOSSFileInfo() {
	// 1. 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 检查存储类型
	if cfg.Storage.Type != "oss" {
		log.Println("当前存储类型不是OSS，请将配置中的 storage.type 设置为 'oss'")
		return
	}

	// 3. 创建OSS存储实例
	ossStorage, err := NewOSSStorage(cfg)
	if err != nil {
		log.Fatalf("创建OSS存储失败: %v", err)
	}

	// 4. 获取文件信息
	ctx := context.Background()
	filePath := "audio/example.mp3" // OSS中的文件路径（key）

	// 方法1: 使用 GetFileInfo 获取完整信息
	fileInfo, err := ossStorage.GetFileInfo(ctx, filePath)
	if err != nil {
		log.Fatalf("获取文件信息失败: %v", err)
	}

	// 打印文件信息
	fmt.Printf("文件路径: %s\n", fileInfo.Path)
	fmt.Printf("文件大小: %d 字节 (%.2f MB)\n", fileInfo.Size, float64(fileInfo.Size)/(1024*1024))
	fmt.Printf("Content-Type: %s\n", fileInfo.ContentType)
	fmt.Printf("ETag: %s\n", fileInfo.ETag)
	fmt.Printf("最后修改时间: %s\n", fileInfo.LastModified.Format("2006-01-02 15:04:05"))
	fmt.Printf("访问URL: %s\n", fileInfo.URL)

	// 方法2: 检查文件是否存在
	exists, err := ossStorage.Exists(ctx, filePath)
	if err != nil {
		log.Fatalf("检查文件是否存在失败: %v", err)
	}
	fmt.Printf("文件是否存在: %v\n", exists)

	// 方法3: 获取文件大小
	size, err := ossStorage.GetSize(ctx, filePath)
	if err != nil {
		log.Fatalf("获取文件大小失败: %v", err)
	}
	fmt.Printf("文件大小: %d 字节\n", size)

	// 方法4: 获取访问URL
	url := ossStorage.GetURL(filePath)
	fmt.Printf("访问URL: %s\n", url)
}

// ExampleOSSWithMediaResource 示例：结合MediaResource模型使用OSS
// 这个函数展示了如何将OSS文件信息保存到数据库
func ExampleOSSWithMediaResource() {
	// 1. 加载配置并创建OSS存储
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	ossStorage, err := NewOSSStorage(cfg)
	if err != nil {
		log.Fatalf("创建OSS存储失败: %v", err)
	}

	// 2. 上传文件到OSS（示例）
	ctx := context.Background()
	filePath := "audio/article_123.mp3"
	fileData := []byte("fake audio data") // 实际使用时是真实的音频数据

	url, err := ossStorage.Save(ctx, filePath, fileData)
	if err != nil {
		log.Fatalf("上传文件失败: %v", err)
	}
	fmt.Printf("文件上传成功，URL: %s\n", url)

	// 3. 获取文件信息
	fileInfo, err := ossStorage.GetFileInfo(ctx, filePath)
	if err != nil {
		log.Fatalf("获取文件信息失败: %v", err)
	}

	// 4. 将信息保存到MediaResource（需要结合repository使用）
	// mediaResource := &model.MediaResource{
	//     ArticleID:    123,
	//     ResourceType: model.MediaResourceTypeAudio,
	//     StorageType:  model.StorageTypeOSS,
	//     StoragePath:  filePath,
	//     StorageURL:   fileInfo.URL,
	//     FileName:     "article_123.mp3",
	//     FileSize:     fileInfo.Size,
	//     MimeType:     fileInfo.ContentType,
	//     FileHash:     fileInfo.ETag,
	//     OSSBucket:    cfg.Storage.OSS.Bucket,
	//     OSSRegion:    cfg.Storage.OSS.Region,
	//     Status:       model.MediaResourceStatusCompleted,
	// }
	// repository.DB.Create(mediaResource)

	fmt.Printf("文件信息已获取:\n")
	fmt.Printf("  - 大小: %d 字节\n", fileInfo.Size)
	fmt.Printf("  - 类型: %s\n", fileInfo.ContentType)
	fmt.Printf("  - URL: %s\n", fileInfo.URL)
}

