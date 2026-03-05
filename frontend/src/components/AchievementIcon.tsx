import React from 'react';
import './AchievementIcon.css';

interface AchievementIconProps {
    rarity: 'common' | 'rare' | 'epic' | 'legendary';
    size?: 'small' | 'medium' | 'large';
    locked?: boolean;
    colorVariant?: number; // 用于普通称号的不同颜色变体
    titleName?: string;    // 称号名称，用于特定称号的颜色覆盖
}

// 普通称号的多种颜色变体
const commonColorVariants = [
    { color: 'common-teal', glow: 'rgba(32, 178, 170, 0.7)' },      // 青绿色
    { color: 'common-coral', glow: 'rgba(255, 107, 107, 0.7)' },    // 珊瑩红/粉色
    { color: 'common-mint', glow: 'rgba(78, 205, 196, 0.7)' },      // 薄荷绿
    { color: 'common-brown', glow: 'rgba(139, 115, 85, 0.7)' },     // 暖棕色
    { color: 'common-olive', glow: 'rgba(124, 179, 66, 0.7)' },     // 橄榄绿
    { color: 'common-cyan', glow: 'rgba(67, 233, 123, 0.7)' },      // 青绿色
];

// 特定称号名称的颜色映射（覆盖默认的 colorVariant 计算）
const titleNameColorMap: Record<string, { color: string, glow: string }> = {
    '初心': { color: 'common-coral', glow: 'rgba(255, 107, 107, 0.7)' },  // 粉色
    '笔吏': { color: 'common-cyan', glow: 'rgba(67, 233, 123, 0.7)' },    // 青绿色
};

// 根据稀有度、颜色变体和称号名称获取颜色配置
const getRarityConfig = (rarity: string, colorVariant?: number, titleName?: string) => {
    // 优先检查特定称号名称的颜色映射
    if (titleName && titleNameColorMap[titleName]) {
        const mapped = titleNameColorMap[titleName];
        return {
            name: '普通',
            color: mapped.color,
            glow: mapped.glow
        };
    }
    switch (rarity) {
        case 'legendary':
            return {
                name: '传说',
                color: 'legendary',
                glow: 'rgba(255, 107, 0, 0.8)'
            };
        case 'epic':
            return {
                name: '史诗',
                color: 'epic',
                glow: 'rgba(163, 53, 238, 0.8)'
            };
        case 'rare':
            return {
                name: '稀有',
                color: 'rare',
                glow: 'rgba(0, 112, 221, 0.8)'
            };
        default: {
            // 普通称号使用不同的颜色变体
            const variant = commonColorVariants[(colorVariant || 0) % commonColorVariants.length];
            return {
                name: '普通',
                color: variant.color,
                glow: variant.glow
            };
        }
    }
};

export const AchievementIcon: React.FC<AchievementIconProps> = ({ 
    rarity, 
    size = 'medium',
    locked = false,
    colorVariant = 0,
    titleName
}) => {
    const config = getRarityConfig(rarity, colorVariant, titleName);
    
    return (
        <div className={`achievement-icon achievement-icon-${size}`}>
            <div className={`icon-wrapper ${locked ? 'locked' : `icon-${config.color}`}`}>
                {locked ? (
                    <span className="lock-icon">🔒</span>
                ) : (
                    <>
                        <img 
                            src="/chengjiu.png" 
                            alt={config.name}
                            className="achievement-image"
                        />
                        <div 
                            className="icon-glow" 
                            style={{ boxShadow: `0 0 20px ${config.glow}, 0 0 40px ${config.glow}` }}
                        ></div>
                    </>
                )}
            </div>
        </div>
    );
};
