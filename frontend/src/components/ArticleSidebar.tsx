import React, { useState, useEffect } from 'react';
import type { Article } from '../types';
import './ArticleSidebar.css';

interface ArticleSidebarProps {
  articles: Article[];
  currentArticleId?: number;
  onSelectArticle: (id: number) => void;
  onStartDictation?: () => void;
}

// 格式化日期显示
const formatDate = (dateStr?: string): string => {
  if (!dateStr) return '';
  try {
    const date = new Date(dateStr);
    const month = date.getMonth() + 1;
    const day = date.getDate();
    return `${month}月${day}日`;
  } catch {
    return '';
  }
};

// 判断是否是新文章（3天内）
const isNewArticle = (dateStr?: string): boolean => {
  if (!dateStr) return false;
  try {
    const articleDate = new Date(dateStr);
    const now = new Date();
    const diffDays = Math.floor((now.getTime() - articleDate.getTime()) / (1000 * 60 * 60 * 24));
    return diffDays <= 3;
  } catch {
    return false;
  }
};

export const ArticleSidebar: React.FC<ArticleSidebarProps> = ({
  articles,
  currentArticleId,
  onSelectArticle,
  onStartDictation,
}) => {
  const [isOpen, setIsOpen] = useState(false);

  // 调试：检查文章数据中是否有 pic_url
  useEffect(() => {
    if (articles.length > 0) {
      console.log('=== ArticleSidebar 文章数据调试 ===');
      console.log('文章总数:', articles.length);
      console.log('第一篇文章 pic_url:', articles[0].pic_url);
      const articlesWithPic = articles.filter(a => a.pic_url);
      console.log('有封面图的文章数:', articlesWithPic.length);
      if (articlesWithPic.length > 0) {
        console.log('示例 pic_url:', articlesWithPic[0].pic_url);
      }
    }
  }, [articles]);

  // 点击外部区域关闭侧边栏
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      if (
        isOpen &&
        !target.closest('.article-sidebar') &&
        !target.closest('.sidebar-toggle-button')
      ) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      // 防止背景滚动
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  const handleSelectArticle = (id: number) => {
    onSelectArticle(id);
    setIsOpen(false);
  };

  // 获取文章状态
  const getArticleStatus = (article: Article): 'new' | 'reading' | null => {
    if (currentArticleId === article.id) {
      return 'reading';
    }
    // 只有3天内的文章才显示NEW
    if (isNewArticle(article.publish_date || article.created_at)) {
      return 'new';
    }
    return null;
  };

  return (
    <>
      {/* 浮动按钮 */}
      <button
        className="sidebar-toggle-button"
        onClick={() => setIsOpen(!isOpen)}
        aria-label={isOpen ? "关闭文章列表" : "打开文章列表"}
      >
        <svg
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          {isOpen ? (
            // 左箭头（关闭）
            <>
              <line x1="19" y1="12" x2="5" y2="12"></line>
              <polyline points="12 19 5 12 12 5"></polyline>
            </>
          ) : (
            // 列表图标
            <>
              <line x1="8" y1="6" x2="21" y2="6"></line>
              <line x1="8" y1="12" x2="21" y2="12"></line>
              <line x1="8" y1="18" x2="21" y2="18"></line>
              <line x1="3" y1="6" x2="3.01" y2="6"></line>
              <line x1="3" y1="12" x2="3.01" y2="12"></line>
              <line x1="3" y1="18" x2="3.01" y2="18"></line>
            </>
          )}
        </svg>
        <span className="sidebar-button-text">
          {isOpen ? '收起' : '文章'}
        </span>
      </button>

      {/* 遮罩层 */}
      {isOpen && (
        <div
          className="sidebar-overlay"
          onClick={() => setIsOpen(false)}
        />
      )}

      {/* 侧边栏 */}
      <aside className={`article-sidebar ${isOpen ? 'open' : ''}`}>
        <div className="sidebar-header">
          <h2 className="sidebar-title">文章列表</h2>
          <button
            className="sidebar-close-button"
            onClick={() => setIsOpen(false)}
            aria-label="关闭侧边栏"
          >
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>

        <div className="sidebar-content">
          {articles.length === 0 ? (
            <div className="sidebar-empty">
              <svg className="sidebar-empty-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                <path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <p>暂无文章</p>
            </div>
          ) : (
            <ul className="article-list">
              {articles.map((article, index) => {
                const status = getArticleStatus(article);
                const dateStr = formatDate(article.publish_date || article.created_at);
                const sentenceCount = article.sentences?.length || 0;
                const wordCount = article.words?.length || 0;
                
                return (
                <li
                  key={article.id}
                  className={`article-item ${
                    currentArticleId === article.id ? 'active' : ''
                  }`}
                  onClick={() => handleSelectArticle(article.id)}
                >
                    {/* 封面图 */}
                    {article.pic_url && (
                      <div className="article-item-cover">
                        <img 
                          src={article.pic_url} 
                          alt={article.title}
                          loading="lazy"
                          onError={(e) => {
                            // 图片加载失败时隐藏封面区域
                            (e.target as HTMLImageElement).style.display = 'none';
                          }}
                        />
                      </div>
                    )}
                    
                    {/* 序号 */}
                    <div className="article-item-number">
                      {String(articles.length - index).padStart(2, '0')}
                    </div>
                    
                  <div className="article-item-content">
                      {/* 标题 */}
                    <h3 className="article-item-title">{article.title}</h3>
                      
                      {/* 元信息 */}
                      <div className="article-item-meta">
                        {/* 日期 */}
                        {dateStr && (
                          <span className="article-meta-item">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                              <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                              <line x1="16" y1="2" x2="16" y2="6"></line>
                              <line x1="8" y1="2" x2="8" y2="6"></line>
                              <line x1="3" y1="10" x2="21" y2="10"></line>
                            </svg>
                            {dateStr}
                          </span>
                        )}
                        
                        {/* 句子数 */}
                        {sentenceCount > 0 && (
                          <span className="article-meta-item">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                            </svg>
                            {sentenceCount}句
                          </span>
                        )}
                        
                        {/* 单词数 */}
                        {wordCount > 0 && (
                          <span className="article-meta-item">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                              <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path>
                              <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path>
                            </svg>
                            {wordCount}词
                          </span>
                        )}
                      </div>
                      
                      {/* 分类标签 */}
                    {article.category && (
                      <span className="article-item-category">
                        {article.category.name}
                      </span>
                    )}
                  </div>
                    
                    {/* 右下角状态图标 */}
                    {status && (
                      <div className={`article-status-icon ${status}`}>
                        <img 
                          src={status === 'reading' ? '/icon/yueduzhong.png' : '/icon/zuixin.png'} 
                          alt={status === 'reading' ? '阅读中' : '最新'}
                        />
                    </div>
                  )}
                </li>
                );
              })}
            </ul>
          )}
        </div>

        <div className="sidebar-footer">
          <p className="sidebar-footer-text">
            双语精读 · 第 {articles.length} 期
          </p>
        </div>

        {/* 默写练习按钮 */}
        {onStartDictation && (
          <div className="sidebar-dictation-entry">
            <button
              className="dictation-sidebar-btn"
              onClick={() => {
                onStartDictation();
                setIsOpen(false);
              }}
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                <path d="M2 17l10 5 10-5M2 12l10 5 10-5"/>
              </svg>
              <span>默写练习</span>
            </button>
          </div>
        )}
      </aside>
    </>
  );
};
