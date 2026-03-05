import React, { useState, useEffect, useCallback } from 'react';
import {
    getVocabularyList,
    getVocabularyStats,
    getTodayReviewList,
    deleteVocabulary,
    toggleVocabularyStar,
    submitVocabularyReview,
    type Vocabulary,
    type VocabularyStats
} from '../services/api';
import ReviewSession from './ReviewSession';
import { FloatingBackButton } from './FloatingBackButton';
import './VocabularyPage.css';

// 简单的内存缓存
let vocabularyCache = {
    stats: null as VocabularyStats | null,
    defaultList: null as { data: Vocabulary[], total: number } | null
};

interface VocabularyPageProps {
    onClose: () => void;
}

type TabType = 'all' | 'review' | 'starred' | 'mastered';

export const VocabularyPage: React.FC<VocabularyPageProps> = ({ onClose }) => {
    const [activeTab, setActiveTab] = useState<TabType>('all');
    const [vocabularyList, setVocabularyList] = useState<Vocabulary[]>(vocabularyCache.defaultList?.data || []);
    const [stats, setStats] = useState<VocabularyStats | null>(vocabularyCache.stats);
    // 如果有默认列表缓存，就不显示初始loading
    const [loading, setLoading] = useState(!vocabularyCache.defaultList);
    const [keyword, setKeyword] = useState('');
    const [page, setPage] = useState(1);
    const [total, setTotal] = useState(vocabularyCache.defaultList?.total || 0);
    const PAGE_SIZE = 20;

    // 复习模式状态
    const [showReviewSession, setShowReviewSession] = useState(false);
    const [reviewMode, setReviewMode] = useState(false);
    const [reviewList, _setReviewList] = useState<Vocabulary[]>([]);
    const [currentReviewIndex, setCurrentReviewIndex] = useState(0);
    const [showAnswer, setShowAnswer] = useState(false);
    const [reviewLoading, setReviewLoading] = useState(false);

    // 防止未使用变量警告
    void _setReviewList;

    // 获取统计
    const fetchStats = useCallback(async () => {
        try {
            const data = await getVocabularyStats();
            setStats(data);
            vocabularyCache.stats = data; // 更新缓存
        } catch (error) {
            console.error('获取统计失败:', error);
        }
    }, []);

    // 获取列表
    const fetchList = useCallback(async () => {
        // 如果是默认列表且有缓存，第一次不显示loading
        const isDefaultList = activeTab === 'all' && !keyword && page === 1;
        if (!isDefaultList || !vocabularyCache.defaultList) {
            setLoading(true);
        }
        
        try {
            let params: any = {
                limit: PAGE_SIZE,
                offset: (page - 1) * PAGE_SIZE
            };

            if (keyword) {
                params.keyword = keyword;
            }

            switch (activeTab) {
                case 'starred':
                    params.is_starred = true;
                    break;
                case 'mastered':
                    params.mastery_level = 5;
                    break;
                case 'review':
                    // 待复习使用专门的接口
                    const reviewData = await getTodayReviewList(PAGE_SIZE);
                    setVocabularyList(reviewData.data || []);
                    setTotal(reviewData.count || 0);
                    setLoading(false);
                    return;
            }

            const data = await getVocabularyList(params);
            setVocabularyList(data.data || []);
            setTotal(data.total || 0);

            // 如果是默认列表，更新缓存
            if (isDefaultList) {
                vocabularyCache.defaultList = {
                    data: data.data || [],
                    total: data.total || 0
                };
            }
        } catch (error) {
            console.error('获取生词列表失败:', error);
        } finally {
            setLoading(false);
        }
    }, [activeTab, keyword, page]);

    useEffect(() => {
        fetchStats();
    }, [fetchStats]);

    useEffect(() => {
        fetchList();
    }, [fetchList]);

    // 当切换标签或搜索时，重置页码
    useEffect(() => {
        setPage(1);
    }, [activeTab, keyword]);

    // 开始复习 - 打开全屏复习会话
    const startReview = () => {
        // 请求全屏
        const elem = document.documentElement;
        if (elem.requestFullscreen) {
            elem.requestFullscreen().catch(err => {
                console.warn('全屏模式启动失败:', err);
            });
        } else if ((elem as any).webkitRequestFullscreen) {
            (elem as any).webkitRequestFullscreen();
        }

        setShowReviewSession(true);
        // 推送历史记录，以便后退键能关闭复习
        if (window.history.state?.depth !== 2) {
            window.history.pushState({ isSubPage: true, depth: 2 }, '');
        }
    };

    // 监听后退键关闭复习
    useEffect(() => {
        const handlePopState = (event: PopStateEvent) => {
            if (event.state?.depth < 2) {
                setShowReviewSession(false);
            }
        };
        window.addEventListener('popstate', handlePopState);
        return () => window.removeEventListener('popstate', handlePopState);
    }, []);

    // 复习完成后的回调
    const handleReviewComplete = () => {
        fetchStats();
        fetchList();
    };

    // 提交复习结果
    const handleReviewSubmit = async (quality: number) => {
        if (reviewLoading) return;

        const currentVocab = reviewList[currentReviewIndex];
        if (!currentVocab) return;

        setReviewLoading(true);
        try {
            await submitVocabularyReview(currentVocab.id, {
                quality,
                review_type: 'card'
            });

            // 下一个
            if (currentReviewIndex < reviewList.length - 1) {
                setCurrentReviewIndex(prev => prev + 1);
                setShowAnswer(false);
            } else {
                // 复习完成
                setReviewMode(false);
                fetchStats();
                fetchList();
                alert('🎉 今日复习完成！');
            }
        } catch (error) {
            console.error('提交复习失败:', error);
        } finally {
            setReviewLoading(false);
        }
    };

    // 删除生词
    const handleDelete = async (id: number) => {
        if (!confirm('确定要删除这个生词吗？')) return;

        try {
            await deleteVocabulary(id);
            fetchList();
            fetchStats();
        } catch (error) {
            console.error('删除失败:', error);
        }
    };

    // 切换标星
    const handleToggleStar = async (id: number) => {
        try {
            await toggleVocabularyStar(id);
            fetchList();
            fetchStats();
        } catch (error) {
            console.error('操作失败:', error);
        }
    };

    // 获取掌握度进度条（基于复习次数，最多10次显示满）
    const getMasteryProgress = (reviewCount: number) => {
        return Math.min((reviewCount / 10) * 100, 100);
    };

    // 获取掌握度文本（基于复习次数）
    const getMasteryText = (reviewCount: number) => {
        if (reviewCount === 0) return '未学习';
        if (reviewCount <= 2) return `复习${reviewCount}次`;
        if (reviewCount <= 5) return `复习${reviewCount}次`;
        if (reviewCount <= 10) return `复习${reviewCount}次`;
        return `复习${reviewCount}次`;
    };

    // 获取进度条颜色类
    const getMasteryColorClass = (reviewCount: number) => {
        if (reviewCount === 0) return 'level-0';
        if (reviewCount <= 2) return 'level-1';
        if (reviewCount <= 5) return 'level-2';
        if (reviewCount <= 10) return 'level-3';
        return 'level-4';
    };

    // 复习模式渲染
    if (reviewMode && reviewList.length > 0) {
        const currentVocab = reviewList[currentReviewIndex];

        return (
            <>
                <FloatingBackButton show={true} onBack={() => setReviewMode(false)} />
                <div className="vocabulary-page">
                    <div className="vocabulary-header">
                        <button className="back-btn" onClick={() => setReviewMode(false)}>
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <path d="M19 12H5M12 19l-7-7 7-7"/>
                        </svg>
                    </button>
                    <h1>复习模式</h1>
                    <div className="review-progress-text">
                        {currentReviewIndex + 1} / {reviewList.length}
                    </div>
                </div>

                <div className="review-card-container">
                    <div className={`review-card ${showAnswer ? 'flipped' : ''}`}>
                        <div className="card-front">
                            <div className="card-type">{currentVocab.type === 'word' ? '单词' : currentVocab.type === 'phrase' ? '短语' : '句子'}</div>
                            <div className="card-content">{currentVocab.content}</div>
                            {currentVocab.phonetic && (
                                <div className="card-phonetic">{currentVocab.phonetic}</div>
                            )}
                            <button className="show-answer-btn" onClick={() => setShowAnswer(true)}>
                                点击显示释义
                            </button>
                        </div>

                        {showAnswer && (
                            <div className="card-back">
                                <div className="card-meaning">{currentVocab.meaning || '暂无释义'}</div>
                                {currentVocab.example && (
                                    <div className="card-example">
                                        <span className="example-label">例句：</span>
                                        {currentVocab.example}
                                    </div>
                                )}
                                {currentVocab.context && (
                                    <div className="card-context">
                                        <span className="context-label">上下文：</span>
                                        {currentVocab.context}
                                    </div>
                                )}
                            </div>
                        )}
                    </div>

                    {showAnswer && (
                        <div className="review-actions">
                            <button
                                className="review-btn forget"
                                onClick={() => handleReviewSubmit(0)}
                                disabled={reviewLoading}
                            >
                                <span className="btn-emoji">😰</span>
                                <span className="btn-text">忘了</span>
                            </button>
                            <button
                                className="review-btn hard"
                                onClick={() => handleReviewSubmit(2)}
                                disabled={reviewLoading}
                            >
                                <span className="btn-emoji">😕</span>
                                <span className="btn-text">困难</span>
                            </button>
                            <button
                                className="review-btn normal"
                                onClick={() => handleReviewSubmit(3)}
                                disabled={reviewLoading}
                            >
                                <span className="btn-emoji">🤔</span>
                                <span className="btn-text">一般</span>
                            </button>
                            <button
                                className="review-btn easy"
                                onClick={() => handleReviewSubmit(4)}
                                disabled={reviewLoading}
                            >
                                <span className="btn-emoji">😊</span>
                                <span className="btn-text">简单</span>
                            </button>
                            <button
                                className="review-btn perfect"
                                onClick={() => handleReviewSubmit(5)}
                                disabled={reviewLoading}
                            >
                                <span className="btn-emoji">😎</span>
                                <span className="btn-text">太简单</span>
                            </button>
                        </div>
                    )}
                </div>

                {/* 进度条 */}
                <div className="review-progress-bar">
                    <div
                        className="progress-fill"
                        style={{ width: `${((currentReviewIndex + 1) / reviewList.length) * 100}%` }}
                    />
                </div>
            </div>
            </>
        );
    }

    return (
        <>
            <FloatingBackButton show={true} onBack={onClose} />
            <div className="vocabulary-page">
            {/* 统计卡片 */}
            {stats && (
                <div className="stats-cards">
                    <div className="stat-card total">
                        <div className="stat-value">{stats.total}</div>
                        <div className="stat-label">总计</div>
                    </div>
                    <div className="stat-card review">
                        <div className="stat-value">{stats.today_review}</div>
                        <div className="stat-label">待复习</div>
                    </div>
                    <div className="stat-card learning">
                        <div className="stat-value">{stats.learning + stats.reviewing}</div>
                        <div className="stat-label">学习中</div>
                    </div>
                    <div className="stat-card mastered">
                        <div className="stat-value">{stats.mastered}</div>
                        <div className="stat-label">已掌握</div>
                    </div>
                </div>
            )}

            {/* 标签页 */}
            <div className="vocabulary-tabs">
                <div className="tabs-left">
                    <button
                        className={`tab-btn ${activeTab === 'all' ? 'active' : ''}`}
                        onClick={() => setActiveTab('all')}
                    >
                        全部
                    </button>
                    <button
                        className={`tab-btn ${activeTab === 'review' ? 'active' : ''}`}
                        onClick={() => setActiveTab('review')}
                    >
                        待复习
                    </button>
                    <button
                        className={`tab-btn ${activeTab === 'starred' ? 'active' : ''}`}
                        onClick={() => setActiveTab('starred')}
                    >
                        <img src="/icon/shoucang.png" alt="" className="tab-icon" /> 收藏
                    </button>
                    <button
                        className={`tab-btn ${activeTab === 'mastered' ? 'active' : ''}`}
                        onClick={() => setActiveTab('mastered')}
                    >
                        已掌握
                    </button>
                </div>
                <button className="start-review-btn" onClick={startReview}>
                    开始复习
                </button>
            </div>

            {/* 搜索框 */}
            <div className="search-box">
                <input
                    type="text"
                    placeholder="搜索生词..."
                    value={keyword}
                    onChange={(e) => setKeyword(e.target.value)}
                />
                <svg className="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <circle cx="11" cy="11" r="8"/>
                    <path d="M21 21l-4.35-4.35"/>
                </svg>
            </div>

            {/* 生词列表 */}
            <div className="vocabulary-list">
                {loading ? (
                    <div className="loading-state">加载中...</div>
                ) : vocabularyList.length === 0 ? (
                    <div className="empty-state">
                        <div className="empty-icon">📖</div>
                        <div className="empty-text">暂无生词</div>
                        <div className="empty-hint">在阅读文章时选中单词添加到生词本</div>
                    </div>
                ) : (
                    <>
                        {vocabularyList.map(vocab => (
                            <div key={vocab.id} className="vocabulary-item">
                                <div className="vocab-main">
                                    <div className="vocab-content">
                                        <span className="vocab-text">{vocab.content}</span>
                                        {vocab.phonetic && (
                                            <span className="vocab-phonetic">{vocab.phonetic}</span>
                                        )}
                                    </div>
                                    <div className="vocab-meaning">{vocab.meaning || '暂无释义'}</div>
                                    <div className="vocab-meta">
                                        <span className={`vocab-type type-${vocab.type}`}>
                                            {vocab.type === 'word' ? '单词' : vocab.type === 'phrase' ? '短语' : '句子'}
                                        </span>
                                        <span className="vocab-mastery">
                                            <span className={`mastery-bar ${getMasteryColorClass(vocab.review_count || 0)}`}>
                                                <span className="mastery-fill" style={{ width: `${getMasteryProgress(vocab.review_count || 0)}%` }} />
                                            </span>
                                            <span className="mastery-text">{getMasteryText(vocab.review_count || 0)}</span>
                                        </span>
                                    </div>
                                </div>
                                <div className="vocab-actions">
                                    <button
                                        className={`star-btn ${vocab.is_starred ? 'starred' : ''}`}
                                        onClick={() => handleToggleStar(vocab.id)}
                                        title={vocab.is_starred ? '取消收藏' : '收藏'}
                                    >
                                        <img
                                            src={vocab.is_starred ? '/icon/yishoucang.png' : '/icon/shoucang.png'}
                                            alt={vocab.is_starred ? '已收藏' : '收藏'}
                                            className="action-icon"
                                        />
                                    </button>
                                    <button
                                        className="delete-btn"
                                        onClick={() => handleDelete(vocab.id)}
                                        title="删除"
                                    >
                                        <img src="/icon/shanchu.png" alt="删除" className="action-icon" />
                                    </button>
                                </div>
                            </div>
                        ))}

                        {/* 分页控制 */}
                        {total > PAGE_SIZE && (
                            <div className="pagination-controls">
                                <button
                                    className="page-btn"
                                    disabled={page === 1}
                                    onClick={() => {
                                        setPage(prev => prev - 1);
                                        document.querySelector('.vocabulary-page')?.scrollTo(0, 0);
                                    }}
                                >
                                    上一页
                                </button>
                                <span className="page-info">
                                    第 {page} / {Math.ceil(total / PAGE_SIZE)} 页 (共 {total} 条)
                                </span>
                                <button
                                    className="page-btn"
                                    disabled={page >= Math.ceil(total / PAGE_SIZE)}
                                    onClick={() => {
                                        setPage(prev => prev + 1);
                                        document.querySelector('.vocabulary-page')?.scrollTo(0, 0);
                                    }}
                                >
                                    下一页
                                </button>
                            </div>
                        )}
                    </>
                )}
            </div>

            {/* 全屏复习会话 */}
            {showReviewSession && (
                <ReviewSession
                    onClose={() => {
                        // 尝试退出全屏
                        if (document.fullscreenElement) {
                            document.exitFullscreen().catch(() => {});
                        } else if ((document as any).webkitFullscreenElement) {
                            (document as any).webkitExitFullscreen();
                        }

                        setShowReviewSession(false);
                        handleReviewComplete();
                    }}
                    onComplete={handleReviewComplete}
                />
            )}
        </div>
        </>
    );
};

export default VocabularyPage;
