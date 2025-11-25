#!/bin/bash
# VoicePaper 前端启动脚本

echo "🚀 启动 VoicePaper 前端服务..."
echo "📂 工作目录: $(pwd)"
echo ""

# 检查Python是否安装
if command -v python3 &> /dev/null; then
    echo "✅ 使用 Python3 启动 HTTP 服务器"
    echo "🌐 访问地址: http://localhost:8000/frontend/"
    echo "⚠️  按 Ctrl+C 停止服务"
    echo ""
    python3 -m http.server 8000
elif command -v python &> /dev/null; then
    echo "✅ 使用 Python2 启动 HTTP 服务器"
    echo "🌐 访问地址: http://localhost:8000/frontend/"
    echo "⚠️  按 Ctrl+C 停止服务"
    echo ""
    python -m SimpleHTTPServer 8000
else
    echo "❌ 错误: 未找到 Python"
    echo "请安装 Python 或使用其他方式启动服务器"
    echo ""
    echo "其他启动方式:"
    echo "  - Node.js: npx http-server -p 8000"
    echo "  - VS Code: 使用 Live Server 插件"
    exit 1
fi

