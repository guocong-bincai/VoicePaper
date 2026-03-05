#!/bin/bash
# VoicePaper - 仅部署前端
# 用于快速部署前端文件到生产服务器

set -e  # 遇到错误立即退出

# 配置变量 - 从 .env 文件或环境变量获取
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ -f "$SCRIPT_DIR/../.env" ]; then
    source "$SCRIPT_DIR/../.env"
fi
SERVER_HOST="${SERVER_IP:?请设置 SERVER_IP 环境变量或在 .env 文件中配置}"
SERVER_USER="${SERVER_USER:-root}"
FRONTEND_DIR="/var/www/voicepaper/frontend/dist"
CADDY_CONFIG="/etc/caddy/Caddyfile"

echo "=========================================="
echo "🚀 VoicePaper 前端部署"
echo "=========================================="

# 1. 检查是否在项目根目录
if [ ! -d "frontend" ]; then
    echo "❌ 错误: 请在项目根目录运行此脚本"
    exit 1
fi

# 2. 安装依赖
echo ""
echo "[1/4] 检查依赖..."
cd frontend
if [ ! -d "node_modules" ]; then
    echo "正在安装依赖..."
    npm install
else
    echo "✓ 依赖已存在"
fi

# 3. 构建前端
echo ""
echo "[2/4] 构建前端..."
echo "环境: 生产环境"
echo "目标: ES2015+"
npm run build
if [ $? -ne 0 ]; then
    echo "❌ 前端构建失败"
    exit 1
fi
echo "✓ 前端构建成功"
cd ..

# 4. 上传前端文件
echo ""
echo "[3/4] 上传前端文件..."
ssh ${SERVER_USER}@${SERVER_HOST} "mkdir -p ${FRONTEND_DIR}"

# 清空旧文件
ssh ${SERVER_USER}@${SERVER_HOST} "rm -rf ${FRONTEND_DIR}/*"

# 上传新文件
scp -r frontend/dist/* ${SERVER_USER}@${SERVER_HOST}:${FRONTEND_DIR}/
echo "✓ 前端文件上传成功"

# 5. 更新Caddy配置并重载（可选）
echo ""
echo "[4/4] 重载 Caddy..."
if [ -f "deploy/Caddyfile" ]; then
    scp deploy/Caddyfile ${SERVER_USER}@${SERVER_HOST}:${CADDY_CONFIG}
    ssh ${SERVER_USER}@${SERVER_HOST} "caddy reload --config /etc/caddy/Caddyfile"
    echo "✓ Caddy 配置已更新"
else
    ssh ${SERVER_USER}@${SERVER_HOST} "caddy reload --config /etc/caddy/Caddyfile"
    echo "✓ Caddy 已重载"
fi

echo ""
echo "=========================================="
echo "✅ 前端部署完成！"
echo "=========================================="
echo ""
echo "🌐 访问网站: https://voicepaper.online"
echo "🔍 测试前端: curl -I https://voicepaper.online"
echo ""
echo "💡 提示: 如果看到旧内容，请清除浏览器缓存 (Ctrl+Shift+R)"
echo ""
