import React from 'react';
import './LevelBadge.css';

interface LevelBadgeProps {
    level: number;
    size?: 'small' | 'medium' | 'large';
    showLabel?: boolean;
    displayName?: string;   // 穿戴的称号名称
    displayRarity?: string; // 穿戴称号的稀有度
    titleId?: number;       // 穿戴称号的ID，用于普通称号颜色变体
}

// 根据等级获取徽章样式配置
const getLevelConfig = (level: number) => {
    if (level >= 30) {
        return {
            name: '传奇大师',
            color: 'legendary',
            glowColor: '#FFD700',
            glowIntensity: 'strong'
        };
    } else if (level >= 20) {
        return {
            name: '资深专家',
            color: 'master',
            glowColor: '#8A2BE2',
            glowIntensity: 'strong'
        };
    } else if (level >= 15) {
        return {
            name: '高级学者',
            color: 'expert',
            glowColor: '#FFA500',
            glowIntensity: 'medium'
        };
    } else if (level >= 10) {
        return {
            name: '进阶学者',
            color: 'advanced',
            glowColor: '#C0C0C0',
            glowIntensity: 'medium'
        };
    } else if (level >= 5) {
        return {
            name: '努力学习',
            color: 'intermediate',
            glowColor: '#CD7F32',
            glowIntensity: 'light'
        };
    } else {
        return {
            name: '初学者',
            color: 'beginner',
            glowColor: '#667eea',
            glowIntensity: 'light'
        };
    }
};

// 普通称号的多种颜色变体
const commonColorVariants = [
    '#20b2aa',  // 青绿色 teal
    '#ff6b6b',  // 珊瑩红/粉色 coral
    '#4ecdc4',  // 薄荷绿 mint
    '#8b7355',  // 暖棕色 brown
    '#7cb342',  // 橄榄绿 olive
    '#43e97b',  // 青绿色 cyan
];

// 特定称号名称的颜色映射（覆盖默认的 titleId 计算）
const titleNameColorMap: Record<string, string> = {
    '初心': '#ff6b6b',  // 粉色
    '笔吏': '#43e97b',  // 青绿色
};

// 根据稀有度、称号ID和名称获取光晕颜色
const getRarityGlowColor = (rarity?: string, titleId?: number, titleName?: string) => {
    // 优先检查特定称号名称的颜色映射
    if (titleName && titleNameColorMap[titleName]) {
        return titleNameColorMap[titleName];
    }
    
    switch (rarity) {
        case 'legendary': return '#FFD700';  // 金色
        case 'epic': return '#a335ee';       // 紫色
        case 'rare': return '#0070dd';       // 蓝色
        case 'common': {
            // 普通称号根据 titleId 使用不同的颜色
            const variantIndex = (titleId || 0) % commonColorVariants.length;
            return commonColorVariants[variantIndex];
        }
        default: return '';                  // 返回空，使用默认等级光晕
    }
};

export const LevelBadge: React.FC<LevelBadgeProps> = ({ 
    level, 
    size = 'medium',
    showLabel = false,
    displayName,
    displayRarity,
    titleId
}) => {
    const config = getLevelConfig(level);
    // 如果有穿戴称号稀有度，使用对应的光晕颜色；否则使用等级对应的光晕颜色
    const rarityGlow = getRarityGlowColor(displayRarity, titleId, displayName);
    const glowColor = rarityGlow || config.glowColor;
    // 显示的称号名称：优先使用穿戴的称号名称
    const badgeName = displayName || config.name;
    
    return (
        <div className={`level-badge level-badge-${size}`}>
            <div className={`badge-icon badge-${config.color} glow-${config.glowIntensity}`}>
                {/* 光晕层 */}
                <div 
                    className="badge-glow-layer"
                    style={{
                        boxShadow: `
                            0 0 20px ${glowColor},
                            0 0 40px ${glowColor},
                            0 0 60px ${glowColor},
                            0 0 80px ${glowColor}
                        `
                    }}
                />
                {/* 图标 */}
                <img 
                    src="/chengjiu.png" 
                    alt={badgeName}
                    className="badge-image"
                    style={{
                        filter: `drop-shadow(0 0 15px ${glowColor})`
                    }}
                />
            </div>
            {showLabel && (
                <div className="badge-label">
                    <span className="badge-level">Lv.{level}</span>
                    <span className="badge-name">{badgeName}</span>
                </div>
            )}
        </div>
    );
};
