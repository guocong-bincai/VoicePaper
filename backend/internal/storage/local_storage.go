package storage

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"voicepaper/config"
)

// LocalStorage 本地文件存储实现
type LocalStorage struct {
	baseDir string
	baseURL string // 用于生成访问 URL，如 "http://localhost:8080/audio"
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(cfg *config.Config) *LocalStorage {
	return &LocalStorage{
		baseDir: cfg.Storage.OutputDir,
		baseURL: fmt.Sprintf("http://localhost%s", cfg.Service.Port), // 可以从配置中读取
	}
}

// Save 保存文件到本地
func (s *LocalStorage) Save(ctx context.Context, path string, data []byte) (string, error) {
	fullPath := filepath.Join(s.baseDir, path)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	// 返回访问 URL
	return s.GetURL(path), nil
}

// SaveFromReader 从 Reader 保存文件
func (s *LocalStorage) SaveFromReader(ctx context.Context, path string, reader io.Reader, size int64) (string, error) {
	fullPath := filepath.Join(s.baseDir, path)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 复制数据
	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return s.GetURL(path), nil
}

// SaveFromReaderWithPublicRead 从 Reader 保存文件并设置为公开可读
// 本地存储不需要特殊处理，直接调用 SaveFromReader
func (s *LocalStorage) SaveFromReaderWithPublicRead(ctx context.Context, path string, reader io.Reader, size int64) (string, error) {
	return s.SaveFromReader(ctx, path, reader, size)
}

// Get 读取文件内容
func (s *LocalStorage) Get(ctx context.Context, path string) ([]byte, error) {
	fullPath := filepath.Join(s.baseDir, path)
	return os.ReadFile(fullPath)
}

// GetURL 获取文件的访问 URL
func (s *LocalStorage) GetURL(path string) string {
	// 根据路径类型生成不同的 URL
	// audio/xxx.mp3 -> /audio/xxx.mp3
	// timeline/xxx.json -> /api/v1/timeline/xxx.json
	if filepath.Dir(path) == "audio" {
		return fmt.Sprintf("%s/audio/%s", s.baseURL, filepath.Base(path))
	}
	return fmt.Sprintf("%s/api/v1/%s", s.baseURL, path)
}

// GetSignedURL 获取文件的签名URL（本地存储直接返回普通URL）
func (s *LocalStorage) GetSignedURL(ctx context.Context, path string, expires int64) (string, error) {
	return s.GetURL(path), nil
}

// GetSignedImageURL 获取带图片处理参数的签名URL（本地存储忽略处理参数）
func (s *LocalStorage) GetSignedImageURL(ctx context.Context, path string, expires int64, width int) (string, error) {
	// 本地存储不支持动态图片处理，直接返回原图URL
	return s.GetURL(path), nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.baseDir, path)
	return os.Remove(fullPath)
}

// Exists 检查文件是否存在
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(s.baseDir, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GetSize 获取文件大小
func (s *LocalStorage) GetSize(ctx context.Context, path string) (int64, error) {
	info, err := s.GetFileInfo(ctx, path)
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

// GetFileInfo 获取文件详细信息
func (s *LocalStorage) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	fullPath := filepath.Join(s.baseDir, path)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 读取文件内容计算ETag（MD5）
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 计算MD5作为ETag
	etag := fmt.Sprintf("%x", md5.Sum(data))

	// 根据文件扩展名推断Content-Type
	ext := strings.ToLower(filepath.Ext(path))
	var contentType string
	switch ext {
	case ".mp3":
		contentType = "audio/mpeg"
	case ".json":
		contentType = "application/json"
	case ".txt":
		contentType = "text/plain"
	case ".md":
		contentType = "text/markdown"
	default:
		contentType = "application/octet-stream"
	}

	return &FileInfo{
		Path:         path,
		Size:         info.Size(),
		ContentType:  contentType,
		ETag:         etag,
		LastModified: info.ModTime(),
		URL:          s.GetURL(path),
	}, nil
}
