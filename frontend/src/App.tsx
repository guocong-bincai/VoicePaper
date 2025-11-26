import React, { useEffect } from 'react';
import { AudioPlayer } from './components/AudioPlayer';
import { ArticleViewer } from './components/ArticleViewer';
import { useStore } from './store/useStore';
import type { Article } from './types';

function App() {
  const { setCurrentArticle, currentArticle } = useStore();

  useEffect(() => {
    loadLocalArticle();
  }, []);

  const loadLocalArticle = async () => {
    try {
      // 直接加载本地 markdown 文件
      const response = await fetch('/article.md');
      const content = await response.text();
      
      // 创建文章对象
      const article: Article = {
        id: 1,
        title: '《卫报》双语精读课：帕特农神庙大理石之争 —— 一场跨越 200 年的文化博弈',
        content: content,
        audio_path: 'audio.mp3', // 本地音频文件路径
        status: 'completed' as const,
      };
      
      setCurrentArticle(article);
      console.log('✅ 本地文章加载成功');
    } catch (error) {
      console.error("❌ 加载本地文章失败:", error);
    }
  };

  return (
    <div>
      {/* 顶部导航 (极简) */}
      <nav className="top-nav">
        <div className="logo">VoicePaper<span className="dot">.</span></div>
        <div className="nav-info">双语精读 · 第 142 期</div>
      </nav>

      <div className="app-container">
        {/* 文章阅读区 */}
        <main className="reading-area">
          <article className="paper-content">
            {/* 封面图/装饰区域 */}
            <div className="article-header-decoration">
              <div className="category-tag">CULTURE</div>
              <div className="read-time">⏱ 17 min read</div>
            </div>

            <ArticleViewer />
          </article>
        </main>
      </div>

      {/* 底部悬浮播放器 */}
      <AudioPlayer />
    </div>
  );
}

export default App;