package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"voicepaper/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSStorage 阿里云OSS存储实现
type OSSStorage struct {
	client     *oss.Client
	bucket     *oss.Bucket
	bucketName string
	baseURL    string // 自定义域名或默认域名
	useHTTPS   bool
}

// NewOSSStorage 创建OSS存储实例
func NewOSSStorage(cfg *config.Config) (*OSSStorage, error) {
	ossCfg := cfg.Storage.OSS

	// 验证配置
	if ossCfg.Endpoint == "" {
		return nil, fmt.Errorf("OSS Endpoint 不能为空")
	}
	if ossCfg.AccessKeyID == "" {
		return nil, fmt.Errorf("OSS AccessKeyID 不能为空")
	}
	if ossCfg.AccessKeySecret == "" {
		return nil, fmt.Errorf("OSS AccessKeySecret 不能为空")
	}
	if ossCfg.Bucket == "" {
		return nil, fmt.Errorf("OSS Bucket 不能为空")
	}

	// 创建OSS客户端
	client, err := oss.New(ossCfg.Endpoint, ossCfg.AccessKeyID, ossCfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建OSS客户端失败: %w", err)
	}

	// 获取Bucket
	bucket, err := client.Bucket(ossCfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("获取OSS Bucket失败: %w", err)
	}

	// 生成基础URL
	var baseURL string
	if ossCfg.BaseURL != "" {
		// 使用自定义域名
		baseURL = strings.TrimSuffix(ossCfg.BaseURL, "/")
	} else {
		// 使用默认域名
		protocol := "https"
		if !ossCfg.UseHTTPS {
			protocol = "http"
		}
		baseURL = fmt.Sprintf("%s://%s.%s", protocol, ossCfg.Bucket, ossCfg.Endpoint)
	}

	return &OSSStorage{
		client:     client,
		bucket:     bucket,
		bucketName: ossCfg.Bucket,
		baseURL:    baseURL,
		useHTTPS:   ossCfg.UseHTTPS,
	}, nil
}

// Save 保存文件到OSS
func (s *OSSStorage) Save(ctx context.Context, path string, data []byte) (string, error) {
	// 确保路径不以/开头
	key := strings.TrimPrefix(path, "/")

	err := s.bucket.PutObject(key, strings.NewReader(string(data)))
	if err != nil {
		return "", fmt.Errorf("上传文件到OSS失败: %w", err)
	}

	return s.GetURL(path), nil
}

// SaveFromReader 从 Reader 保存文件到OSS
func (s *OSSStorage) SaveFromReader(ctx context.Context, path string, reader io.Reader, size int64) (string, error) {
	// 确保路径不以/开头
	key := strings.TrimPrefix(path, "/")

	err := s.bucket.PutObject(key, reader)
	if err != nil {
		return "", fmt.Errorf("上传文件到OSS失败: %w", err)
	}

	return s.GetURL(path), nil
}

// SaveFromReaderWithPublicRead 从 Reader 保存文件到OSS并设置为公开可读
// 适用于头像等需要公开访问的文件
func (s *OSSStorage) SaveFromReaderWithPublicRead(ctx context.Context, path string, reader io.Reader, size int64) (string, error) {
	// 确保路径不以/开头
	key := strings.TrimPrefix(path, "/")

	// 上传时设置 ACL 为公开可读
	err := s.bucket.PutObject(key, reader, oss.ObjectACL(oss.ACLPublicRead))
	if err != nil {
		return "", fmt.Errorf("上传文件到OSS失败: %w", err)
	}

	return s.GetURL(path), nil
}

// Get 从OSS获取文件内容
func (s *OSSStorage) Get(ctx context.Context, path string) ([]byte, error) {
	key := strings.TrimPrefix(path, "/")

	body, err := s.bucket.GetObject(key)
	if err != nil {
		return nil, fmt.Errorf("从OSS获取文件失败: %w", err)
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}

	return data, nil
}

// GetURL 获取文件的访问URL
func (s *OSSStorage) GetURL(path string) string {
	// 确保路径不以/开头
	key := strings.TrimPrefix(path, "/")
	return fmt.Sprintf("%s/%s", s.baseURL, key)
}

