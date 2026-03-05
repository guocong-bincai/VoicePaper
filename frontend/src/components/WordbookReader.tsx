import React, { useState, useEffect, useCallback, useRef } from 'react';
import confetti from 'canvas-confetti';
import { getWordbookWordsOrdered, switchWordbookOrderMode, updateWordbookOrderIndex, studyWordbookWord } from '../services/api';
import type { WordbookInfo, WordbookWord } from '../services/api';
import { useStore } from '../store/useStore';
import { FloatingBackButton } from './FloatingBackButton';
import './WordbookReader.css';

interface WordbookReaderProps {
    book: WordbookInfo;
    onBack: () => void;
}

type ReviewQuality = 'know' | 'fuzzy' | 'forget';

// 音效
const successSound = new Audio('/sounds/right.mp3');
successSound.volume = 0.6;

export const WordbookReader: React.FC<WordbookReaderProps> = ({ book, onBack }) => {
    const { theme, setTheme } = useStore();
    const [words, setWords] = useState<WordbookWord[]>([]);
    const [currentIndex, setCurrentIndex] = useState(0);
    const [startOffset, setStartOffset] = useState(0);
    const [loading, setLoading] = useState(true);
    const [fetchingNext, setFetchingNext] = useState(false);
    const [totalWords, setTotalWords] = useState(0);
    const [submitting, setSubmitting] = useState(false);
    const [showShortcuts, setShowShortcuts] = useState(false);
    const [vibratingButton, setVibratingButton] = useState<ReviewQuality | null>(null);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const [isRandom, setIsRandom] = useState(false);  // 顺序/乱序模式
    const [switchingMode, setSwitchingMode] = useState(false);  // 切换模式中
    const manualAudioPlayedRef = useRef(false); // 标记是否手动播放了音频，用于避免useEffect重复播放

    // 积分动画状态
    const [pointsAnimation, setPointsAnimation] = useState<{ points: number; show: boolean }>({ points: 0, show: false });

    // 进入全屏模式
    const enterFullscreen = useCallback(() => {
        setIsFullscreen(true);
        const elem = document.documentElement;
        if (elem.requestFullscreen) {
            elem.requestFullscreen();
        } else if ((elem as any).webkitRequestFullscreen) {
            (elem as any).webkitRequestFullscreen();
        } else if ((elem as any).msRequestFullscreen) {
            (elem as any).msRequestFullscreen();
        }
    }, []);

    // 退出全屏模式
    const exitFullscreen = useCallback(() => {
        setIsFullscreen(false);
        if (document.fullscreenElement) {
            document.exitFullscreen();
        } else if ((document as any).webkitExitFullscreen) {
            (document as any).webkitExitFullscreen();
        } else if ((document as any).msExitFullscreen) {
            (document as any).msExitFullscreen();
        }
    }, []);

    // 监听全屏变化
    useEffect(() => {
        const handleFullscreenChange = () => {
            setIsFullscreen(!!document.fullscreenElement);
        };
        document.addEventListener('fullscreenchange', handleFullscreenChange);
        document.addEventListener('webkitfullscreenchange', handleFullscreenChange);
        return () => {
            document.removeEventListener('fullscreenchange', handleFullscreenChange);
            document.removeEventListener('webkitfullscreenchange', handleFullscreenChange);
        };
    }, []);

    // 切换主题
    const toggleTheme = () => {
        setTheme(theme === 'light' ? 'dark' : 'light');
    };

    const PAGE_SIZE = 20;

    // 加载指定页码的单词（使用用户序列）
    const fetchPage = useCallback(async (page: number, indexInPage: number = 0) => {
        try {
            const data = await getWordbookWordsOrdered(book.type, page, PAGE_SIZE);
            if (data.words && data.words.length > 0) {
                setWords(data.words);
                setTotalWords(data.total);
                setStartOffset((page - 1) * PAGE_SIZE);
                setCurrentIndex(indexInPage);
                setIsRandom(data.is_random);  // 同步模式状态
                return data.words;
            }
            return null;
        } catch (error) {
            console.error('加载单词失败:', error);
            return null;
        }
    }, [book.type]);

    // 加载进度和初始单词
    useEffect(() => {
        const init = async () => {
            try {
                setLoading(true);
                // 使用新API，同时获取单词和进度信息
                const data = await getWordbookWordsOrdered(book.type, 1, PAGE_SIZE);
                console.log('📚 WordbookReader init data:', {
                    wordsCount: data.words?.length,
                    total: data.total,
                    current_index: data.current_index,
                    page: 1
                });

                if (data.words && data.words.length > 0) {
                    setWords(data.words);
                    setTotalWords(data.total);
                    setIsRandom(data.is_random);

                    // 根据进度计算页码和页内位置
                    const globalIndex = data.current_index || 0;
                    const page = Math.floor(globalIndex / PAGE_SIZE) + 1;
                    const indexInPage = globalIndex % PAGE_SIZE;

                    if (page > 1) {
                        // 需要加载正确的页
                        await fetchPage(page, indexInPage);
                    } else {
                        // 确保 indexInPage 在有效范围内
                        const safeIndex = Math.min(indexInPage, data.words.length - 1);
                        console.log('📚 Setting currentIndex:', { indexInPage, safeIndex, wordsLength: data.words.length });
                        setCurrentIndex(safeIndex);
                        setStartOffset(0);
                    }
                } else {
                    console.error('❌ No words returned from API');
                }
            } catch (error) {
                console.error('初始化单词书失败:', error);
            } finally {
                setLoading(false);
            }
        };
        init();
    }, [book.type, fetchPage]);

    // 保存进度（使用新的序列进度API）
    const handleSaveProgress = useCallback(async (globalIndex: number) => {
        try {
            await updateWordbookOrderIndex(book.type, globalIndex);
        } catch (error) {
            console.warn('保存进度失败:', error);
        }
    }, [book.type]);

    // 切换顺序/乱序模式
    const handleToggleMode = async () => {
        if (switchingMode) return;
        setSwitchingMode(true);
        try {
            const newIsRandom = !isRandom;
            await switchWordbookOrderMode(book.type, newIsRandom);
            setIsRandom(newIsRandom);

            // 重新加载当前页的单词
            const globalIndex = startOffset + currentIndex;
            const page = Math.floor(globalIndex / PAGE_SIZE) + 1;
            const indexInPage = globalIndex % PAGE_SIZE;
            await fetchPage(page, indexInPage);
        } catch (error) {
            console.error('切换模式失败:', error);
        } finally {
            setSwitchingMode(false);
        }
    };

    // 播放成功音效
    const playSuccessSound = useCallback(() => {
        try {
            successSound.currentTime = 0;
            successSound.play().catch(e => console.warn('播放音效失败:', e));
        } catch (e) {
            console.warn('播放音效失败:', e);
        }
    }, []);

    // 庆祝动效 - 彩色碎片版（与复习页面一致）
    const triggerCelebration = useCallback(() => {
        if (typeof confetti === 'undefined') {
            console.warn('confetti 库未加载');
            return;
        }

        try {
            const count = 150;
            const defaults = {
                origin: { y: 0.7 },
                zIndex: 999999
            };

            function fire(particleRatio: number, opts: any) {
                confetti({
                    ...defaults,
                    ...opts,
                    particleCount: Math.floor(count * particleRatio),
                    zIndex: 999999
                });
            }

            // 彩色爆炸效果
            fire(0.25, {
                spread: 26,
                startVelocity: 55,
            });
            fire(0.2, {
                spread: 60,
            });
            fire(0.35, {
                spread: 100,
                decay: 0.91,
                scalar: 0.8
            });
            fire(0.1, {
                spread: 120,
                startVelocity: 25,
                decay: 0.92,
                scalar: 1.2
            });
            fire(0.1, {
                spread: 120,
                startVelocity: 45,
            });

            // 持续喷射 1.5 秒
            const duration = 1500;
            const animationEnd = Date.now() + duration;

            (function frame() {
                confetti({
                    particleCount: 5,
                    angle: 60,
                    spread: 55,
                    origin: { x: 0, y: 0.8 },
                    zIndex: 999999,
                    colors: ['#10b981', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']
                });
                confetti({
                    particleCount: 5,
                    angle: 120,
                    spread: 55,
                    origin: { x: 1, y: 0.8 },
                    zIndex: 999999,
                    colors: ['#10b981', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']
                });

                if (Date.now() < animationEnd) {
                    requestAnimationFrame(frame);
                }
            }());
        } catch (error) {
            console.error('庆祝动画执行出错:', error);
        }
    }, []);

    // 提交学习结果并移动到下一个单词
    const handleSubmitReview = async (quality: ReviewQuality) => {
        const currentWord = words[currentIndex];
        if (!currentWord || submitting || fetchingNext) return;

        playSuccessSound();
        if (quality === 'know') triggerCelebration();
        setVibratingButton(quality);
        setTimeout(() => setVibratingButton(null), 500);

        setSubmitting(true);

        try {
            // 调用学习 API，包含积分奖励和生词本导入
            const response = await studyWordbookWord(currentWord.id, quality, currentWord.word);

            // 显示积分动画
            if (response.points_earned && response.points_earned > 0) {
                setPointsAnimation({ points: response.points_earned, show: true });
                // 2秒后隐藏动画
                setTimeout(() => {
                    setPointsAnimation({ points: 0, show: false });
                }, 2000);
            }

            // 延迟后进入下一个
            setTimeout(async () => {
                setSubmitting(false);
                const nextGlobalIndex = startOffset + currentIndex + 1;

                if (currentIndex < words.length - 1) {
                    // 还在当前页
                    const nextIdx = currentIndex + 1;
                    setCurrentIndex(nextIdx);
                    handleSaveProgress(nextGlobalIndex).catch(console.error);
                } else {
                    // 到达当前页末尾，尝试加载下一页
                    if (nextGlobalIndex < totalWords) {
                        setFetchingNext(true);
                        const nextPage = Math.floor(nextGlobalIndex / PAGE_SIZE) + 1;
                        const newWords = await fetchPage(nextPage, 0);
                        setFetchingNext(false);

                        if (newWords && newWords.length > 0) {
                            // fetchPage 内部已经更新了 words 和 currentIndex
                            // 这里只需要保存新页第一个词的进度
                            handleSaveProgress(nextGlobalIndex).catch(console.error);
                        } else {
                            console.log('📚 已学完所有单词');
                        }
                    } else {
                        console.log('📚 已学完所有单词');
                        // 可以显示一个完成页面
                    }
                }
            }, 600);
        } catch (error) {
            console.error('处理学习结果失败:', error);
            setSubmitting(false);
        }
    };

    const playPronunciation = useCallback((text: string) => {
        // 取消之前的播放
        window.speechSynthesis.cancel();

        const utterance = new SpeechSynthesisUtterance(text);
        utterance.lang = 'en-US';
        utterance.rate = 0.8;
        window.speechSynthesis.speak(utterance);
    }, []);

    // 当切换到新单词时自动播放发音
    useEffect(() => {
        // 检查是否已经手动播放了音频
        if (manualAudioPlayedRef.current) {
            manualAudioPlayedRef.current = false;
            return;
        }

        const currentWord = words[currentIndex];
        if (currentWord) {
            // 立即播放发音（延迟50ms让UI更新完成）
            setTimeout(() => {
                playPronunciation(currentWord.word);
            }, 50);
        }
    }, [currentIndex, words, playPronunciation]);

    // 键盘快捷键
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (submitting || loading) return;

            const currentWord = words[currentIndex];
            if (!currentWord) return;

            // Ctrl + / : 播放发音
            if (e.ctrlKey && e.code === 'Slash') {
                e.preventDefault();
                playPronunciation(currentWord.word);
                return;
            }

            // 左箭头：不认识
            if (e.key === 'ArrowLeft') {
                e.preventDefault();
                handleSubmitReview('forget');
                return;
            }

            // 上箭头：模糊
            if (e.key === 'ArrowUp') {
                e.preventDefault();
                handleSubmitReview('fuzzy');
                return;
            }

            // 右箭头：认识
            if (e.key === 'ArrowRight') {
                e.preventDefault();
                handleSubmitReview('know');
                return;
            }
        };

        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [currentIndex, words, submitting, loading, handleSubmitReview]);

    const renderMemoryTips = (text: string) => {
        // 检查是否符合特定的编号列表格式 (例如: "1. **联想记忆法**: ...")
        const isFormatted = /\d+\.\s*\*\*.*?\*\*:/.test(text);

        if (!isFormatted) {
            return <div className="rich-text-content">{renderRichText(text)}</div>;
        }

        // 使用正则拆分，保留分隔符以便后续处理，或者直接匹配所有项
        // 这里采用匹配所有项的方式，更稳健
        // 匹配模式：数字. **标题**: 内容
        // 使用 exec 循环匹配或者 matchAll
        const regex = /(\d+\.\s*\*\*(.*?)\*\*:\s*(.*?))(?=\s*\d+\.\s*\*\*|$)/gs;
        const matches = [...text.matchAll(regex)];

        if (matches.length === 0) {
             return <div className="rich-text-content">{renderRichText(text)}</div>;
        }

        return (
            <div className="memory-tips-list">
                {matches.map((match, index) => {
                    // const fullMatch = match[0];
                    const title = match[2]; // 捕获组2是标题
                    const content = match[3]; // 捕获组3是内容

                    return (
                        <div key={index} className="memory-tip-item">
                            <span className="memory-tip-title">{title}：</span>
                            <span className="memory-tip-content">{content}</span>
                        </div>
                    );
                })}
            </div>
        );
    };

    // 简单的 Markdown 解析器
    const renderRichText = (text: string) => {
        if (!text) return null;

        return text.split('\n').map((line, i) => {
            const trimmed = line.trim();
            if (!trimmed) return <br key={i} />;

            // 处理列表项
            const isListItem = trimmed.startsWith('- ') || trimmed.startsWith('* ');
            const content = isListItem ? trimmed.substring(2) : trimmed;

            // 处理加粗 **text**
            const parts = content.split(/(\*\*.*?\*\*)/g);

            const renderedContent = parts.map((part, j) => {
                if (part.startsWith('**') && part.endsWith('**')) {
                    return <strong key={j} className="highlight-text">{part.slice(2, -2)}</strong>;
                }
                return part;
            });

            return (
                <div key={i} className={`rich-line ${isListItem ? 'list-item' : ''}`}>
                    {isListItem && <span className="bullet-point">•</span>}
                    <span className="line-content">{renderedContent}</span>
                </div>
            );
        });
    };

    if (loading) {
        return (
            <div className="wordbook-loading-container">
                <div className="earthworm-loader"></div>
                <p className="loading-text">正在准备高效单词书...</p>
            </div>
        );
    }

    const currentWord = words[currentIndex];
    if (!currentWord) {
        console.error('❌ currentWord is null/undefined:', {
            wordsLength: words.length,
            currentIndex,
            startOffset,
            totalWords,
            words: words.slice(0, 3) // 打印前3个单词看看
        });
        return (
            <div className="wordbook-loading-container">
                <p className="loading-text">数据加载异常，请刷新重试</p>
                <p style={{fontSize: '12px', color: '#999'}}>
                    Debug: words.length={words.length}, currentIndex={currentIndex}
                </p>
            </div>
        );
    }

    // 解析 JSON 字段
    let parsedExamples = [];
    try {
        const rawExamples = currentWord.examples_json ? JSON.parse(currentWord.examples_json) : [];
        // 过滤掉无效的例句（text 为空或者只是空对象的）
        parsedExamples = Array.isArray(rawExamples) ? rawExamples.filter((ex: any) => {
            // 如果是字符串
            if (typeof ex === 'string') return ex.trim().length > 0;
            // 如果是对象，必须有 text 且不为空
            return ex && ex.text && ex.text.trim().length > 0;
        }) : [];
    } catch (e) {
        console.warn('解析例句失败:', e);
    }

    let parsedPhrases = [];
    try {
        const rawPhrases = currentWord.phrases ? JSON.parse(currentWord.phrases) : [];
        // 过滤无效短语
        parsedPhrases = Array.isArray(rawPhrases) ? rawPhrases.filter((ph: any) => {
             if (typeof ph === 'string') return ph.trim().length > 0;
             return ph && (ph.phrase || ph.en) && (ph.phrase || ph.en).trim().length > 0;
        }) : [];
    } catch (e) {
        console.warn('解析短语失败:', e);
    }

    // 尝试解析 meaning_analysis (处理JSON格式的释义)
    let parsedMeaningAnalysis: any[] | null = null;
    if (currentWord.meaning_analysis && currentWord.meaning_analysis.trim().startsWith('[')) {
        try {
            parsedMeaningAnalysis = JSON.parse(currentWord.meaning_analysis);
        } catch (e) {
            // 解析失败，保持原样
        }
    }

    // 解析单词变形
    let parsedWordForms: any = null;
    if (currentWord.word_forms) {
        try {
            parsedWordForms = JSON.parse(currentWord.word_forms);
        } catch (e) {
            // 解析失败，保持原样
        }
    }

    // 检查是否有例句内容
    const hasExamples = parsedExamples.length > 0 || (currentWord.example && currentWord.example.trim().length > 0);

    // 检查是否有词源信息
    const hasEtymology = currentWord.etymology && currentWord.etymology.trim().length > 0;

    // 检查是否有文化背景
    const hasCultural = currentWord.cultural_background && currentWord.cultural_background.trim().length > 0;

    // 检查是否有助记图像
    const hasImage = currentWord.image_url && currentWord.image_url.trim().length > 0;

    // 检查是否有AI绘画解释
    const hasDrawExplain = currentWord.draw_explain && currentWord.draw_explain.trim().length > 0;

    return (
        <>
            <FloatingBackButton show={true} onBack={onBack} />
            <div className={`wordbook-reader ${isFullscreen ? 'immersive' : ''} ${theme === 'dark' ? 'dark' : ''}`}>
                {/* 积分动画 */}
                {pointsAnimation.show && (
                    <div className="points-animation-wordbook">
                        <span className="points-badge">+{pointsAnimation.points}</span>
                        <span className="points-icon">💎</span>
                    </div>
                )}
                {/* 顶部栏 */}
                <div className={`desktop-header ${isFullscreen ? 'hidden' : ''}`}>
                    <div className="header-left">
                        <span className="book-name">{book.name}</span>
                    </div>

                    <div className="header-center">
                        <div className="progress-container-header">
                            <div className="progress-info-text">
                                <span className="current-num">{startOffset + currentIndex + 1}</span>
                                <span className="divider">/</span>
                                <span className="total-num">{totalWords}</span>
                            </div>
                            <div className="progress-bar-desktop">
                                <div
                                    className="progress-fill"
                                    style={{ width: `${((startOffset + currentIndex + 1) / totalWords) * 100}%` }}
                                ></div>
                            </div>
                            <div className="group-info">
                                <span className="group-label">剩</span>
                                <span className="group-value">
                                    {Math.max(0, Math.ceil(totalWords / PAGE_SIZE) - Math.ceil((startOffset + currentIndex + 1) / PAGE_SIZE))}
                                </span>
                                <span className="group-label">组</span>
                            </div>
                        </div>
                    </div>

                    <div className="header-right">
                        <div className="header-controls">
                            {/* 顺序/乱序切换按钮 */}
                            <button
                                className={`control-btn order-btn ${isRandom ? 'random' : ''}`}
                                onClick={handleToggleMode}
                                disabled={switchingMode}
                                title={isRandom ? '当前：乱序模式，点击切换为顺序' : '当前：顺序模式，点击切换为乱序'}
                            >
                                {switchingMode ? (
                                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="spinning">
                                        <circle cx="12" cy="12" r="10"/>
                                        <path d="M12 6v6l4 2"/>
                                    </svg>
                                ) : isRandom ? (
                                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                        <polyline points="16 3 21 3 21 8"></polyline>
                                        <line x1="4" y1="20" x2="21" y2="3"></line>
                                        <polyline points="21 16 21 21 16 21"></polyline>
                                        <line x1="15" y1="15" x2="21" y2="21"></line>
                                        <line x1="4" y1="4" x2="9" y2="9"></line>
                                    </svg>
                                ) : (
                                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                        <line x1="8" y1="6" x2="21" y2="6"></line>
                                        <line x1="8" y1="12" x2="21" y2="12"></line>
                                        <line x1="8" y1="18" x2="21" y2="18"></line>
                                        <line x1="3" y1="6" x2="3.01" y2="6"></line>
                                        <line x1="3" y1="12" x2="3.01" y2="12"></line>
                                        <line x1="3" y1="18" x2="3.01" y2="18"></line>
                                    </svg>
                                )}
                            </button>
                            {/* 全屏沉浸按钮 */}
                            <button
                                className={`control-btn immersive-btn ${isFullscreen ? 'active' : ''}`}
                                onClick={isFullscreen ? exitFullscreen : enterFullscreen}
                                title={isFullscreen ? '退出全屏' : '全屏沉浸'}
                            >
                                {isFullscreen ? (
                                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                        <path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3"/>
                                    </svg>
                                ) : (
                                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                        <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/>
                                    </svg>
                                )}
                            </button>
                            {/* 主题切换 */}
                            <button
                                className="control-btn theme-btn"
                                onClick={toggleTheme}
                                title={theme === 'light' ? '切换深色模式' : '切换浅色模式'}
                            >
                                {theme === 'light' ? (
                                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
                                    </svg>
                                ) : (
                                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                        <circle cx="12" cy="12" r="5"/>
                                        <line x1="12" y1="1" x2="12" y2="3"/>
                                        <line x1="12" y1="21" x2="12" y2="23"/>
                                        <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
                                        <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
                                        <line x1="1" y1="12" x2="3" y2="12"/>
                                        <line x1="21" y1="12" x2="23" y2="12"/>
                                        <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
                                        <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
                                    </svg>
                                )}
                            </button>
                        </div>
                    </div>
                </div>

                {/* 全屏下的迷你控制栏 */}
                {isFullscreen && (
                    <div className="immersive-controls">
                        <button className="mini-control" onClick={exitFullscreen} title="退出全屏">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                <path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3"/>
                            </svg>
                        </button>
                        <span className="mini-progress">{startOffset + currentIndex + 1}/{totalWords}</span>
                    </div>
                )}

                <div className="reader-content">
                    <div className="word-card">
                    {/* 单词头部 - 固定在顶部 */}
                    <div className="word-header">
                        {/* 快捷键提示按钮 */}
                        <div className="card-shortcut-container">
                            <button
                                className={`shortcut-icon-btn ${showShortcuts ? 'active' : ''}`}
                                onClick={() => setShowShortcuts(!showShortcuts)}
                                title="快捷键说明"
                            >
                                <span className="icon-text">!</span>
                            </button>
                            <div className={`shortcut-popup-card ${showShortcuts ? 'visible' : ''}`}>
                                <div className="popup-title">快捷键</div>
                                <div className="shortcut-list">
                                    <div className="shortcut-row">
                                        <span className="key-label">Ctrl + /</span>
                                        <span className="key-desc">播放发音</span>
                                    </div>
                                    <div className="shortcut-row">
                                        <span className="key-label">←</span>
                                        <span className="key-desc">不认识</span>
                                    </div>
                                    <div className="shortcut-row">
                                        <span className="key-label">↑</span>
                                        <span className="key-desc">模糊</span>
                                    </div>
                                    <div className="shortcut-row">
                                        <span className="key-label">→</span>
                                        <span className="key-desc">认识</span>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div className="word-title-row">
                            <h1 className="word-text">{currentWord.word}</h1>
                            <button className="audio-btn" onClick={() => playPronunciation(currentWord.word)} title="点击播放发音">
                                <img src="/icon/laba.png" alt="播放发音" className="audio-icon-img" />
                            </button>
                        </div>
                        <div className="phonetic-row">
                            <span className="phonetic">[{currentWord.phonetic}]</span>
                        </div>
                    </div>

                    {/* 可滚动内容区 */}
                    <div className="card-scroll-area">
                        {/* 第一行：核心释义 - 优先级最高 */}
                        <div className={`main-content-grid ${!hasExamples ? 'single-col' : ''}`}>
                            <div className="left-col">
                                {/* 1. 中文释义（简短） - 必展示 */}
                                {currentWord.meaning && currentWord.meaning.trim() && (
                                    <div className="section">
                                        <div className="section-title">释义</div>
                                        <p className="meaning-text">{currentWord.meaning}</p>
                                    </div>
                                )}

                                {/* 2. 详细释义分析（词性+释义） - 优先展示 */}
                                {currentWord.meaning_analysis && currentWord.meaning_analysis.trim() && (
                                    <div className="section">
                                        <div className="section-title">详细释义</div>
                                        {parsedMeaningAnalysis ? (
                                            <div className="analysis-tags">
                                                {parsedMeaningAnalysis.map((item: any, i: number) => (
                                                    <div key={i} className="analysis-tag-item">
                                                        <span className="pos-tag">{item.pos}</span>
                                                        <span className="cn-text">{item.cn}</span>
                                                    </div>
                                                ))}
                                            </div>
                                        ) : (
                                            <div className="analysis-box rich-text-content">
                                                {renderRichText(currentWord.meaning_analysis)}
                                            </div>
                                        )}
                                    </div>
                                )}

                                {/* 7. 单词变形 - 优先展示 */}
                                {parsedWordForms && Object.keys(parsedWordForms).length > 0 && (
                                    <div className="section word-forms-section">
                                        <div className="section-title">单词变形</div>
                                        <div className="word-forms-grid">
                                            {parsedWordForms.plural && (
                                                <div className="word-form-item">
                                                    <span className="form-label">复数</span>
                                                    <span className="form-value">{parsedWordForms.plural}</span>
                                                </div>
                                            )}
                                            {parsedWordForms.past && (
                                                <div className="word-form-item">
                                                    <span className="form-label">过去式</span>
                                                    <span className="form-value">{parsedWordForms.past}</span>
                                                </div>
                                            )}
                                            {parsedWordForms.past_participle && (
                                                <div className="word-form-item">
                                                    <span className="form-label">过去分词</span>
                                                    <span className="form-value">{parsedWordForms.past_participle}</span>
                                                </div>
                                            )}
                                            {parsedWordForms.present_participle && (
                                                <div className="word-form-item">
                                                    <span className="form-label">现在分词</span>
                                                    <span className="form-value">{parsedWordForms.present_participle}</span>
                                                </div>
                                            )}
                                            {parsedWordForms.third_person && (
                                                <div className="word-form-item">
                                                    <span className="form-label">第三人称</span>
                                                    <span className="form-value">{parsedWordForms.third_person}</span>
                                                </div>
                                            )}
                                            {parsedWordForms.comparative && (
                                                <div className="word-form-item">
                                                    <span className="form-label">比较级</span>
                                                    <span className="form-value">{parsedWordForms.comparative}</span>
                                                </div>
                                            )}
                                            {parsedWordForms.superlative && (
                                                <div className="word-form-item">
                                                    <span className="form-label">最高级</span>
                                                    <span className="form-value">{parsedWordForms.superlative}</span>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                )}
                            </div>

                            {/* 5. 例句 - 优先展示（仅在有例句时显示右侧栏） */}
                            {hasExamples && (
                                <div className="right-col">
                                    <div className="section">
                                        <div className="section-title">精选例句</div>
                                        <div className="example-list">
                                            {parsedExamples.length > 0 ? parsedExamples.map((ex: any, i: number) => (
                                                <div key={i} className="example-item">
                                                    <p className="en">{ex.text}</p>
                                                    <p className="cn">{ex.translation}</p>
                                                </div>
                                            )) : (
                                                currentWord.example && currentWord.example.trim() && (
                                                    <div className="example-item">
                                                        <p className="en">{currentWord.example}</p>
                                                        {currentWord.example_translation && (
                                                            <p className="cn">{currentWord.example_translation}</p>
                                                        )}
                                                    </div>
                                                )
                                            )}
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>

                        {/* 6. 固定搭配和短语 - 优先展示 */}
                        {parsedPhrases.length > 0 && (
                            <div className="section">
                                <div className="section-title">固定搭配</div>
                                <div className="phrases-grid">
                                    {parsedPhrases.map((p: any, i: number) => (
                                        <div key={i} className="phrase-item">
                                            <span className="p-en">{p.phrase || p.en}</span>
                                            <span className="p-cn">{p.translation || p.cn}</span>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}

                        {/* 8. 记忆技巧 - 深度学习层 */}
                        {currentWord.memory_tips && currentWord.memory_tips.trim() && (
                            <div className="section">
                                <div className="section-title">💡 记忆技巧</div>
                                <div className="memory-tips-content">
                                    {renderMemoryTips(currentWord.memory_tips)}
                                </div>
                            </div>
                        )}

                        {/* 9-10. 词根词缀 - 深度学习层 */}
                        {((currentWord.root && currentWord.root.trim()) ||
                          (currentWord.affix && currentWord.affix.trim())) && (
                            <div className="section">
                                <div className="section-title">🔤 词根词缀</div>
                                <div className="root-affix-content">
                                    {currentWord.root && currentWord.root.trim() && (
                                        <div className="root-block">
                                            <span className="root-tag">词根: {currentWord.root}</span>
                                            {currentWord.root_analysis && currentWord.root_analysis.trim() && (
                                                <div className="root-analysis rich-text-container">
                                                    {renderRichText(currentWord.root_analysis)}
                                                </div>
                                            )}
                                        </div>
                                    )}
                                    {currentWord.affix && currentWord.affix.trim() && (
                                        <div className="affix-block">
                                            <span className="affix-tag">词缀: {currentWord.affix}</span>
                                            {currentWord.affix_analysis && currentWord.affix_analysis.trim() && (
                                                <div className="affix-analysis rich-text-container">
                                                    {renderRichText(currentWord.affix_analysis)}
                                                </div>
                                            )}
                                        </div>
                                    )}
                                </div>
                            </div>
                        )}

                        {/* 11-12. 词源 & 文化背景 - 深度学习层 */}
                        {(hasEtymology || hasCultural) && (
                            <div className="etymology-cultural-grid">
                                {hasEtymology && (
                                    <div className="section etymology-section">
                                        <div className="section-title">📜 词源</div>
                                        <div className="rich-text-container">{renderRichText(currentWord.etymology)}</div>
                                    </div>
                                )}
                                {hasCultural && (
                                    <div className="section cultural-section">
                                        <div className="section-title">🌍 文化背景</div>
                                        <div className="rich-text-container">{renderRichText(currentWord.cultural_background)}</div>
                                    </div>
                                )}
                            </div>
                        )}

                        {/* 13. 助记图像 - 深度学习层 */}
                        {hasImage && (
                            <div className="section image-section">
                                <div className="section-title">🎨 助记图像</div>
                                <div className="memory-image-container">
                                    <img src={currentWord.image_url} alt={currentWord.word} className="memory-image" />
                                    {hasDrawExplain && (
                                        <p className="draw-explain">{currentWord.draw_explain}</p>
                                    )}
                                </div>
                            </div>
                        )}

                        {/* 14. 场景故事 - 深度学习层 */}
                        {currentWord.story_en && currentWord.story_en.trim() && (
                            <div className="section">
                                <div className="section-title">📖 场景故事</div>
                                <div className="story-content">
                                    <p className="en" style={{ fontStyle: 'italic', marginBottom: '8px' }}>
                                        {currentWord.story_en}
                                    </p>
                                    {currentWord.story_cn && currentWord.story_cn.trim() && (
                                        <p className="cn" style={{ color: 'var(--text-secondary)' }}>
                                            {currentWord.story_cn}
                                        </p>
                                    )}
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            <div className="reader-footer">
                {/* 三个按钮 - 极简样式 */}
                <div className="action-buttons-wordbook">
                    <button
                        className={`action-btn-wordbook forget ${vibratingButton === 'forget' ? 'vibrating' : ''}`}
                        onClick={() => handleSubmitReview('forget')}
                        disabled={submitting}
                        data-key="←"
                    >
                        <span className="btn-label">不认识</span>
                    </button>
                    <button
                        className={`action-btn-wordbook fuzzy ${vibratingButton === 'fuzzy' ? 'vibrating' : ''}`}
                        onClick={() => handleSubmitReview('fuzzy')}
                        disabled={submitting}
                        data-key="↑"
                    >
                        <span className="btn-label">模糊</span>
                    </button>
                    <button
                        className={`action-btn-wordbook know ${vibratingButton === 'know' ? 'vibrating' : ''}`}
                        onClick={() => handleSubmitReview('know')}
                        disabled={submitting}
                        data-key="→"
                    >
                        <span className="btn-label">认识</span>
                    </button>
                </div>
            </div>
        </div>
        </>
    );
};
