#!/bin/bash

# VoicePaper 自动化部署脚本
# 作者: DeepV Code AI Assistant
# 用途: 从 Mac 本地部署前后端到生产服务器
# 使用方法: ./deploy.sh [选项]
#   --frontend-only  仅部署前端
#   --backend-only   仅部署后端
#   --caddy-only     仅更新Caddy配置
#   --skip-build     跳过构建步骤

set -e  # 遇到错误立即退出

# 解析命令行参数
DEPLOY_FRONTEND=true
DEPLOY_BACKEND=true
DEPLOY_CADDY=false
SKIP_BUILD=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --frontend-only)
            DEPLOY_BACKEND=false
            DEPLOY_CADDY=false
            shift
            ;;
        --backend-only)
            DEPLOY_FRONTEND=false
            DEPLOY_CADDY=false
            shift
            ;;
        --caddy-only)
            DEPLOY_FRONTEND=false
            DEPLOY_BACKEND=false
            DEPLOY_CADDY=true
            shift
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        *)
            echo "未知参数: $1"
            echo "用法: $0 [--frontend-only] [--backend-only] [--caddy-only] [--skip-build]"
            exit 1
            ;;
    esac
done

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 加载本地环境变量 (.env)
if [ -f .env ]; then
    echo -e "${GREEN}加载配置文件: .env${NC}"
    source .env
fi

# 服务器配置
# 优先从环境变量 (.env) 获取，如果不存在则使用默认占位符
SERVER_IP=${SERVER_IP:-"your_server_ip"}
SERVER_USER=${SERVER_USER:-"root"}
FRONTEND_REMOTE_PATH=${FRONTEND_REMOTE_PATH:-"/var/www/voicepaper/frontend/dist/"}
BACKEND_REMOTE_PATH=${BACKEND_REMOTE_PATH:-"/opt/voicepaper/backend/"}
SERVICE_NAME=${SERVICE_NAME:-"voicepaper-backend"}

# 检查必要配置
if [ "$SERVER_IP" == "your_server_ip" ]; then
    echo -e "${YELLOW}提示: 未设置 SERVER_IP 环境变量，将使用默认占位符。${NC}"
    echo -e "${YELLOW}建议运行方式: SERVER_IP=x.x.x.x ./deploy.sh${NC}"
fi

