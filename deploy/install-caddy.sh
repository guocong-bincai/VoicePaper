#!/bin/bash
# VoicePaper - Caddy安装脚本
# 适用于 Ubuntu/Debian 系统

set -e  # 遇到错误立即退出

echo "=========================================="
echo "VoicePaper - Caddy 安装脚本"
echo "=========================================="

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    echo "请使用root权限运行此脚本: sudo bash install-caddy.sh"
    exit 1
fi

# 1. 安装依赖
echo ""
echo "📦 步骤 1/6: 安装依赖包..."
apt update
apt install -y debian-keyring debian-archive-keyring apt-transport-https curl

# 2. 添加Caddy官方仓库
echo ""
echo "📦 步骤 2/6: 添加Caddy官方仓库..."
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list

# 3. 安装Caddy
echo ""
echo "📦 步骤 3/6: 安装Caddy..."
apt update
apt install -y caddy

# 4. 检查Caddy版本
echo ""
echo "✅ 步骤 4/6: 检查Caddy版本..."
caddy version

# 5. 创建必要的目录
echo ""
echo "📁 步骤 5/6: 创建必要的目录..."
mkdir -p /etc/caddy
mkdir -p /var/log/caddy
mkdir -p /var/www/voicepaper/frontend/dist
mkdir -p /opt/voicepaper/backend

# 设置目录权限
chown -R caddy:caddy /var/log/caddy
chown -R caddy:caddy /var/www/voicepaper

# 6. 配置防火墙（如果使用ufw）
echo ""
echo "🔥 步骤 6/6: 配置防火墙..."
if command -v ufw &> /dev/null; then
    ufw allow 80/tcp
    ufw allow 443/tcp
    echo "✅ 已开放 80 和 443 端口"
else
    echo "⚠️  未检测到ufw防火墙，请手动开放 80 和 443 端口"
fi

# 完成提示
echo ""
echo "=========================================="
echo "✅ Caddy 安装完成！"
echo "=========================================="
echo ""
echo "📋 接下来的步骤："
echo ""
echo "1️⃣  复制Caddyfile到服务器："
echo "   scp deploy/Caddyfile root@<YOUR_SERVER_IP>:/etc/caddy/Caddyfile"
echo ""
echo "2️⃣  上传前端构建产物："
echo "   cd frontend && npm run build"
echo "   scp -r dist/* root@<YOUR_SERVER_IP>:/var/www/voicepaper/frontend/dist/"
echo ""
echo "3️⃣  上传后端程序："
echo "   cd backend && go build -o server cmd/server/main.go"
echo "   scp server root@<YOUR_SERVER_IP>:/opt/voicepaper/backend/"
echo "   scp etc/config_pro.yaml root@<YOUR_SERVER_IP>:/opt/voicepaper/backend/config.yaml"
echo ""
echo "4️⃣  重新加载Caddy配置："
echo "   caddy reload --config /etc/caddy/Caddyfile"
echo ""
echo "5️⃣  启动Caddy服务："
echo "   systemctl enable caddy"
echo "   systemctl start caddy"
echo "   systemctl status caddy"
echo ""
echo "6️⃣  查看日志："
echo "   journalctl -u caddy -f"
echo "   tail -f /var/log/caddy/voicepaper.log"
echo ""
echo "=========================================="

