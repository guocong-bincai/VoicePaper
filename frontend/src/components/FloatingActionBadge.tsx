import { useState } from 'react';
import './FloatingActionBadge.css';

interface FloatingActionBadgeProps {
  level: number;
  displayName?: string;  // 穿戴的称号名称
  displayRarity?: string; // 穿戴称号的稀有度
  titleId?: number;       // 穿戴称号的ID，用于普通称号颜色变体
  onLocateReading?: () => void;
  onOpenFeedback: () => void;
  onOpenProfile: () => void;
  showLocateButton: boolean;
}

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
    default: return '#667eea';           // 默认紫蓝色
  }
};

// 根据等级获取徽章样式配置
const getLevelConfig = (level: number) => {
  if (level >= 30) {
    return {
      name: '传奇大师',
      color: 'legendary',
      glowColor: '#FFD700',
    };
  } else if (level >= 20) {
    return {
      name: '资深专家',
      color: 'master',
      glowColor: '#8A2BE2',
    };
  } else if (level >= 15) {
    return {
      name: '高级学者',
      color: 'expert',
      glowColor: '#FFA500',
    };
  } else if (level >= 10) {
    return {
      name: '进阶学者',
      color: 'advanced',
      glowColor: '#C0C0C0',
    };
  } else if (level >= 5) {
    return {
      name: '努力学习',
      color: 'intermediate',
      glowColor: '#CD7F32',
    };
  } else {
    return {
      name: '初学者',
      color: 'beginner',
      glowColor: '#667eea',
    };
  }
};

export const FloatingActionBadge: React.FC<FloatingActionBadgeProps> = ({
  level,
  displayName,
  displayRarity,
  titleId,
  onLocateReading,
  onOpenFeedback,
  onOpenProfile,
  showLocateButton,
}) => {
  const [isHovered, setIsHovered] = useState(false);

  const config = getLevelConfig(level);

  // 中心点击事件
  const handleCenterClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (showLocateButton && onLocateReading) {
      onLocateReading();
    } else {
      onOpenProfile();
    }
  };

  // 长按打开个人中心
  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    onOpenProfile();
  };

  return (
    <div
      className={`floating-action-badge ${showLocateButton ? 'with-locate' : ''}`}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      onContextMenu={handleContextMenu}
      title={showLocateButton ? '点击定位朗读位置 | 右键打开个人中心' : '点击查看个人中心'}
    >
      {/* 环形菜单层 - 仅在显示定位按钮时展示 */}
      {showLocateButton && (
        <div className="fab-menu-container">
          {/* 反馈按钮 (左上扇形) */}
          <div
            className="fab-sector-item fab-sector-feedback"
            onClick={(e) => {
              e.stopPropagation();
              onOpenFeedback();
            }}
            title="意见反馈"
          >
            <div className="fab-sector-icon">
              <img src="/icon/fankuixinxi.png" alt="反馈" />
            </div>
          </div>
        </div>
      )}

      {/* 中心徽章 */}
      {(() => {
        // 如果有穿戴称号，使用称号稀有度对应的光晕颜色
        const glowColor = displayRarity ? getRarityGlowColor(displayRarity, titleId, displayName) : config.glowColor;
        return (
          <button
            className={`fab-badge-center ${isHovered && showLocateButton ? 'hovered' : ''} badge-${config.color}`}
            onClick={handleCenterClick}
          >
            {/* 光晕层 */}
            <div
              className="fab-glow-layer"
              style={{
                boxShadow: `
                  0 0 20px ${glowColor},
                  0 0 40px ${glowColor},
                  0 0 60px ${glowColor}
                `
              }}
            />
            {/* 图标 */}
            <img
              src="/chengjiu.png"
              alt={displayName || config.name}
              className="fab-badge-image"
              style={{
                filter: `drop-shadow(0 0 15px ${glowColor})`
              }}
            />
          </button>
        );
      })()}

      {/* 等级标签 - 位于图标下方 */}
      <div className="fab-badge-label">
        <span className="fab-badge-level">Lv.{level}</span>
        <span className="fab-badge-name">{displayName || config.name}</span>
      </div>
    </div>
  );
};