// GetSignedURL 获取文件的签名URL（用于私有Bucket访问）
// expires: URL有效期（秒），例如 3600 表示1小时
func (s *OSSStorage) GetSignedURL(ctx context.Context, path string, expires int64) (string, error) {
	key := strings.TrimPrefix(path, "/")

	// 生成签名URL
	signedURL, err := s.bucket.SignURL(key, oss.HTTPGet, expires)
	if err != nil {
		return "", fmt.Errorf("生成签名URL失败: %w", err)
	}

	// 强制使用 HTTPS（替换 http:// 为 https://）
	if s.useHTTPS && strings.HasPrefix(signedURL, "http://") {
		signedURL = strings.Replace(signedURL, "http://", "https://", 1)
	}

	return signedURL, nil
}

// GetSignedImageURL 获取带图片处理参数的签名URL
func (s *OSSStorage) GetSignedImageURL(ctx context.Context, path string, expires int64, width int) (string, error) {
	key := strings.TrimPrefix(path, "/")

	// 构建终极加速处理参数:
	// 1. resize: 缩放到指定宽度
	// 2. format,webp: 强制转换为 WebP 格式 (体积缩小 70%)
	// 3. quality,q_85: 质量压缩到 85% (兼顾画质与体积)
	process := "image/format,webp/quality,q_85"
	if width > 0 {
		process = fmt.Sprintf("image/resize,w_%d/format,webp/quality,q_85", width)
	}

	// 构建OSS图片处理选项
	var options []oss.Option
	options = append(options, oss.Process(process))

	// 生成带处理参数的签名URL
	signedURL, err := s.bucket.SignURL(key, oss.HTTPGet, expires, options...)
	if err != nil {
		return "", fmt.Errorf("生成图片签名URL失败: %w", err)
	}

	// 强制使用 HTTPS（替换 http:// 为 https://）
	if s.useHTTPS && strings.HasPrefix(signedURL, "http://") {
		signedURL = strings.Replace(signedURL, "http://", "https://", 1)
	}

	return signedURL, nil
}

// Delete 从OSS删除文件
func (s *OSSStorage) Delete(ctx context.Context, path string) error {
	key := strings.TrimPrefix(path, "/")

	err := s.bucket.DeleteObject(key)
	if err != nil {
		return fmt.Errorf("从OSS删除文件失败: %w", err)
	}

	return nil
}

// Exists 检查文件是否存在
func (s *OSSStorage) Exists(ctx context.Context, path string) (bool, error) {
	key := strings.TrimPrefix(path, "/")

	exists, err := s.bucket.IsObjectExist(key)
	if err != nil {
		return false, fmt.Errorf("检查OSS文件是否存在失败: %w", err)
	}

	return exists, nil
}

// GetSize 获取文件大小
func (s *OSSStorage) GetSize(ctx context.Context, path string) (int64, error) {
	info, err := s.GetFileInfo(ctx, path)
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

// GetFileInfo 获取文件详细信息
// 这是获取OSS文件信息的主要方法
func (s *OSSStorage) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	key := strings.TrimPrefix(path, "/")

	// 获取对象元信息
	props, err := s.bucket.GetObjectMeta(key)
	if err != nil {
		return nil, fmt.Errorf("获取OSS文件信息失败: %w", err)
	}

	// 解析文件大小
	var size int64
	if contentLength := props.Get("Content-Length"); contentLength != "" {
		fmt.Sscanf(contentLength, "%d", &size)
	}

	// 解析最后修改时间
	var lastModified time.Time
	if lastModifiedStr := props.Get("Last-Modified"); lastModifiedStr != "" {
		lastModified, _ = time.Parse(time.RFC1123, lastModifiedStr)
	}

	// 获取ETag
	etag := props.Get("ETag")
	if etag != "" {
		// 移除ETag的引号
		etag = strings.Trim(etag, "\"")
	}

	// 获取Content-Type
	contentType := props.Get("Content-Type")
	if contentType == "" {
		// 根据文件扩展名推断
		ext := strings.ToLower(filepath.Ext(key))
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
	}

	return &FileInfo{
		Path:         path,
		Size:         size,
		ContentType:  contentType,
		ETag:         etag,
		LastModified: lastModified,
		URL:          s.GetURL(path),
	}, nil
}
