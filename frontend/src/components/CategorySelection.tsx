import React, { useEffect, useState } from 'react';
import { getCategories } from '../services/api';
import type { Category } from '../types';
import { FloatingBackButton } from './FloatingBackButton';
import './CategorySelection.css';

// 简单的内存缓存，避免返回时显示加载中
let categoriesCache: Category[] | null = null;

interface CategorySelectionProps {
    onSelectCategory: (categoryId: number, categoryName?: string) => void;
    onBack?: () => void;
}

export const CategorySelection: React.FC<CategorySelectionProps> = ({ onSelectCategory, onBack }) => {
    const [categories, setCategories] = useState<Category[]>(categoriesCache || []);
    // 如果有缓存，就不显示loading
    const [loading, setLoading] = useState(!categoriesCache);
    const [error, setError] = useState('');

    useEffect(() => {
        loadCategories();
    }, []);

    const loadCategories = async () => {
        try {
            // 只有没有缓存时才显示loading
            if (!categoriesCache) {
                setLoading(true);
            }
            const data = await getCategories();
            setCategories(data);
            // 更新缓存
            categoriesCache = data;
        } catch (err: any) {
            console.error('❌ 加载分类失败:', err);
            // 只有在没有缓存数据的情况下才显示错误
            if (!categoriesCache) {
                setError(err.response?.data?.error || '加载失败，请稍后重试');
            }
        } finally {
            setLoading(false);
        }
    };

    // 默认图标映射（如果数据库没有图标）
    const getDefaultIcon = (name: string): string => {
        const iconMap: { [key: string]: string } = {
            '科技': '💻',
            '商业': '💼',
            '文化': '🎭',
            '体育': '⚽',
            '娱乐': '🎬',
            '健康': '🏥',
            '教育': '📚',
            '旅游': '✈️',
            '美食': '🍔',
            '时尚': '👗',
            '艺术': '🎨',
            '音乐': '🎵',
            '默认': '📰',
        };
        return iconMap[name] || iconMap['默认'];
    };

    // 获取随机渐变背景类名
    const getGradientClass = (index: number) => {
        const gradients = [
            'visual-gradient-1',
            'visual-gradient-2',
            'visual-gradient-3',
            'visual-gradient-4'
        ];
        return gradients[index % gradients.length];
    };

    // 判断是否为图片URL
    const isImageUrl = (str?: string) => {
        if (!str) return false;
        return str.match(/\.(jpeg|jpg|gif|png|webp|svg)$/i) || str.startsWith('http');
    };

    if (loading) {
        return (
            <div className="category-loading">
                <div className="loading-spinner"></div>
                <p>正在加载合集...</p>
            </div>
        );
    }

    if (error) {
        return (
            <div className="category-error">
                <div className="error-icon">⚠️</div>
                <p>{error}</p>
                <button className="retry-btn" onClick={loadCategories}>
                    重试
                </button>
            </div>
        );
    }

    return (
        <>
            {onBack && <FloatingBackButton show={true} onBack={onBack} />}
            <div className="category-selection">
                <div className="category-selection-header">
                    <h1 className="category-selection-title">句乐部</h1>
                    <p className="category-selection-subtitle">在真实场景中，让每一句都充满快乐</p>
                </div>
                <div className="categories-grid">
                    {categories.map((category, index) => (
                        <div
                            key={category.id}
                            className="category-card"
                            onClick={() => onSelectCategory(category.id, category.name)}
                        >
                        {/* 顶部视觉区域 - 智能判断显示图片还是渐变图标 */}
                        <div className={`category-card-visual ${!isImageUrl(category.icon) ? getGradientClass(index) : ''}`}>
                            {isImageUrl(category.icon) ? (
                                <img
                                    src={category.icon}
                                    alt={category.name}
                                    className="category-card-image"
                                    onError={(e) => {
                                        // 图片加载失败回退到默认样式
                                        e.currentTarget.style.display = 'none';
                                        e.currentTarget.parentElement?.classList.add(getGradientClass(index));
                                        // 创建一个图标元素并插入
                                        const iconWrapper = document.createElement('div');
                                        iconWrapper.className = 'category-icon-wrapper';
                                        iconWrapper.textContent = getDefaultIcon(category.name);
                                        e.currentTarget.parentElement?.appendChild(iconWrapper);
                                    }}
                                />
                            ) : (
                                <div className="category-icon-wrapper">
                                    {category.icon || getDefaultIcon(category.name)}
                                </div>
                            )}
                        </div>

                        {/* 内容区域 */}
                        <div className="category-card-content">
                            <div className="category-tags">
                                <span className="category-tag">精选合集</span>
                                {category.article_count > 10 && <span className="category-tag">热门</span>}
                            </div>

                            <h3 className="category-name">{category.name}</h3>

                            <p className="category-description">
                                {category.description || '暂无描述'}
                            </p>

                            {/* 底部信息栏 */}
                            <div className="category-footer">
                                <div className="category-stats">
                                    <div className="stat-item">
                                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                            <path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/>
                                        </svg>
                                        {category.article_count} 课
                                    </div>
                                    <div className="stat-item">
                                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                            <circle cx="12" cy="12" r="10"/>
                                            <polyline points="12 6 12 12 16 14"/>
                                        </svg>
                                        {Math.ceil(category.article_count * 15)} 分钟
                                    </div>
                                </div>
                                <div className="action-btn">
                                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                        <path d="M5 12h14M12 5l7 7-7 7"/>
                                    </svg>
                                </div>
                            </div>
                        </div>
                    </div>
                ))}
            </div>

            {/* 空状态 */}
            {categories.length === 0 && (
                <div className="category-empty">
                    <div className="empty-icon">📚</div>
                    <h3>暂无合集</h3>
                    <p>管理员还未创建学习合集</p>
                </div>
            )}
        </div>
        </>
    );
};