# 项目路径（自动获取脚本所在目录）
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
FRONTEND_DIR="$PROJECT_ROOT/frontend"
BACKEND_DIR="$PROJECT_ROOT/backend"
DEPLOY_DIR="$PROJECT_ROOT/deploy"
CADDY_REMOTE_PATH="/etc/caddy"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}   VoicePaper 自动化部署脚本${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 显示部署模式
if [ "$DEPLOY_CADDY" = true ]; then
    echo -e "${YELLOW}部署模式: 仅Caddy配置${NC}"
elif [ "$DEPLOY_FRONTEND" = true ] && [ "$DEPLOY_BACKEND" = true ]; then
    echo -e "${YELLOW}部署模式: 前端 + 后端${NC}"
elif [ "$DEPLOY_FRONTEND" = true ]; then
    echo -e "${YELLOW}部署模式: 仅前端${NC}"
elif [ "$DEPLOY_BACKEND" = true ]; then
    echo -e "${YELLOW}部署模式: 仅后端${NC}"
fi
echo ""

# 前端部署
if [ "$DEPLOY_FRONTEND" = true ]; then
    # 步骤1: 构建前端
    if [ "$SKIP_BUILD" = false ]; then
        echo -e "${YELLOW}[1/3] 构建前端...${NC}"
        cd "$FRONTEND_DIR"
        npm run build
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ 前端构建成功${NC}"
        else
            echo -e "${RED}✗ 前端构建失败${NC}"
            exit 1
        fi
        echo ""
    else
        echo -e "${YELLOW}[1/3] 跳过前端构建${NC}"
        echo ""
    fi

    # 步骤2: 上传前端文件
    echo -e "${YELLOW}[2/3] 上传前端文件到服务器...${NC}"
    cd "$FRONTEND_DIR"
    scp -r dist/* ${SERVER_USER}@${SERVER_IP}:${FRONTEND_REMOTE_PATH}
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 前端文件上传成功${NC}"
    else
        echo -e "${RED}✗ 前端文件上传失败${NC}"
        exit 1
    fi
    echo ""

    echo -e "${GREEN}✓ 前端部署完成${NC}"
    echo ""
fi

# 后端部署
if [ "$DEPLOY_BACKEND" = true ]; then
    # 步骤3: 构建后端
    if [ "$SKIP_BUILD" = false ]; then
        echo -e "${YELLOW}[1/4] 构建后端...${NC}"
        cd "$BACKEND_DIR"

        # 显示构建信息
        echo -e "${YELLOW}目标平台: Linux AMD64${NC}"
        echo -e "${YELLOW}构建模式: 生产环境优化${NC}"

        # 交叉编译为 Linux 可执行文件（带优化参数）
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
            -ldflags="-s -w" \
            -trimpath \
            -o server \
            cmd/server/main.go

        if [ $? -eq 0 ]; then
            SERVER_SIZE=$(ls -lh server | awk '{print $5}')
            echo -e "${GREEN}✓ 后端构建成功 (文件大小: $SERVER_SIZE)${NC}"
        else
            echo -e "${RED}✗ 后端构建失败${NC}"
            exit 1
        fi
        echo ""
    else
        echo -e "${YELLOW}[1/4] 跳过后端构建${NC}"
        echo ""
    fi

    # 步骤4: 停止后端服务
    echo -e "${YELLOW}[2/4] 停止后端服务...${NC}"
    ssh ${SERVER_USER}@${SERVER_IP} "systemctl stop ${SERVICE_NAME}"
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 后端服务已停止${NC}"
    else
        echo -e "${RED}✗ 停止后端服务失败${NC}"
        exit 1
    fi
    echo ""

    # 步骤5: 上传后端文件
    echo -e "${YELLOW}[3/4] 上传后端文件到服务器...${NC}"
    cd "$BACKEND_DIR"
    # 确保远程 etc 目录存在
    ssh ${SERVER_USER}@${SERVER_IP} "mkdir -p ${BACKEND_REMOTE_PATH}/etc"
    # 上传二进制文件
    scp server ${SERVER_USER}@${SERVER_IP}:${BACKEND_REMOTE_PATH}
    # 上传生产环境配置文件 (config_pro.yaml -> config.yaml)
    scp etc/config_pro.yaml ${SERVER_USER}@${SERVER_IP}:${BACKEND_REMOTE_PATH}/etc/config.yaml

    # 上传静态资源 (字体文件)
    echo -e "${YELLOW}上传静态资源 (assets)...${NC}"
    scp -r assets ${SERVER_USER}@${SERVER_IP}:${BACKEND_REMOTE_PATH}

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 后端文件及配置上传成功${NC}"
    else
        echo -e "${RED}✗ 后端文件上传失败${NC}"
        exit 1
    fi
    echo ""

    # 步骤6: 启动后端服务
    echo -e "${YELLOW}[4/4] 启动后端服务...${NC}"
    ssh ${SERVER_USER}@${SERVER_IP} "systemctl start ${SERVICE_NAME}"
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 后端服务已启动${NC}"
    else
        echo -e "${RED}✗ 启动后端服务失败${NC}"
        exit 1
    fi
    echo ""

    # 检查服务状态
    echo -e "${YELLOW}检查服务状态...${NC}"
    ssh ${SERVER_USER}@${SERVER_IP} "systemctl status ${SERVICE_NAME} --no-pager | head -n 20"
    echo ""

    # 检查远程服务器是否安装了 Chromium/Chrome
    echo -e "${YELLOW}检查远程服务器依赖 (Chromium/Chrome)...${NC}"
    CHROMIUM_CHECK=$(ssh ${SERVER_USER}@${SERVER_IP} "which chromium-browser || which chromium || which google-chrome || echo 'NOT_FOUND'")
    if echo "$CHROMIUM_CHECK" | grep -q "NOT_FOUND"; then
        echo -e "${RED}⚠️  警告: 远程服务器未检测到 Chromium 或 Chrome 浏览器！${NC}"
        echo -e "${RED}   PDF 导出功能需要依赖浏览器。请登录服务器执行以下命令安装:${NC}"
        echo -e "${YELLOW}   CentOS: sudo yum install -y chromium${NC}"
        echo -e "${YELLOW}   Ubuntu: sudo apt-get install -y chromium-browser${NC}"
        echo ""
    else
        echo -e "${GREEN}✓ 检测到浏览器: $CHROMIUM_CHECK${NC}"

        # 检查浏览器是否能正常运行（PDF导出功能测试）
        echo -e "${YELLOW}测试 PDF 导出环境...${NC}"
        ssh ${SERVER_USER}@${SERVER_IP} << 'REMOTE_TEST'
# 测试 Chromium 是否能正常生成 PDF（带国内服务器优化参数）
CHROMIUM_PATH=$(which chromium-browser || which chromium || which google-chrome)
if [ -n "$CHROMIUM_PATH" ]; then
    timeout 15 $CHROMIUM_PATH \
        --headless \
        --disable-gpu \
        --no-sandbox \
        --disable-dev-shm-usage \
        --disable-background-networking \
        --disable-component-update \
        --disable-sync \
        --disable-extensions \
        --no-first-run \
        --print-to-pdf=/tmp/chromium_deploy_test.pdf \
        about:blank 2>&1 | head -5

    if [ -f /tmp/chromium_deploy_test.pdf ]; then
        PDF_SIZE=$(ls -lh /tmp/chromium_deploy_test.pdf | awk '{print $5}')
        echo "✓ PDF 导出测试成功 (文件大小: $PDF_SIZE)"
        rm -f /tmp/chromium_deploy_test.pdf
        exit 0
    else
        echo "✗ PDF 导出测试失败"
        exit 1
    fi
else
    echo "✗ 未找到浏览器可执行文件"
    exit 1
fi
REMOTE_TEST

        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ PDF 导出环境测试通过${NC}"
        else
            echo -e "${RED}⚠️  PDF 导出环境测试失败，可能需要安装额外依赖${NC}"
            echo -e "${YELLOW}   解决方案: ssh ${SERVER_USER}@${SERVER_IP} './diagnose_server.sh'${NC}"
        fi
        echo ""
    fi

    # 上传诊断脚本到服务器（方便后续排查问题）
    echo -e "${YELLOW}上传诊断工具到服务器...${NC}"
    scp "$PROJECT_ROOT/diagnose_server.sh" ${SERVER_USER}@${SERVER_IP}:${BACKEND_REMOTE_PATH}/
    ssh ${SERVER_USER}@${SERVER_IP} "chmod +x ${BACKEND_REMOTE_PATH}/diagnose_server.sh"
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 诊断工具已上传 (${BACKEND_REMOTE_PATH}/diagnose_server.sh)${NC}"
    fi
    echo ""
fi

# Caddy配置部署
if [ "$DEPLOY_CADDY" = true ]; then
    echo -e "${YELLOW}[1/2] 上传Caddyfile到服务器...${NC}"
    cd "$DEPLOY_DIR"
    scp Caddyfile ${SERVER_USER}@${SERVER_IP}:${CADDY_REMOTE_PATH}/Caddyfile
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Caddyfile上传成功${NC}"
    else
        echo -e "${RED}✗ Caddyfile上传失败${NC}"
        exit 1
    fi
    echo ""

    echo -e "${YELLOW}[2/2] 重载Caddy配置...${NC}"
    ssh ${SERVER_USER}@${SERVER_IP} "systemctl reload caddy"
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Caddy配置重载成功${NC}"
    else
        echo -e "${RED}✗ Caddy配置重载失败${NC}"
        exit 1
    fi
    echo ""

    echo -e "${YELLOW}检查Caddy状态...${NC}"
    ssh ${SERVER_USER}@${SERVER_IP} "systemctl status caddy --no-pager | head -n 15"
    echo ""
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}   🎉 部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 部署后自动测试
if [ "$DEPLOY_BACKEND" = true ]; then
    echo -e "${YELLOW}🧪 运行部署后验证测试...${NC}"
    sleep 3  # 等待服务完全启动

    # 测试1: 检查服务是否运行
    echo -e "${YELLOW}1. 检查后端服务状态...${NC}"
    SERVICE_STATUS=$(ssh ${SERVER_USER}@${SERVER_IP} "systemctl is-active ${SERVICE_NAME}")
    if [ "$SERVICE_STATUS" = "active" ]; then
        echo -e "${GREEN}   ✓ 后端服务运行正常${NC}"
    else
        echo -e "${RED}   ✗ 后端服务未运行 (状态: $SERVICE_STATUS)${NC}"
    fi

    # 测试2: 检查端口监听
    echo -e "${YELLOW}2. 检查端口监听...${NC}"
    PORT_CHECK=$(ssh ${SERVER_USER}@${SERVER_IP} "ss -tuln | grep :8080 || netstat -tuln | grep :8080 || echo 'NOT_LISTENING'")
    if echo "$PORT_CHECK" | grep -q "NOT_LISTENING"; then
        echo -e "${RED}   ✗ 端口 8080 未监听${NC}"
    else
        echo -e "${GREEN}   ✓ 端口 8080 正常监听${NC}"
    fi

    # 测试3: 检查字体文件
    echo -e "${YELLOW}3. 检查静态资源 (字体文件)...${NC}"
    FONT_CHECK=$(ssh ${SERVER_USER}@${SERVER_IP} "test -f ${BACKEND_REMOTE_PATH}/assets/fonts/simhei.ttf && echo 'OK' || echo 'MISSING'")
    if [ "$FONT_CHECK" = "OK" ]; then
        echo -e "${GREEN}   ✓ 字体文件存在${NC}"
    else
        echo -e "${RED}   ✗ 字体文件缺失${NC}"
    fi

    # 测试4: HTTP 健康检查
    echo -e "${YELLOW}4. HTTP 健康检查...${NC}"
    HTTP_CHECK=$(curl -s -o /dev/null -w "%{http_code}" https://voicepaper.top/api/v1/articles 2>/dev/null || echo "FAILED")
    if [ "$HTTP_CHECK" = "200" ]; then
        echo -e "${GREEN}   ✓ API 响应正常 (HTTP 200)${NC}"
    else
        echo -e "${RED}   ✗ API 无响应或异常 (HTTP $HTTP_CHECK)${NC}"
    fi

    echo ""
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}访问地址: https://voicepaper.online${NC}"
echo -e "${GREEN}访问地址: https://voicepaper.top${NC}"
echo ""
echo -e "${YELLOW}📋 常用命令:${NC}"
echo -e "${YELLOW}  查看日志: ssh ${SERVER_USER}@${SERVER_IP} 'journalctl -u ${SERVICE_NAME} -f'${NC}"
echo -e "${YELLOW}  重启服务: ssh ${SERVER_USER}@${SERVER_IP} 'systemctl restart ${SERVICE_NAME}'${NC}"
echo -e "${YELLOW}  运行诊断: ssh ${SERVER_USER}@${SERVER_IP} 'cd ${BACKEND_REMOTE_PATH} && ./diagnose_server.sh'${NC}"
echo ""
