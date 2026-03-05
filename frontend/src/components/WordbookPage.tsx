import React, { useState, useEffect } from 'react';
import { getWordbooks } from '../services/api';
import type { WordbookInfo } from '../services/api';
import { WordbookReader } from './WordbookReader';
import { LoadingPage } from './LoadingPage';
import { FloatingBackButton } from './FloatingBackButton';
import './WordbookPage.css';

// 简单的内存缓存，避免重复加载显示loading
let wordbooksCache: WordbookInfo[] | null = null;

interface WordbookPageProps {
    onClose: () => void;
}

// 默认封面映射（当数据库 cover_url 为空时使用）
const DEFAULT_COVERS: Record<string, string> = {
    'junior': '/icon/dancishu/junior.png',
    'senior': '/icon/dancishu/senior.png',
    'cet4': '/icon/dancishu/cet4.png',
    'cet6': '/icon/dancishu/cet6.png',
    'postgrad': '/icon/dancishu/postgrad.png',
    'toefl': '/icon/dancishu/toefl.png',
    'gre': '/icon/dancishu/gre.png',
    'ielts': '/icon/dancishu/ielts.png',
};

export const WordbookPage: React.FC<WordbookPageProps> = ({ onClose }) => {
    const [wordbooks, setWordbooks] = useState<WordbookInfo[]>(wordbooksCache || []);
    // 如果有缓存，就不显示loading
    const [loading, setLoading] = useState(!wordbooksCache);
    const [selectedBook, setSelectedBook] = useState<WordbookInfo | null>(() => {
        const saved = localStorage.getItem('selectedWordbook');
        return saved ? JSON.parse(saved) : null;
    });

    useEffect(() => {
        console.log('📚 WordbookPage mounted');
        loadWordbooks();
    }, []);

    // 获取封面图片的辅助函数
    const getCoverUrl = (book: WordbookInfo) => {
        // 调试日志：查看每本书的封面逻辑
        if (book.type === 'cet4') {
            console.log('🔍 CET4 Cover Debug:', {
                db_url: book.cover_url,
                default_map: DEFAULT_COVERS[book.type],
                icon: book.icon
            });
        }

        // 1. 优先使用数据库中的 cover_url
        if (book.cover_url && book.cover_url.trim() !== '') {
            return book.cover_url;
        }
        // 2. 其次使用类型映射的默认封面
        if (DEFAULT_COVERS[book.type]) {
            return DEFAULT_COVERS[book.type];
        }
        // 3. 最后使用 icon 或通用默认图
        return book.icon || '/icon/dancishu/cet4.png';
    };

    // 保存选中的单词书到 localStorage
    useEffect(() => {
        if (selectedBook) {
            localStorage.setItem('selectedWordbook', JSON.stringify(selectedBook));
            // 推送历史记录，以便后退键能返回列表
            if (window.history.state?.depth !== 2) {
                window.history.pushState({ isSubPage: true, depth: 2 }, '');
            }
        } else {
            localStorage.removeItem('selectedWordbook');
        }
    }, [selectedBook]);

    // 监听后退键返回列表
    useEffect(() => {
        const handlePopState = (event: PopStateEvent) => {
            if (event.state?.depth < 2) {
                setSelectedBook(null);
            }
        };
        window.addEventListener('popstate', handlePopState);
        return () => window.removeEventListener('popstate', handlePopState);
    }, []);

    const loadWordbooks = async () => {
        try {
            // 只有没有缓存时才显示loading
            if (!wordbooksCache) {
                setLoading(true);
            }
            console.log('📚 Fetching wordbooks...');
            const data = await getWordbooks();
            console.log('📚 Wordbooks loaded:', data);
            setWordbooks(data);
            // 更新缓存
            wordbooksCache = data;
        } catch (error) {
            console.error('加载单词书失败:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleBookClick = (book: WordbookInfo) => {
        console.log('📚 Book clicked:', book);
        setSelectedBook(book);
    };

    if (loading) {
        return <LoadingPage />;
    }

    if (selectedBook) {
        return (
            <>
                <FloatingBackButton show={true} onBack={() => setSelectedBook(null)} />
                <WordbookReader
                    book={selectedBook}
                    onBack={() => setSelectedBook(null)}
                />
            </>
        );
    }

    return (
        <>
            <FloatingBackButton show={true} onBack={onClose} />
            <div className="wordbook-page">
            <div className="wordbook-content">
                <div className="wordbook-grid-modern">
                        {wordbooks.map((book, index) => (
                            <div
                                key={book.type}
                                className="wordbook-card-modern"
                                onClick={() => handleBookClick(book)}
                                style={{ '--card-index': index } as any}
                            >
                                {/* 卡片背景装饰 */}
                                <div className="card-bg-gradient"></div>

                                {/* 顶部 Badge 区 */}
                                <div className="card-badges">
                                    {book.hotness && book.hotness > 1000 && (
                                        <span className="badge badge-hot">🔥 热门</span>
                                    )}
                                    <span className="badge badge-words">{book.word_count} 词</span>
                                </div>

                                {/* 封面区 */}
                                <div className="book-cover-modern">
                                    <div className="cover-image-wrapper">
                                        <img
                                            src={getCoverUrl(book)}
                                            alt={book.name}
                                            onError={(e) => {
                                                const target = e.target as HTMLImageElement;
                                                // 防止无限循环：如果当前已经是默认图，就不再重试
                                                if (!target.src.includes('/icon/dancishu/cet4.png')) {
                                                    target.src = '/icon/dancishu/cet4.png';
                                                }
                                            }}
                                        />
                                    </div>
                                </div>

                                {/* 信息区 */}
                                <div className="card-info-modern">
                                    <h3 className="book-title-modern">{book.name}</h3>

                                    {/* 统计行 */}
                                    <div className="stats-modern">
                                        <div className="stat-modern">
                                            <img src="/dancishu/renshu.png" alt="人数" className="stat-icon-img" />
                                            <span className="stat-text">{(book.study_count || 0).toLocaleString()} 人学习</span>
                                        </div>
                                        <div className="stat-modern">
                                            <img src="/dancishu/redu.png" alt="热度" className="stat-icon-img" />
                                            <span className="stat-text">{(book.hotness || 0).toLocaleString()} 热度</span>
                                        </div>
                                    </div>

                                    {/* 行动按钮 */}
                                    <button className="cta-button">
                                        <span>开始学习</span>
                                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                                            <polyline points="9 18 15 12 9 6"></polyline>
                                        </svg>
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
            </div>
        </div>
        </>
    );
};
