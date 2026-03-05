import React, { useMemo } from 'react';
import { useStore } from '../store/useStore';
import './LoadingPage.css';

interface LoadingPageProps {
  message?: string;
}

// 符合项目风格的标语列表
const SLOGANS = [
  { cn: '以深度，见高度', en: 'Depth Defines Height' },
  { cn: '让每一次阅读，都成为一次深度对话', en: 'Every Read, A Deep Conversation' },
  { cn: '聆听文字的声音', en: 'Hear the Voice of Words' },
  { cn: '双语精读，智慧启航', en: 'Bilingual Reading, Wisdom Sets Sail' },
  { cn: '在声音中遇见文字', en: 'Meet Words in Sound' },
  { cn: '阅读与聆听，双重学习', en: 'Read and Listen, Double Learning' },
  { cn: '用声音点亮阅读', en: 'Illuminate Reading with Voice' },
  { cn: '精读每一字，聆听每一句', en: 'Read Every Word, Hear Every Sentence' },
];

export const LoadingPage: React.FC<LoadingPageProps> = ({ message }) => {
  const { theme } = useStore();

  // 使用 useMemo 确保标语在组件生命周期内保持不变
  const randomSlogan = useMemo(() => {
    return SLOGANS[Math.floor(Math.random() * SLOGANS.length)];
  }, []);

  return (
    <div className={`loading-page ${theme === 'dark' ? 'dark' : ''}`}>
      {/* 背景装饰 */}
      <div className="loading-background">
        <div className="loading-gradient"></div>
        <div className="loading-pattern"></div>
      </div>

      {/* 主内容 */}
      <div className="loading-content">
        {/* Logo */}
        <div className="loading-logo">
          <span className="loading-logo-text">VoicePaper</span>
          <span className="loading-logo-dot">.</span>
        </div>

        {/* 加载动画 */}
        <div className="loading-animation">
          <div className="loading-spinner">
            <div className="spinner-ring"></div>
            <div className="spinner-ring"></div>
            <div className="spinner-ring"></div>
          </div>
        </div>

        {/* 标语 */}
        <div className="loading-slogan">
          <h1 className="loading-slogan-cn">{randomSlogan.cn}</h1>
          <div className="loading-divider"></div>
          <h2 className="loading-slogan-en">{randomSlogan.en}</h2>
        </div>

        {/* 加载提示 */}
        <div className="loading-message">
          <p className="loading-text">
            {message || '正在为您准备精彩内容...'}
          </p>
          <div className="loading-dots">
            <span></span>
            <span></span>
            <span></span>
          </div>
        </div>
      </div>

      {/* 装饰元素 */}
      <div className="loading-decoration">
        <div className="decoration-circle circle-1"></div>
        <div className="decoration-circle circle-2"></div>
        <div className="decoration-circle circle-3"></div>
      </div>
    </div>
  );
};

