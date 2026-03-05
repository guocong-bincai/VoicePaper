#!/bin/bash
# VoicePaper - 一键部署脚本
# 用于将应用部署到生产服务器

set -e  # 遇到错误立即退出

# 配置变量 - 从 .env 文件或环境变量获取
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ -f "$SCRIPT_DIR/../.env" ]; then
    source "$SCRIPT_DIR/../.env"
fi
SERVER_HOST="${SERVER_IP:?请设置 SERVER_IP 环境变量或在 .env 文件中配置}"
SERVER_USER="${SERVER_USER:-root}"
BACKEND_DIR="/opt/voicepaper/backend"
FRONTEND_DIR="/var/www/voicepaper/frontend/dist"
CADDY_CONFIG="/etc/caddy/Caddyfile"

echo "=========================================="
echo "🚀 VoicePaper 部署脚本"
echo "=========================================="

# 1. 构建前端
echo ""
echo "📦 步骤 1/5: 构建前端..."
cd frontend
npm install
npm run build
cd ..

# 2. 构建后端
echo ""
echo "📦 步骤 2/5: 构建后端..."
cd backend
go build -o server cmd/server/main.go
cd ..

# 3. 上传前端文件
echo ""
echo "📤 步骤 3/5: 上传前端文件..."
ssh ${SERVER_USER}@${SERVER_HOST} "mkdir -p ${FRONTEND_DIR}"
scp -r frontend/dist/* ${SERVER_USER}@${SERVER_HOST}:${FRONTEND_DIR}/

# 4. 上传后端文件
echo ""
echo "📤 步骤 4/5: 上传后端文件..."
ssh ${SERVER_USER}@${SERVER_HOST} "mkdir -p ${BACKEND_DIR}"
ssh ${SERVER_USER}@${SERVER_HOST} "mkdir -p ${BACKEND_DIR}/assets/fonts"  # 创建字体目录
scp backend/server ${SERVER_USER}@${SERVER_HOST}:${BACKEND_DIR}/
scp backend/etc/config_pro.yaml ${SERVER_USER}@${SERVER_HOST}:${BACKEND_DIR}/config.yaml
scp backend/assets/fonts/simhei.ttf ${SERVER_USER}@${SERVER_HOST}:${BACKEND_DIR}/assets/fonts/  # 上传字体文件

# 5. 上传并重载Caddy配置
echo ""
echo "📤 步骤 5/5: 更新Caddy配置..."
scp deploy/Caddyfile ${SERVER_USER}@${SERVER_HOST}:${CADDY_CONFIG}

# 6. 重启服务
echo ""
echo "🔄 重启服务..."
ssh ${SERVER_USER}@${SERVER_HOST} << 'EOF'
# 重启后端服务（假设使用systemd）
systemctl restart voicepaper-backend 2>/dev/null || echo "⚠️  后端服务未配置systemd"

# 重新加载Caddy
caddy reload --config /etc/caddy/Caddyfile

echo "✅ 服务重启完成"
EOF

echo ""
echo "=========================================="
echo "✅ 部署完成！"
echo "=========================================="
echo ""
echo "🌐 访问网站: https://voicepaper.online"
echo "📊 查看日志: ssh ${SERVER_USER}@${SERVER_HOST} 'journalctl -u caddy -f'"
echo ""

