#!/bin/bash
# VoicePaper - 服务器初始化脚本
# 用于在新服务器上安装所有必要的环境和依赖

set -e  # 遇到错误立即退出

echo "=========================================="
echo "🚀 VoicePaper 服务器初始化脚本"
echo "=========================================="

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    echo "请使用root权限运行此脚本: sudo bash server-setup.sh"
    exit 1
fi

# 1. 更新系统
echo ""
echo "📦 步骤 1/8: 更新系统..."
apt update
apt upgrade -y

# 2. 安装基础工具
echo ""
echo "📦 步骤 2/8: 安装基础工具..."
apt install -y curl wget git vim unzip build-essential

# 3. 安装 Go 1.21+
echo ""
echo "📦 步骤 3/8: 安装 Go 1.21..."
if command -v go &> /dev/null; then
    echo "✅ Go 已安装: $(go version)"
else
    wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
    rm go1.21.6.linux-amd64.tar.gz

    # 配置环境变量
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
    source ~/.bashrc

    # 立即生效
    export PATH=$PATH:/usr/local/go/bin
    echo "✅ Go 安装完成: $(go version)"
fi

# 4. 安装 Node.js 18+
echo ""
echo "📦 步骤 4/8: 安装 Node.js 18..."
if command -v node &> /dev/null; then
    echo "✅ Node.js 已安装: $(node --version)"
else
    curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
    apt install -y nodejs
    echo "✅ Node.js 安装完成: $(node --version)"
    echo "✅ npm 版本: $(npm --version)"
fi

# 5. 安装 Caddy
echo ""
echo "📦 步骤 5/8: 安装 Caddy..."
if command -v caddy &> /dev/null; then
    echo "✅ Caddy 已安装: $(caddy version)"
else
    apt install -y debian-keyring debian-archive-keyring apt-transport-https
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
    apt update
    apt install -y caddy
    echo "✅ Caddy 安装完成: $(caddy version)"
fi

# 6. 创建项目目录
echo ""
echo "📁 步骤 6/8: 创建项目目录..."
mkdir -p /opt/voicepaper
mkdir -p /var/www/voicepaper/frontend/dist
mkdir -p /var/log/caddy
mkdir -p /etc/caddy

# 设置目录权限
chown -R caddy:caddy /var/log/caddy
chown -R caddy:caddy /var/www/voicepaper

# 7. 克隆代码
echo ""
echo "📥 步骤 7/8: 克隆代码仓库..."
cd /opt/voicepaper
if [ -d "voice_paper" ]; then
    echo "⚠️  代码目录已存在，拉取最新代码..."
    cd voice_paper
    git pull
else
    echo "📥 克隆代码仓库..."
    git clone "${GIT_REPO_URL:?请设置 GIT_REPO_URL 环境变量，例如 https://github.com/yourname/voicepaper.git}"
    cd voice_paper
fi

echo "✅ 代码克隆完成！当前分支: $(git branch --show-current)"

# 8. 配置防火墙
echo ""
echo "🔥 步骤 8/8: 配置防火墙..."
if command -v ufw &> /dev/null; then
    ufw allow 22/tcp    # SSH
    ufw allow 80/tcp    # HTTP
    ufw allow 443/tcp   # HTTPS
    ufw allow 8080/tcp  # Backend (可选，后面可以关闭)
    echo "✅ 防火墙端口已开放"
else
    echo "⚠️  未检测到ufw防火墙，请手动配置端口: 22, 80, 443, 8080"
fi

# 完成提示
echo ""
echo "=========================================="
echo "✅ 服务器初始化完成！"
echo "=========================================="
echo ""
echo "📋 环境信息："
echo "   Go版本: $(go version 2>/dev/null || echo '未安装')"
echo "   Node版本: $(node --version 2>/dev/null || echo '未安装')"
echo "   npm版本: $(npm --version 2>/dev/null || echo '未安装')"
echo "   Caddy版本: $(caddy version 2>/dev/null || echo '未安装')"
echo "   代码目录: /opt/voicepaper/voice_paper"
echo ""
echo "📋 接下来的步骤："
echo ""
echo "1️⃣  复制生产环境配置文件到服务器："
echo "   scp backend/etc/config_pro.yaml root@<YOUR_SERVER_IP>:/opt/voicepaper/voice_paper/backend/etc/"
echo ""
echo "2️⃣  构建和部署项目："
echo "   cd /opt/voicepaper/voice_paper"
echo "   bash deploy/build-and-deploy-local.sh"
echo ""
echo "3️⃣  或者从本地一键部署："
echo "   ./deploy/deploy.sh"
echo ""
echo "=========================================="

