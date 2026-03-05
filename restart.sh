#!/bin/bash

# 定义颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 捕获 Ctrl+C 信号，清理后台进程
cleanup() {
    echo -e "\n${YELLOW}🛑 收到退出信号，正在停止所有服务...${NC}"

    # 杀掉后端进程
    if [ -n "$BACKEND_PID" ]; then
        kill -9 $BACKEND_PID 2>/dev/null
        echo -e "${GREEN}✅ 已停止后端服务 (PID: $BACKEND_PID)${NC}"
    fi

    # 杀掉前端进程
    if [ -n "$FRONTEND_PID" ]; then
        kill -9 $FRONTEND_PID 2>/dev/null
        echo -e "${GREEN}✅ 已停止前端服务 (PID: $FRONTEND_PID)${NC}"
    fi

    # 额外保险：通过端口再次清理
    PID_8080=$(lsof -ti:8080)
    PID_5173=$(lsof -ti:5173)
    [ -n "$PID_8080" ] && kill -9 $PID_8080 2>/dev/null
    [ -n "$PID_5173" ] && kill -9 $PID_5173 2>/dev/null

    echo -e "${GREEN}✅ 所有服务已停止${NC}"
    exit 0
}

# 设置信号捕获
trap cleanup SIGINT SIGTERM

echo -e "${YELLOW}🔄 正在准备重启 VoicePaper 服务...${NC}"

# 1. 杀掉旧进程 (通过端口)
# 后端端口: 8080
# 前端端口: 5173

echo -e "${YELLOW}🔪 正在清理旧进程...${NC}"

# 杀掉占用 8080 端口的进程 (后端)
PID_BACKEND=$(lsof -ti:8080)
if [ -n "$PID_BACKEND" ]; then
  kill -9 $PID_BACKEND
  echo -e "${GREEN}✅ 已结束后端进程 (PID: $PID_BACKEND)${NC}"
else
  echo "后端端口 (8080) 未被占用"
fi

# 杀掉占用 5173 端口的进程 (前端)
PID_FRONTEND=$(lsof -ti:5173)
if [ -n "$PID_FRONTEND" ]; then
  kill -9 $PID_FRONTEND
  echo -e "${GREEN}✅ 已结束前端进程 (PID: $PID_FRONTEND)${NC}"
else
  echo "前端端口 (5173) 未被占用"
fi

echo "-----------------------------------"

# 2. 启动后端
echo -e "${YELLOW}🚀 正在启动后端服务 (Go)...${NC}"
cd backend || exit
# 使用 zsh -i -c 确保加载环境变量，并指定使用 config_pro.yaml
nohup zsh -i -c "CONFIG_PATH=etc/config_pro.yaml go run cmd/server/main.go" > backend.log 2>&1 &
BACKEND_PID=$!
echo -e "${GREEN}✅ 后端服务已在后台启动 (PID: $BACKEND_PID, 配置: config_pro.yaml)${NC}"
echo "日志文件: backend/backend.log"

# 回到根目录
cd ..

# 3. 启动前端
echo -e "${YELLOW}🚀 正在启动前端服务 (Vite)...${NC}"
cd frontend || exit
# 使用 zsh -i -c 确保加载环境变量 (如 PATH)
nohup zsh -i -c "npm run dev" > frontend.log 2>&1 &
FRONTEND_PID=$!
echo -e "${GREEN}✅ 前端服务已在后台启动 (PID: $FRONTEND_PID)${NC}"
echo "日志文件: frontend/frontend.log"

# 回到根目录
cd ..

echo "-----------------------------------"
echo -e "${GREEN}🎉 VoicePaper 服务重启完成！${NC}"
echo -e "后端地址: http://localhost:8080"
echo -e "前端地址: http://localhost:5173"

echo "-----------------------------------"
echo -e "${YELLOW}📋 正在实时显示日志 (按 Ctrl+C 可停止所有服务并退出)${NC}"
echo -e "${RED}💡 提示: 如果只想退出日志查看但保持服务运行，请直接关闭终端窗口${NC}"
sleep 1

# 等待日志文件生成
sleep 2

# 实时显示日志
tail -f backend/backend.log frontend/frontend.log
