#!/bin/bash
# VoicePaper 前端启动脚本

echo "🚀 启动 VoicePaper 前端开发服务器..."
echo "📂 工作目录: $(pwd)"
echo ""

cd frontend

# 检查 node_modules 是否存在
if [ ! -d "node_modules" ]; then
    echo "📦 安装依赖..."
    npm install
fi

echo "✅ 使用 Vite 启动前端开发服务器"
echo "🌐 访问地址: http://localhost:5173/"
echo "⚠️  按 Ctrl+C 停止服务"
echo ""

npm run dev