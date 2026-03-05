package storage

import (
	"context"
	"io"
	"time"
)

// FileInfo 文件信息
type FileInfo struct {
	Path         string    // 文件路径/Key
	Size         int64     // 文件大小（字节）
	ContentType  string    // MIME类型
	ETag         string    // ETag（用于校验）
	LastModified time.Time // 最后修改时间
	URL          string    // 访问URL
}

// Storage 存储接口抽象层
// 支持本地存储、对象存储等多种实现
type Storage interface {
	// Save 保存文件
	// path: 存储路径（相对路径或 OSS key）
	// data: 文件内容
	// 返回: 完整访问 URL
	Save(ctx context.Context, path string, data []byte) (string, error)

	// SaveFromReader 从 Reader 保存文件
	SaveFromReader(ctx context.Context, path string, reader io.Reader, size int64) (string, error)

	// SaveFromReaderWithPublicRead 从 Reader 保存文件并设置为公开可读
	// 适用于头像等需要公开访问的文件
	SaveFromReaderWithPublicRead(ctx context.Context, path string, reader io.Reader, size int64) (string, error)

	// Get 获取文件内容
	Get(ctx context.Context, path string) ([]byte, error)

	// GetURL 获取文件的访问 URL
	GetURL(path string) string

	// GetSignedURL 获取文件的签名URL（用于私有Bucket访问）
	// expires: URL有效期（秒）
	GetSignedURL(ctx context.Context, path string, expires int64) (string, error)

	// GetSignedImageURL 获取带图片处理参数的签名URL
	// width: 缩放宽度，0表示不缩放
	GetSignedImageURL(ctx context.Context, path string, expires int64, width int) (string, error)

	// Delete 删除文件
	Delete(ctx context.Context, path string) error

	// Exists 检查文件是否存在
	Exists(ctx context.Context, path string) (bool, error)

	// GetSize 获取文件大小
	GetSize(ctx context.Context, path string) (int64, error)

	// GetFileInfo 获取文件详细信息
	// 包括文件大小、MIME类型、ETag、最后修改时间等
	GetFileInfo(ctx context.Context, path string) (*FileInfo, error)
}
