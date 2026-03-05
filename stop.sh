#!/bin/bash

# 定义颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🛑 正在停止 VoicePaper 服务...${NC}"

# 杀掉占用 8080 端口的进程 (后端)
PID_BACKEND=$(lsof -ti:8080)
if [ -n "$PID_BACKEND" ]; then
  kill -9 $PID_BACKEND
  echo -e "${GREEN}✅ 已停止后端服务 (PID: $PID_BACKEND)${NC}"
else
  echo "❌ 后端服务未运行 (端口 8080 未被占用)"
fi

# 杀掉占用 5173 端口的进程 (前端)
PID_FRONTEND=$(lsof -ti:5173)
if [ -n "$PID_FRONTEND" ]; then
  kill -9 $PID_FRONTEND
  echo -e "${GREEN}✅ 已停止前端服务 (PID: $PID_FRONTEND)${NC}"
else
  echo "❌ 前端服务未运行 (端口 5173 未被占用)"
fi

echo -e "${GREEN}✅ 停止操作完成${NC}"
