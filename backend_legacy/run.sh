#!/bin/bash
# VoicePaper 后端运行脚本

echo "🎤 VoicePaper - 语音合成服务"
echo "================================"
echo ""

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到 Go"
    echo "请安装 Go 1.21+ 后再运行"
    exit 1
fi

echo "✅ Go 版本: $(go version)"
echo ""

# 检查data目录
if [ ! -d "../data" ]; then
    echo "⚠️  警告: data 目录不存在，正在创建..."
    mkdir -p ../data
fi

# 检查输入文件
if [ ! -f "../data/1.md" ]; then
    echo "⚠️  警告: ../data/1.md 不存在"
    echo "请在 data 目录下创建 1.md 文件"
    exit 1
fi

echo "📄 输入文件: ../data/1.md"
echo "📊 文件大小: $(wc -c < ../data/2.md) 字节"
echo ""
echo "🚀 开始生成音频..."
echo "⏳ 这可能需要几分钟时间，请耐心等待..."
echo ""

# 运行Go程序
go run main.go

if [ $? -eq 0 ]; then
    echo ""
    echo "================================"
    echo "✅ 音频生成成功！"
    echo "📁 输出位置: ../data/output.mp3"
    echo ""
    echo "下一步："
    echo "  1. 运行前端: cd .. && ./start-frontend.sh"
    echo "  2. 访问: http://localhost:8001/frontend/"
else
    echo ""
    echo "❌ 音频生成失败，请检查错误信息"
    exit 1
fi

