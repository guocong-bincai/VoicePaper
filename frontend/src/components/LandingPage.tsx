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
            <a
              href="https://github.com/guocong-bincai/VoicePaper"
              target="_blank"
              rel="noopener noreferrer"
              className="nav-btn github-btn"
              title="GitHub"
            >
              <svg viewBox="0 0 16 16" width="20" height="20" fill="currentColor">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
              </svg>
            </a>
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

