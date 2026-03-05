#!/bin/bash
# VoicePaper 环境变量快速设置脚本

echo "🔧 VoicePaper 环境变量设置"
echo "================================"
echo ""

# 检查是否已存在 .env 文件
if [ -f ".env" ]; then
    echo "⚠️  .env 文件已存在"
    read -p "是否覆盖现有 .env 文件? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ 已取消操作"
        exit 1
    fi
fi

# 从 env.example 复制
if [ -f "env.example" ]; then
    cp env.example .env
    echo "✅ 已创建 .env 文件（从 env.example）"
    echo ""
    echo "📝 请编辑 .env 文件，填入你的实际配置："
    echo "   - MINIMAX_API_KEY"
    echo "   - DB_PASSWORD"
    echo "   - 其他数据库配置（如需要）"
    echo ""
    echo "💡 提示：编辑完成后，使用以下命令加载环境变量："
    echo "   source .env"
    echo "   或者"
    echo "   export \$(cat .env | xargs)"
else
    echo "❌ 未找到 env.example 文件"
    exit 1
fi

