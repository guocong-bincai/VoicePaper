import React from 'react';
import { useStore } from '../store/useStore';
import './LandingPage.css';

interface LandingPageProps {
  onShowLogin: () => void;
  onShowRegister: () => void;
}

export const LandingPage: React.FC<LandingPageProps> = ({ onShowLogin, onShowRegister }) => {
  const { theme, setTheme } = useStore();

  const handleThemeToggle = () => {
    const newTheme = theme === 'light' ? 'dark' : 'light';
    setTheme(newTheme);
  };

  return (
    <div className="landing-page">
      {/* 背景装饰 */}
      <div className="landing-background">
        <div className="bg-gradient"></div>
        <div className="bg-pattern"></div>
      </div>

      {/* 顶部导航栏 */}
      <nav className="landing-nav">
        <div className="landing-nav-content">
          <button 
            className="landing-logo landing-logo-btn" 
            onClick={handleThemeToggle}
            title={theme === 'light' ? '切换到深色模式' : '切换到浅色模式'}
          >
            VoicePaper<span className="dot">.</span>
          </button>
          <div className="landing-nav-buttons">
            <button className="nav-btn login-btn" onClick={onShowLogin}>
              登录
            </button>
            <button className="nav-btn register-btn" onClick={onShowRegister}>
              注册
            </button>
          </div>
        </div>
      </nav>

      {/* 主内容区 */}
      <div className="landing-content">
        <div className="landing-hero">
          {/* 主Slogan */}
          <h1 className="landing-slogan-cn">
            以深度，见高度
          </h1>
          <div className="landing-divider"></div>
          <h2 className="landing-slogan-en">
            Depth Defines Height
          </h2>
          <p className="landing-subtitle">
            VoicePaper: Redefine How You Engage with English Articles
          </p>
        </div>

        {/* 装饰元素 */}
        <div className="landing-decoration">
          <div className="decoration-circle circle-1"></div>
          <div className="decoration-circle circle-2"></div>
          <div className="decoration-circle circle-3"></div>
        </div>
      </div>

      {/* 底部信息 */}
      <footer className="landing-footer">
        <p>让每一次阅读，都成为一次深度对话</p>
      </footer>
    </div>
  );
};

