import { useState } from 'react';
import { ArrowDown } from 'lucide-react';
import './SpeedDialMenu.css';

interface SpeedDialMenuProps {
  onLocateReading?: () => void;
  onOpenFeedback: () => void;
  showLocateButton: boolean; // 是否显示定位朗读点按钮
}

export const SpeedDialMenu: React.FC<SpeedDialMenuProps> = ({
  onLocateReading,
  onOpenFeedback,
  showLocateButton,
}) => {
  const [isHovered, setIsHovered] = useState(false);
  const [isFeedbackHovered, setIsFeedbackHovered] = useState(false);

  // 鼠标移动事件：检测是否在反馈扇形区域
  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const centerX = rect.left + rect.width / 2;
    const centerY = rect.top + rect.height / 2;
    
    const mouseX = e.clientX - centerX;
    const mouseY = e.clientY - centerY;
    const angle = Math.atan2(mouseY, mouseX) * (180 / Math.PI);
    
    // 检测是否在左上扇形区域
    const inFeedbackArea = angle >= -180 && angle <= -90;
    setIsFeedbackHovered(inFeedbackArea);
  };

  // 点击事件：根据点击位置判断功能
  const handleClick = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const centerX = rect.left + rect.width / 2;
    const centerY = rect.top + rect.height / 2;
    
    // 计算点击位置相对于中心的角度
    const clickX = e.clientX - centerX;
    const clickY = e.clientY - centerY;
    const angle = Math.atan2(clickY, clickX) * (180 / Math.PI);
    
    // 左上扇形区域（225度到315度，即西北方向）
    // 调整角度范围：-180到180，西北方向是-135到-45度
    if (angle >= -180 && angle <= -90) {
      // 左上扇形：反馈
      onOpenFeedback();
    } else {
      // 其他区域：定位朗读点
      if (onLocateReading) {
        onLocateReading();
      }
    }
  };

  // 如果不需要显示定位按钮，就不渲染
  if (!showLocateButton) {
    return null;
  }

  return (
    <div 
      className="radial-menu-container"
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => {
        setIsHovered(false);
        setIsFeedbackHovered(false);
      }}
      onMouseMove={handleMouseMove}
      onClick={handleClick}
    >
      {/* 扇形分区 - 反馈（左上） */}
      <div className={`sector-feedback ${isHovered ? 'show' : ''} ${isFeedbackHovered ? 'active' : ''}`}>
        <div className="sector-icon">
          <img src="/icon/fankuixinxi.png" alt="反馈" style={{ width: '18px', height: '18px' }} />
        </div>
      </div>

      {/* 中心按钮 - 定位朗读点 */}
      <div className={`radial-main-btn ${isHovered ? 'hovered' : ''}`}>
        <div className="btn-content">
          <ArrowDown />
          <span className="btn-text">定位朗读</span>
        </div>
      </div>
    </div>
  );
};
