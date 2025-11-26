#!/bin/bash
# VoicePaper 前端启动脚本

echo "🚀 启动 VoicePaper 服务..."
echo "📂 工作目录: $(pwd)"
echo ""

# 检查 Go 是否安装
if command -v go &> /dev/null; then
    echo "✅ 使用 Go 启动文件服务器 (支持 Range 请求)"
    echo "🌐 访问地址: http://localhost:8000/frontend/"
    echo "⚠️  按 Ctrl+C 停止服务"
    echo ""

    # 运行 Go 服务器
    go run backend/server.go
else
    echo "❌ 错误: 未找到 Go 环境"
    echo "请安装 Go 或使用 Python 启动 (但可能不支持音频拖动)"
    echo ""
    echo "备选方案 (Python):"
    echo "python3 -m http.server 8000"
    exit 1
fi