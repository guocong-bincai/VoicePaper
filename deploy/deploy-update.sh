#!/bin/bash
# VoicePaper - 快速更新部署脚本
# 用于更新已部署的生产环境

set -e  # 遇到错误立即退出

# 配置变量 - 从 .env 文件或环境变量获取
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ -f "$SCRIPT_DIR/../.env" ]; then
    source "$SCRIPT_DIR/../.env"
fi
SERVER="${SERVER_USER:-root}@${SERVER_IP:?请设置 SERVER_IP 环境变量或在 .env 文件中配置}"
FRONTEND_DIR="/var/www/voicepaper/frontend/dist"
BACKEND_DIR="/opt/voicepaper/backend"

echo "=========================================="
echo "🚀 VoicePaper 快速更新部署"
echo "=========================================="

# 1. 构建前端
echo ""
echo "📦 步骤 1/5: 构建前端..."
cd frontend
npm run build
cd ..

# 2. 上传前端
echo ""
echo "📤 步骤 2/5: 上传前端文件..."
scp -r frontend/dist/* ${SERVER}:${FRONTEND_DIR}/

# 3. 编译后端
echo ""
echo "🔨 步骤 3/5: 编译后端..."
cd backend
GOOS=linux GOARCH=amd64 go build -o server cmd/server/main.go
cd ..

# 4. 停止后端服务
echo ""
echo "⏸️  步骤 4/6: 停止后端服务..."
ssh ${SERVER} "systemctl stop voicepaper-backend"

# 5. 上传后端
echo ""
echo "📤 步骤 5/6: 上传后端..."
scp backend/server ${SERVER}:${BACKEND_DIR}/

# 6. 启动服务
echo ""
echo "🔄 步骤 6/6: 启动后端服务..."
ssh ${SERVER} "systemctl start voicepaper-backend && systemctl status voicepaper-backend --no-pager -l"

echo ""
echo "=========================================="
echo "✅ 部署完成！"
echo "=========================================="
echo ""
echo "🌐 访问网站: https://voicepaper.online"
echo "📊 查看日志: ssh ${SERVER} 'journalctl -u voicepaper-backend -f'"
echo ""

