package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// 如果我们在 backend 目录下运行，我们需要向上移动一级到项目根目录
	// 这样才能正确服务 /frontend 和 /data 目录
	if filepath.Base(workDir) == "backend" {
		workDir = filepath.Dir(workDir)
	}

	log.Printf("🚀 VoicePaper 文件服务器启动")
	log.Printf("📂 服务根目录: %s", workDir)
	log.Printf("🌐 访问地址: http://localhost:8000/frontend/")

	// 创建文件服务器处理程序
	// http.FileServer 默认支持 Range 请求 (206 Partial Content)
	// 这对于音频/视频的拖动播放至关重要
	fs := http.FileServer(http.Dir(workDir))

	// 包装处理程序以添加 CORS 头（如果需要）和日志
	http.Handle("/", corsMiddleware(loggingMiddleware(fs)))

	// 启动服务器
	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("服务器启动失败: ", err)
	}
}

// 日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// CORS 中间件 (允许跨域，虽然本地开发可能不需要，但加上更保险)
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Range")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
