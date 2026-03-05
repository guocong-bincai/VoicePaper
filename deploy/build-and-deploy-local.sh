#!/bin/bash
# VoicePaper - 在服务器上本地构建和部署
# 适用于已经克隆了代码的服务器

set -e  # 遇到错误立即退出

PROJECT_ROOT="/opt/voicepaper/voice_paper"
BACKEND_BIN="/opt/voicepaper/backend"
FRONTEND_DIST="/var/www/voicepaper/frontend/dist"

echo "=========================================="
echo "🚀 VoicePaper 本地构建部署"
echo "=========================================="

# 检查项目目录
if [ ! -d "$PROJECT_ROOT" ]; then
    echo "❌ 项目目录不存在: $PROJECT_ROOT"
    echo "请先运行 server-setup.sh 初始化服务器"
    exit 1
fi

cd $PROJECT_ROOT

# 1. 拉取最新代码
echo ""
echo "📥 步骤 1/6: 拉取最新代码..."
git pull

# 2. 构建后端
echo ""
echo "🔨 步骤 2/6: 构建后端..."
cd backend

# 配置Go环境（国内加速）
export GOPROXY=https://goproxy.cn,direct
export GO111MODULE=on

# 下载依赖
go mod download

# 构建
go build -o server cmd/server/main.go

# 创建目标目录
mkdir -p $BACKEND_BIN

# 复制可执行文件和配置
cp server $BACKEND_BIN/
cp etc/config_pro.yaml $BACKEND_BIN/config.yaml

echo "✅ 后端构建完成"

# 3. 构建前端
echo ""
echo "🔨 步骤 3/6: 构建前端..."
cd $PROJECT_ROOT/frontend

# 安装依赖（国内加速）
npm config set registry https://registry.npmmirror.com
npm install

# 构建
npm run build

# 复制到目标目录
mkdir -p $FRONTEND_DIST
rm -rf $FRONTEND_DIST/*
cp -r dist/* $FRONTEND_DIST/

echo "✅ 前端构建完成"

# 4. 配置 Systemd 服务
echo ""
echo "⚙️  步骤 4/6: 配置后端服务..."
cd $PROJECT_ROOT

# 复制 systemd 配置
cp deploy/voicepaper-backend.service /etc/systemd/system/

# 重新加载 systemd
systemctl daemon-reload

# 启用服务
systemctl enable voicepaper-backend

echo "✅ 后端服务已配置"

# 5. 配置 Caddy
echo ""
echo "⚙️  步骤 5/6: 配置 Caddy..."
cp deploy/Caddyfile /etc/caddy/Caddyfile

# 验证配置
if caddy validate --config /etc/caddy/Caddyfile; then
    echo "✅ Caddy 配置验证通过"
else
    echo "❌ Caddy 配置有误"
    exit 1
fi

# 6. 启动服务
echo ""
echo "🚀 步骤 6/6: 启动服务..."

# 重启后端
systemctl restart voicepaper-backend

# 等待后端启动
sleep 2

# 检查后端状态
if systemctl is-active --quiet voicepaper-backend; then
    echo "✅ 后端服务启动成功"
else
    echo "❌ 后端服务启动失败"
    journalctl -u voicepaper-backend -n 20
    exit 1
fi

# 重启/重新加载 Caddy
if systemctl is-active --quiet caddy; then
    caddy reload --config /etc/caddy/Caddyfile
    echo "✅ Caddy 配置已重新加载"
else
    systemctl start caddy
    systemctl enable caddy
    echo "✅ Caddy 服务启动成功"
fi

# 完成
echo ""
echo "=========================================="
echo "✅ 部署完成！"
echo "=========================================="
echo ""
echo "📊 服务状态："
systemctl status voicepaper-backend --no-pager -l | head -10
echo ""
systemctl status caddy --no-pager -l | head -10
echo ""
echo "🌐 访问地址: https://voicepaper.online"
echo ""
echo "📋 常用命令："
echo "   查看后端日志: journalctl -u voicepaper-backend -f"
echo "   查看Caddy日志: journalctl -u caddy -f"
echo "   重启后端: systemctl restart voicepaper-backend"
echo "   重新加载Caddy: caddy reload --config /etc/caddy/Caddyfile"
echo ""

