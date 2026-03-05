import React, { useEffect, useState } from 'react';
import { getArticlesByCategory } from '../services/api';
import type { Article } from '../types';
import { FloatingBackButton } from './FloatingBackButton';
import './CategoryArticleList.css';

// 简单的内存缓存，避免重复加载显示loading
const articlesCache: Record<number, Article[]> = {};

interface CategoryArticleListProps {
    categoryId: number;
    categoryName: string;
    onSelectArticle: (articleId: number) => void;
    onBack: () => void;
}

export const CategoryArticleList: React.FC<CategoryArticleListProps> = ({
    categoryId,
    categoryName,
    onSelectArticle,
    onBack
}) => {
    const [articles, setArticles] = useState<Article[]>(articlesCache[categoryId] || []);
    // 如果有缓存，就不显示loading
    const [loading, setLoading] = useState(!articlesCache[categoryId]);
    const [error, setError] = useState('');

    useEffect(() => {
        loadArticles();
    }, [categoryId]);

    const loadArticles = async () => {
        try {
            // 只有没有缓存时才显示loading
            if (!articlesCache[categoryId]) {
                setLoading(true);
            }
            const data = await getArticlesByCategory(categoryId);
            // 按发布时间倒序排序（最新的在最上面）
            const sortedData = data.sort((a, b) => {
                const dateA = new Date(a.publish_date || a.created_at);
                const dateB = new Date(b.publish_date || b.created_at);
                return dateB.getTime() - dateA.getTime(); // 倒序
            });
            setArticles(sortedData);
            // 更新缓存
            articlesCache[categoryId] = sortedData;
        } catch (err: any) {
            console.error('❌ 加载文章失败:', err);
            // 只有在没有缓存数据的情况下才显示错误
            if (!articlesCache[categoryId]) {
                setError('加载文章列表失败，请稍后重试');
            }
        } finally {
            setLoading(false);
        }
    };

    // 格式化日期显示
    const formatDate = (dateString: string) => {
        const date = new Date(dateString);
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        return `${year}-${month}-${day}`;
    };

    if (loading) {
        return (
            <div className="article-list-loading">
                <div className="loading-spinner"></div>
                <p>正在加载文章...</p>
            </div>
        );
    }

    return (
        <>
            <FloatingBackButton show={true} onBack={onBack} />
            <div className="category-article-list">
                {/* 顶部导航栏 */}
                <div className="list-header">
                    <div className="header-placeholder"></div>
                    <h1 className="list-title">{categoryName}</h1>
                    <div className="header-placeholder"></div>
                </div>

                <div className="list-container">
                    {error ? (
                        <div className="list-error">
                            <p>{error}</p>
                            <button className="retry-btn" onClick={loadArticles}>重试</button>
                        </div>
                    ) : articles.length === 0 ? (
                        <div className="list-empty">
                            <div className="empty-icon">📝</div>
                            <h3>暂无文章</h3>
                            <p>该合集下暂时没有文章</p>
                        </div>
                    ) : (
                        <div className="articles-grid">
                            {articles.map((article, index) => (
                                <div
                                    key={article.id}
                                    className="article-card"
                                    onClick={() => onSelectArticle(article.id)}
                                >
                                    <div className="article-cover">
                                        <img 
                                            src={article.pic_url || '/logo.png'} 
                                            alt={article.title}
                                            onError={(e) => {
                                                const target = e.target as HTMLImageElement;
                                                target.src = '/logo.png';
                                            }}
                                        />
                                        <div className="article-card-number">
                                            {String(index + 1).padStart(2, '0')}
                                        </div>
                                        <div className="article-card-action">
                                            <button
                                                className="start-btn"
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    onSelectArticle(article.id);
                                                }}
                                            >
                                                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                    <path d="M5 12h14M12 5l7 7-7 7"/>
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
                                    
                                    <div className="article-card-content">
                                        <h3 className="article-title">{article.title}</h3>
                                        <div className="article-meta">
                                            <span className="meta-item">
                                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                    <circle cx="12" cy="12" r="10"/>
                                                    <polyline points="12 6 12 12 16 14"/>
                                                </svg>
                                                {formatDate(article.publish_date || article.created_at)}
                                            </span>
                                            {/* 最新的文章（第一篇）显示"今日推荐" */}
                                            {index === 0 ? (
                                                <span className="meta-tag today">今日推荐</span>
                                            ) : article.is_daily && (
                                                <span className="meta-tag">每日精选</span>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            ))}
                    </div>
                )}
            </div>
        </div>
        </>
    );
};
