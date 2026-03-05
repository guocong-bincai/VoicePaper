#!/bin/bash
# VoicePaper - 仅部署后端
# 用于快速部署后端服务到生产服务器

set -e  # 遇到错误立即退出

# 配置变量 - 从 .env 文件或环境变量获取
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ -f "$SCRIPT_DIR/../.env" ]; then
    source "$SCRIPT_DIR/../.env"
fi
SERVER_HOST="${SERVER_IP:?请设置 SERVER_IP 环境变量或在 .env 文件中配置}"
SERVER_USER="${SERVER_USER:-root}"
BACKEND_DIR="/opt/voicepaper/backend"

echo "=========================================="
echo "🚀 VoicePaper 后端部署"
echo "=========================================="

# 1. 检查是否在项目根目录
if [ ! -d "backend" ]; then
    echo "❌ 错误: 请在项目根目录运行此脚本"
    exit 1
fi

# 2. 构建后端
echo ""
echo "[1/4] 构建后端..."
echo "目标平台: Linux AMD64"
echo "构建模式: 生产环境优化"
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server cmd/server/main.go
if [ $? -ne 0 ]; then
    echo "❌ 后端构建失败"
    exit 1
fi
echo "✓ 后端构建成功"
cd ..

# 3. 上传后端文件
echo ""
echo "[2/4] 上传后端文件..."
ssh ${SERVER_USER}@${SERVER_HOST} "mkdir -p ${BACKEND_DIR}"
ssh ${SERVER_USER}@${SERVER_HOST} "mkdir -p ${BACKEND_DIR}/assets/fonts"
scp backend/server ${SERVER_USER}@${SERVER_HOST}:${BACKEND_DIR}/
scp backend/etc/config_pro.yaml ${SERVER_USER}@${SERVER_HOST}:${BACKEND_DIR}/config.yaml

# 上传字体文件（如果存在）
if [ -f "backend/assets/fonts/simhei.ttf" ]; then
    scp backend/assets/fonts/simhei.ttf ${SERVER_USER}@${SERVER_HOST}:${BACKEND_DIR}/assets/fonts/
    echo "✓ 字体文件上传成功"
fi

echo "✓ 后端文件上传成功"

# 4. 设置执行权限
echo ""
echo "[3/4] 设置权限..."
ssh ${SERVER_USER}@${SERVER_HOST} "chmod +x ${BACKEND_DIR}/server"
echo "✓ 权限设置完成"

# 5. 重启后端服务
echo ""
echo "[4/4] 重启后端服务..."
ssh ${SERVER_USER}@${SERVER_HOST} << 'EOF'
# 方式1: 使用 systemd (推荐)
if systemctl is-active --quiet voicepaper-backend; then
    systemctl restart voicepaper-backend
    echo "✓ systemd 服务重启成功"
# 方式2: 手动重启
else
    echo "⚠️  systemd 服务未配置，尝试手动重启..."
    pkill -f '/opt/voicepaper/backend/server' || true
    sleep 2
    cd /opt/voicepaper/backend
    nohup ./server > /dev/null 2>&1 &
    echo "✓ 后端服务已启动"
fi

# 检查服务状态
sleep 2
if curl -s http://localhost:8080/api/v1/ping > /dev/null; then
    echo "✅ 后端服务运行正常"
else
    echo "⚠️  后端服务可能未正常启动，请检查日志"
fi
EOF

echo ""
echo "=========================================="
echo "✅ 后端部署完成！"
echo "=========================================="
echo ""
echo "📊 查看日志:"
echo "   ssh ${SERVER_USER}@${SERVER_HOST} 'journalctl -u voicepaper-backend -f'"
echo ""
echo "🔍 测试接口:"
echo "   curl https://voicepaper.online/api/v1/ping"
echo ""
