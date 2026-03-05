#!/bin/bash

# VoicePaper Database Initialization Script
# Usage: ./init_db.sh [db_password]

# 颜色
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 加载环境变量
if [ -f ../.env ]; then
    source ../.env
fi

# 默认配置
DB_HOST="127.0.0.1"
DB_PORT="3306"
DB_USER="root"
DB_PASS=${1:-$DB_PASSWORD}
DB_NAME="voice_paper"

if [ -z "$DB_PASS" ]; then
    echo -e "${YELLOW}请输入数据库密码:${NC}"
    read -s DB_PASS
fi

echo -e "${GREEN}正在初始化 VoicePaper 数据库...${NC}"

# 1. 创建数据库
echo -e "${YELLOW}[1/4] 创建数据库 ${DB_NAME}...${NC}"
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS -e "CREATE DATABASE IF NOT EXISTS \\\`${DB_NAME}\\" CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

if [ $? -ne 0 ]; then
    echo -e "${RED}数据库创建失败，请检查密码和连接设置。${NC}"
    exit 1
fi

# 2. 导入表结构
echo -e "${YELLOW}[2/4] 导入表结构 (Schema)...${NC}"
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME < database/schema.sql

if [ $? -ne 0 ]; then
    echo -e "${RED}表结构导入失败。${NC}"
    exit 1
fi

# 3. 导入初始数据
echo -e "${YELLOW}[3/4] 导入初始数据 (Seed)...${NC}"
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME < database/seed.sql

if [ $? -ne 0 ]; then
    echo -e "${RED}初始数据导入失败。${NC}"
    exit 1
fi

# 4. 准备示例音频
echo -e "${YELLOW}[4/4] 准备示例资源...${NC}"
mkdir -p data/audio
# 复制一个现有的音频作为示例（临时方案，防止404）
# 假设从前端复制一个音效文件作为 placeholder
if [ -f "../frontend/public/sounds/right.mp3" ]; then
    cp "../frontend/public/sounds/right.mp3" "data/audio/sample_welcome.mp3"
    echo -e "${GREEN}已创建示例音频: data/audio/sample_welcome.mp3 (使用占位音效)${NC}"
else
    echo -e "${YELLOW}未找到示例音频源，请手动放置音频文件到 backend/data/audio/sample_welcome.mp3${NC}"
fi

echo ""
echo -e "${GREEN}🎉 数据库初始化完成！${NC}"
echo -e "现在你可以启动后端服务了: go run cmd/server/main.go"
