import React, { useState, useEffect, useCallback, useRef } from 'react';
import confetti from 'canvas-confetti';
import {
    getTodayReviewList,
    getVocabularyList,
    submitVocabularyReview,
    type Vocabulary
} from '../services/api';
import { useStore } from '../store/useStore';
import { useTypingSound } from '../hooks/useTypingSound';
import { FloatingBackButton } from './FloatingBackButton';
import './ReviewSession.css';
import './ReviewSession.theme.css';

// 音效
const successSound = new Audio('/sounds/right.mp3');
successSound.volume = 0.6;

interface ReviewSessionProps {
    onClose: () => void;
    onComplete?: () => void;
}

type ReviewMode = 'card' | 'spell';  // 卡片模式 / 默写模式
type ReviewQuality = 'know' | 'fuzzy' | 'forget';  // 认识 / 模糊 / 忘记

// 将用户选择转换为SM-2评分 (0-5)
const qualityToScore = (quality: ReviewQuality): number => {
    switch (quality) {
        case 'know': return 5;    // 完全记得
        case 'fuzzy': return 3;   // 模糊记得
        case 'forget': return 1;  // 完全忘记
        default: return 3;
    }
};

export const ReviewSession: React.FC<ReviewSessionProps> = ({ onClose, onComplete }) => {
    const { theme, setTheme } = useStore();
    const [mode, setMode] = useState<ReviewMode>('card');
    const [reviewList, setReviewList] = useState<Vocabulary[]>([]);
    const [currentIndex, setCurrentIndex] = useState(0);
    const [loading, setLoading] = useState(true);
    const [userInput, setUserInput] = useState('');
    const [showSpellResult, setShowSpellResult] = useState(false);
    const [showHint, setShowHint] = useState(false);  // Tab键显示提示
    const [showShortcuts, setShowShortcuts] = useState(false); // 快捷键提示开关
    const [submitting, setSubmitting] = useState(false);
    const [sessionStats, setSessionStats] = useState({ know: 0, fuzzy: 0, forget: 0 });
    const [isComplete, setIsComplete] = useState(false);
    const [startTime, setStartTime] = useState<number>(0);
    const [isFullscreen, setIsFullscreen] = useState(false);  // 真正的全屏状态
    const inputRef = useRef<HTMLInputElement>(null);
    const manualAudioPlayedRef = useRef(false); // 标记是否手动播放了音频，用于避免useEffect重复播放

    // 智能复习追踪
    const [reviewedIds, setReviewedIds] = useState<Set<number>>(new Set());  // 本轮已复习的单词ID
    const [fuzzyWords, setFuzzyWords] = useState<Vocabulary[]>([]);  // 本轮标记为"模糊"的单词
    const [forgetWords, setForgetWords] = useState<Vocabulary[]>([]);  // 本轮标记为"忘记"的单词
    const BATCH_SIZE = 20;  // 每批最多20个单词

    // 积分动画状态
    const [pointsAnimation, setPointsAnimation] = useState<{ points: number; show: boolean }>({ points: 0, show: false });

    // 打字音效
    const { playTypingSound, shouldPlayTypingSound } = useTypingSound();

    // 进入全屏模式
    const enterFullscreen = useCallback(() => {
        console.log('📺 enterFullscreen called in ReviewSession');
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
        console.log('📺 exitFullscreen called in ReviewSession');
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
        
        // 初始化检查
        handleFullscreenChange();

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

    const currentVocab = reviewList[currentIndex];
    const progress = reviewList.length > 0 ? ((currentIndex + 1) / reviewList.length) * 100 : 0;

    // 获取待复习列表
    useEffect(() => {
        fetchReviewList();
    }, []);

    // 过滤有效单词的辅助函数
    const filterValidWords = (words: Vocabulary[]) => {
        // 只要有单词内容即可复习，释义为空也可以（UI会显示暂无释义）
        return words.filter((vocab: Vocabulary) =>
            vocab.content && vocab.content.trim()
        );
    };

    const fetchReviewList = async () => {
        setLoading(true);
        try {
            // 首次获取今日待复习列表
            const result = await getTodayReviewList(BATCH_SIZE);
            if (result.data && result.data.length > 0) {
                const validWords = filterValidWords(result.data);

                if (validWords.length > 0) {
                    const shuffled = [...validWords].sort(() => Math.random() - 0.5);
                    setReviewList(shuffled);
                    // 记录已复习的单词ID
                    setReviewedIds(new Set(shuffled.map(w => w.id)));
                }
            } else {
                // 如果没有今日待复习，从生词本获取复习次数最少的单词
                await fetchMoreWords(new Set());
            }
        } catch (error) {
            console.error('获取复习列表失败:', error);
        } finally {
            setLoading(false);
        }
    };

    // 从生词本获取更多单词（按复习次数排序，排除已复习的）
    const fetchMoreWords = async (excludeIds: Set<number>) => {
        try {
            // 获取按复习次数升序排列的单词（复习次数少的优先）
            const result = await getVocabularyList({
                order_by: 'review_count',
                limit: BATCH_SIZE * 2,  // 多获取一些，以便过滤
            });

            if (result.data && result.data.length > 0) {
                // 过滤掉已复习的和无效的单词
                const validWords = filterValidWords(result.data)
                    .filter(w => !excludeIds.has(w.id))
                    .slice(0, BATCH_SIZE);

                if (validWords.length > 0) {
                    const shuffled = [...validWords].sort(() => Math.random() - 0.5);
                    setReviewList(shuffled);
                    return shuffled;
                }
            }
            return [];
        } catch (error) {
            console.error('获取更多单词失败:', error);
            return [];
        }
    };

    // 再来一轮 - 智能复习逻辑
    const restartReview = async () => {
        setLoading(true);

        try {
            // 收集上一轮"模糊"和"忘记"的单词（优先复习）
            const priorityWords = [...forgetWords, ...fuzzyWords];
            const priorityIds = new Set(priorityWords.map(w => w.id));

            let newList: Vocabulary[] = [];

            // 1. 首先加入"忘记"和"模糊"的单词
            if (priorityWords.length > 0) {
                newList = [...priorityWords].slice(0, BATCH_SIZE);
            }

            // 2. 如果不够20个，从生词本补充（按复习次数少的优先，排除本轮已复习的）
            if (newList.length < BATCH_SIZE) {
                const needed = BATCH_SIZE - newList.length;
                const result = await getVocabularyList({
                    order_by: 'review_count',
                    limit: needed * 2,
                });

                if (result.data && result.data.length > 0) {
                    const additionalWords = filterValidWords(result.data)
                        .filter(w => !reviewedIds.has(w.id) && !priorityIds.has(w.id))
                        .slice(0, needed);
                    newList = [...newList, ...additionalWords];
                }
            }

            // 3. 如果还是没有，说明所有单词都复习过了，重置已复习列表，从头再来
            if (newList.length === 0) {
                setReviewedIds(new Set());
                const result = await getVocabularyList({
                    order_by: 'review_count',
                    limit: BATCH_SIZE,
                });
                if (result.data && result.data.length > 0) {
                    newList = filterValidWords(result.data).slice(0, BATCH_SIZE);
                }
            }

            if (newList.length > 0) {
                // 优化排序逻辑：优先复习"忘记"和"模糊"的单词，不打乱它们与新词的顺序
                // 1. 分离优先词和新词
                const priorityPart = newList.filter(w => priorityIds.has(w.id));
                const newPart = newList.filter(w => !priorityIds.has(w.id));
                
                // 2. 分别打乱（或者保持优先词的相对顺序，仅打乱新词）
                // 这里选择分别打乱，但保证优先词总是在前面
                const shuffledPriority = [...priorityPart].sort(() => Math.random() - 0.5);
                const shuffledNew = [...newPart].sort(() => Math.random() - 0.5);
                
                // 3. 合并
                const finalOrderedList = [...shuffledPriority, ...shuffledNew];
                
                setReviewList(finalOrderedList);
                // 更新已复习ID
                setReviewedIds(prev => {
                    const newIds = new Set(prev);
                    finalOrderedList.forEach(w => newIds.add(w.id));
                    return newIds;
                });
            }

            // 重置状态
            setCurrentIndex(0);
            setIsComplete(false);
            setSessionStats({ know: 0, fuzzy: 0, forget: 0 });
            setFuzzyWords([]);
            setForgetWords([]);
            setUserInput('');
            setShowSpellResult(false);
            setShowHint(false);
            setStartTime(Date.now());
        } catch (error) {
            console.error('获取新一轮复习列表失败:', error);
        } finally {
            setLoading(false);
        }
    };

    // 播放指定文本
    const speakText = useCallback((text: string) => {
        if (text) {
            // 取消之前的播放
            speechSynthesis.cancel();
            
            const utterance = new SpeechSynthesisUtterance(text);
            utterance.lang = 'en-US';
            utterance.rate = 0.8;
            speechSynthesis.speak(utterance);
        }
    }, []);

    // 播放单词发音
    const playWordAudio = useCallback(() => {
        if (currentVocab?.content) {
            speakText(currentVocab.content);
        }
    }, [currentVocab, speakText]);

    // 播放成功音效
    const playSuccessSound = useCallback(() => {
        try {
            successSound.currentTime = 0;
            successSound.play().catch(e => console.warn('播放音效失败:', e));
        } catch (e) {
            console.warn('播放音效失败:', e);
        }
    }, []);

    // 庆祝动效 - 点击"认识"时喷射彩色粒子
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

    // 当切换到新单词时自动播放发音
    useEffect(() => {
        setUserInput('');
        setShowSpellResult(false);
        setShowHint(false);  // 重置提示状态
        setSubmitting(false); // 解除提交锁定
        setStartTime(Date.now());

        // 检查是否已经手动播放了音频（在点击按钮时）
        if (manualAudioPlayedRef.current) {
            manualAudioPlayedRef.current = false;
            // 依然需要聚焦输入框
            if (mode === 'spell' && inputRef.current) {
                setTimeout(() => {
                    inputRef.current?.focus();
                }, 100);
            }
            return;
        }

        // 无论是卡片模式还是默写模式，都播放发音
        if (currentVocab) {
            // 尝试直接播放，如果不工作则使用setTimeout作为回退
            // 注意：在移动端如果没有用户交互，这里可能无法播放，
            // 所以主要依赖 handleSubmitReview 中的同步播放
            setTimeout(() => {
                playWordAudio();
            }, 300);
        }

        if (mode === 'spell' && inputRef.current) {
            setTimeout(() => {
                inputRef.current?.focus();
            }, 100);
        }
    }, [currentIndex, mode, currentVocab, playWordAudio]);

    // 生成提示信息 - 显示完整单词
    const getHintInfo = () => {
        if (!currentVocab?.content) return null;
        const word = currentVocab.content;
        const length = word.length;
        // 直接显示完整单词
        return { word, length };
    };

    // 提交复习结果
    const handleSubmitReview = async (quality: ReviewQuality) => {
        if (!currentVocab || submitting) return;

        // 1. 播放音效与特效
        playSuccessSound();
        if (quality === 'know') triggerCelebration();

        // 2. 锁定并准备数据
        setSubmitting(true);
        const vocabToSubmit = currentVocab;
        const responseTime = Date.now() - startTime;
        const score = qualityToScore(quality);
        const userAnswer = mode === 'spell' ? userInput : undefined;

        // 3. 乐观更新 (Optimistic Updates)
        if (quality === 'fuzzy') {
            setFuzzyWords(prev => {
                if (!prev.find(w => w.id === vocabToSubmit.id)) return [...prev, vocabToSubmit];
                return prev;
            });
        } else if (quality === 'forget') {
            setForgetWords(prev => {
                if (!prev.find(w => w.id === vocabToSubmit.id)) return [...prev, vocabToSubmit];
                return prev;
            });
        }

        setSessionStats(prev => ({
            ...prev,
            [quality]: prev[quality] + 1
        }));

        // 4. 立即切换到下一个 (确保音频在用户交互中触发)
        if (currentIndex < reviewList.length - 1) {
            const nextIndex = currentIndex + 1;
            const nextVocab = reviewList[nextIndex];
            
            // 关键修复：在用户交互事件中直接播放下一个音频，
            // 这样移动端浏览器（Safari/Chrome）就不会拦截自动播放
            if (nextVocab?.content) {
                speakText(nextVocab.content);
                manualAudioPlayedRef.current = true;
            }
            
            setCurrentIndex(nextIndex);
        } else {
            setIsComplete(true);
            onComplete?.();
        }

        // 5. 后台静默提交
        try {
            const response = await submitVocabularyReview(vocabToSubmit.id, {
                review_type: mode,
                quality: score,
                response_time_ms: responseTime,
                user_answer: userAnswer,
            });

            // 积分动画 (仅成功时显示)
            if (response.points_earned && response.points_earned > 0) {
                setPointsAnimation({ points: response.points_earned, show: true });
                setTimeout(() => {
                    setPointsAnimation({ points: 0, show: false });
                }, 2000);
            }
        } catch (error) {
            console.error('提交复习结果失败:', error);
        }
        // 注意：setSubmitting(false) 移至 useEffect 中，在索引切换完成后执行
    };

    // 检查默写答案
    const checkSpellAnswer = () => {
        if (!currentVocab) return false;
        return userInput.toLowerCase().trim() === currentVocab.content.toLowerCase().trim();
    };

    // 按钮震动状态
    const [vibratingButton, setVibratingButton] = useState<ReviewQuality | null>(null);

    // 自动评判默写结果并提交
    const autoJudgeAndSubmit = useCallback(() => {
        if (!currentVocab || submitting) return;

        const isCorrect = userInput.toLowerCase().trim() === currentVocab.content.toLowerCase().trim();
        let quality: ReviewQuality;

        if (!showHint) {
            // 没有看提示
            if (isCorrect) {
                quality = 'know';  // 没看提示 + 正确 = 认识
            } else {
                quality = 'fuzzy'; // 没看提示 + 错误 = 模糊
            }
        } else {
            // 看了提示
            if (isCorrect) {
                quality = 'fuzzy';  // 看了提示 + 正确 = 模糊
            } else {
                quality = 'forget'; // 看了提示 + 错误 = 忘记
            }
        }

        // 显示结果
        setShowSpellResult(true);

        // 触发按钮震动效果
        setVibratingButton(quality);
        setTimeout(() => setVibratingButton(null), 500);

        // 延迟一点提交，让用户看到结果
        setTimeout(() => {
            handleSubmitReview(quality);
        }, 800);
    }, [currentVocab, userInput, showHint, submitting, handleSubmitReview]);

    // 解析词根词缀JSON
    const parseMorphemes = (morphemesJson?: string) => {
        if (!morphemesJson) return [];
        try {
            return JSON.parse(morphemesJson);
        } catch {
            return [];
        }
    };

    // 处理键盘事件
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (isComplete) return;

            // Esc键：关闭或重新输入
            if (e.key === 'Escape') {
                if (mode === 'spell' && userInput && !showSpellResult) {
                    // 在默写模式有输入时，清空输入
                    e.preventDefault();
                    setUserInput('');
                    setShowHint(false);
                    inputRef.current?.focus();
                } else {
                    onClose();
                }
                return;
            }

            // Ctrl + / : 播放发音
            if (e.ctrlKey && e.code === 'Slash') {
                e.preventDefault();
                playWordAudio();
                return;
            }

            // Tab键切换提示（仅在默写模式且未显示答案时）
            if (e.key === 'Tab' && mode === 'spell' && !showSpellResult) {
                e.preventDefault();
                setShowHint(prev => !prev);
                return;
            }

            // R键：重新默写（显示结果后且答错时）
            if (e.code === 'KeyR' && mode === 'spell' && showSpellResult && !checkSpellAnswer()) {
                e.preventDefault();
                setUserInput('');
                setShowSpellResult(false);
                setShowHint(false);
                inputRef.current?.focus();
                return;
            }

            if (mode === 'card' || (mode === 'spell' && showSpellResult)) {
                if (e.key === 'ArrowLeft') {
                    e.preventDefault();
                    handleSubmitReview('forget');
                } else if (e.key === 'ArrowUp') {
                    e.preventDefault();
                    handleSubmitReview('fuzzy');
                } else if (e.key === 'ArrowRight') {
                    e.preventDefault();
                    handleSubmitReview('know');
                }
            }
        };

        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [mode, isComplete, currentVocab, showSpellResult, showHint, userInput, playWordAudio]);

    // 加载中
    if (loading) {
        return (
            <div className="review-session-desktop">
                <div className="review-loading">
                    <div className="loading-spinner-large"></div>
                    <p>正在加载复习内容...</p>
                </div>
            </div>
        );
    }

    // 没有待复习内容
    if (reviewList.length === 0) {
        return (
            <div className="review-session-desktop">
                <div className="review-empty">
                    <div className="empty-icon">🎉</div>
                    <h2>太棒了！</h2>
                    <p>今天没有需要复习的单词</p>
                    <div className="empty-actions">
                        <button className="empty-back-btn" onClick={onClose}>返回生词本</button>
                        <button className="empty-restart-btn" onClick={restartReview}>
                            开始复习
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    // 复习完成
    if (isComplete) {
        const total = sessionStats.know + sessionStats.fuzzy + sessionStats.forget;
        const accuracy = total > 0 ? Math.round((sessionStats.know / total) * 100) : 0;
        const getMessage = () => {
            if (accuracy >= 80) return { text: '太棒了！记忆力超群！', level: 'excellent' };
            if (accuracy >= 60) return { text: '做得不错，继续加油！', level: 'good' };
            return { text: '别灰心，多复习几次！', level: 'normal' };
        };
        const message = getMessage();

        return (
            <>
                <FloatingBackButton show={true} onBack={onClose} />
                <div className="review-session-desktop">
                    {/* 庆祝粒子效果 */}
                    <div className="confetti-container">
                    {[...Array(30)].map((_, i) => (
                        <div key={i} className={`confetti confetti-${i % 6}`} style={{
                            left: `${Math.random() * 100}%`,
                            animationDelay: `${Math.random() * 2}s`,
                            animationDuration: `${3 + Math.random() * 2}s`
                        }} />
                    ))}
                </div>

                <div className="complete-page">
                    {/* 奖杯区域 */}
                    <div className="trophy-section">
                        <div className="trophy-glow"></div>
                        <img src="/icon/jiangbei.png" alt="奖杯" className="trophy-icon" />
                    </div>

                    {/* 标题 */}
                    <h1 className="complete-title">复习完成！</h1>
                    <p className={`complete-message ${message.level}`}>{message.text}</p>

                    {/* 数据卡片 */}
                    <div className="stats-grid">
                        <div className="stat-card main-stat">
                            <div className="stat-number">{total}</div>
                            <div className="stat-label">复习单词</div>
                        </div>
                        <div className="stat-card accuracy-stat">
                            <div className="stat-number">{accuracy}<span className="stat-unit">%</span></div>
                            <div className="stat-label">正确率</div>
                            <div className="accuracy-bar">
                                <div className="accuracy-fill" style={{ width: `${accuracy}%` }}></div>
                            </div>
                        </div>
                    </div>

                    {/* 详细统计 */}
                    <div className="breakdown-cards">
                        <div className="breakdown-card know">
                            <div className="breakdown-number">{sessionStats.know}</div>
                            <div className="breakdown-text">认识</div>
                            <div className="breakdown-icon">✓</div>
                        </div>
                        <div className="breakdown-card fuzzy">
                            <div className="breakdown-number">{sessionStats.fuzzy}</div>
                            <div className="breakdown-text">模糊</div>
                            <div className="breakdown-icon">~</div>
                        </div>
                        <div className="breakdown-card forget">
                            <div className="breakdown-number">{sessionStats.forget}</div>
                            <div className="breakdown-text">忘记</div>
                            <div className="breakdown-icon">✗</div>
                        </div>
                    </div>

                    {/* 操作按钮 */}
                    <div className="complete-actions">
                        <button className="action-btn-secondary" onClick={onClose}>
                            返回生词本
                        </button>
                        <button className="action-btn-primary" onClick={restartReview}>
                            再来一轮
                        </button>
                    </div>
                </div>
            </div>
            </>
        );
    }

    const morphemes = parseMorphemes(currentVocab?.morphemes);

    return (
        <>
            <FloatingBackButton show={true} onBack={onClose} />
            <div className={`review-session-desktop ${isFullscreen ? 'immersive' : ''} ${theme === 'dark' ? 'dark' : ''}`}>
            {/* 顶部栏 */}
            <div className={`desktop-header ${isFullscreen ? 'hidden' : ''}`}>
                <div className="progress-container">
                    <span className="header-label">每日复习</span>
                    <div className="progress-bar-desktop">
                        <div className="progress-fill" style={{ width: `${progress}%` }}></div>
                    </div>
                    <span className="progress-text">{currentIndex + 1} / {reviewList.length}</span>
                </div>

                {/* 积分动画 */}
                {pointsAnimation.show && (
                    <div className="points-animation">
                        <span className="points-badge">+{pointsAnimation.points}</span>
                        <span className="points-icon">💎</span>
                    </div>
                )}
                <div className="header-controls">
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
                    {/* 模式切换 */}
                    <div className="mode-switch">
                        <button
                            className={`mode-btn ${mode === 'card' ? 'active' : ''}`}
                            onClick={() => setMode('card')}
                        >
                            卡片模式
                        </button>
                        <button
                            className={`mode-btn ${mode === 'spell' ? 'active' : ''}`}
                            onClick={() => setMode('spell')}
                        >
                            默写模式
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
                    <span className="mini-progress">{currentIndex + 1}/{reviewList.length}</span>
                </div>
            )}

            {/* 主内容区 */}
            <div className="desktop-main">
                {mode === 'card' ? (
                    // 卡片模式 - 商业化大卡片布局
                    <div className="review-card-wrapper">
                        <div className="word-card-commercial">
                            {/* 头部：单词 + 发音 */}
                            <div className="word-header">
                                {/* 快捷键按钮 - 卡片模式 */}
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
                                                <span className="key-desc">忘记</span>
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
                                    <h1 className="word-text">{currentVocab?.content}</h1>
                                    <button className="audio-btn" onClick={playWordAudio} title="点击播放发音">
                                        <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
                                            <path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.26 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"/>
                                        </svg>
                                    </button>
                                </div>
                                {currentVocab?.phonetic && (
                                    <div className="phonetic-row">
                                        <span className="phonetic">[{currentVocab.phonetic}]</span>
                                    </div>
                                )}
                            </div>

                            {/* 滚动内容区 */}
                            <div className="card-scroll-content">
                                <div className="info-grid">
                                    {/* 左侧：释义 + 词根 */}
                                    <div className="info-left">
                                        <div className="section meaning-section">
                                            <div className="section-title">中文释义</div>
                                            <div className="meaning-text">{currentVocab?.meaning || '暂无释义'}</div>
                                        </div>

                                        {morphemes.length > 0 && (
                                            <div className="section morpheme-section">
                                                <div className="section-title">词根词缀</div>
                                                <div className="morpheme-list">
                                                    {morphemes.map((m: any, i: number) => (
                                                        <div key={i} className={`morpheme-item ${m.type}`}>
                                                            <span className="morpheme-name">{m.morpheme}</span>
                                                            <span className="morpheme-meaning">{m.meaning}</span>
                                                        </div>
                                                    ))}
                                                </div>
                                            </div>
                                        )}
                                    </div>

                                    {/* 右侧：例句 + 上下文 */}
                                    <div className="info-right">
                                        <div className="section example-section">
                                            <div className="section-title">精选例句</div>
                                            {currentVocab?.example ? (
                                                <div className="example-item">
                                                    <p className="en">{currentVocab.example}</p>
                                                    {currentVocab.example_translation && (
                                                        <p className="cn">{currentVocab.example_translation}</p>
                                                    )}
                                                </div>
                                            ) : (
                                                <div className="empty-text">暂无例句</div>
                                            )}
                                        </div>

                                        {currentVocab?.context && (
                                            <div className="section context-section">
                                                <div className="section-title">来源语境</div>
                                                <div className="context-card">
                                                    <div className="context-icon">❝</div>
                                                    <div className="context-text">{currentVocab.context}</div>
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                ) : (
                    // 默写模式 - 左右布局
                    <div className={`spell-layout-split ${(!showHint && !showSpellResult) ? 'center-mode' : ''}`}>
                        {/* 左侧 - 输入区域 */}
                        <div className="spell-left">
                            {/* 快捷键按钮 - 默写模式 */}
                            <div className="spell-shortcut-container">
                                <button 
                                    className={`shortcut-icon-btn ${showShortcuts ? 'active' : ''}`}
                                    onClick={() => setShowShortcuts(!showShortcuts)}
                                    title="快捷键说明"
                                >
                                    <span className="icon-text">!</span>
                                </button>
                                <div className={`shortcut-popup-spell ${showShortcuts ? 'visible' : ''}`}>
                                    <div className="popup-title">快捷键</div>
                                    <div className="shortcut-list">
                                        <div className="shortcut-row">
                                            <span className="key-label">Ctrl + /</span>
                                            <span className="key-desc">播放发音</span>
                                        </div>
                                        {!showSpellResult && (
                                            <>
                                                <div className="shortcut-row">
                                                    <span className="key-label">Space</span>
                                                    <span className="key-desc">自动评判</span>
                                                </div>
                                                <div className="shortcut-row">
                                                    <span className="key-label">Tab</span>
                                                    <span className="key-desc">查看提示</span>
                                                </div>
                                                <div className="shortcut-row">
                                                    <span className="key-label">Esc</span>
                                                    <span className="key-desc">清空输入</span>
                                                </div>
                                            </>
                                        )}
                                        {showSpellResult && !checkSpellAnswer() && (
                                            <div className="shortcut-row">
                                                <span className="key-label">R</span>
                                                <span className="key-desc">重新默写</span>
                                            </div>
                                        )}
                                        {showSpellResult && (
                                            <>
                                                <div className="shortcut-row">
                                                    <span className="key-label">←</span>
                                                    <span className="key-desc">忘记</span>
                                                </div>
                                                <div className="shortcut-row">
                                                    <span className="key-label">↑</span>
                                                    <span className="key-desc">模糊</span>
                                                </div>
                                                <div className="shortcut-row">
                                                    <span className="key-label">→</span>
                                                    <span className="key-desc">认识</span>
                                                </div>
                                            </>
                                        )}
                                    </div>
                                </div>
                            </div>

                            {/* 提示信息 */}
                            <div className="spell-hint-section">
                                <div className="prompt-meaning-large">
                                    {(() => {
                                        const meaning = currentVocab?.meaning || '暂无释义';
                                        const word = currentVocab?.content?.toLowerCase() || '';
                                        let filtered = meaning.replace(new RegExp(word, 'gi'), '').trim();
                                        filtered = filtered.replace(/\s{2,}/g, ' ').replace(/,\s*$/, '').replace(/^\s*,/, '').trim();
                                        return filtered || meaning;
                                    })()}
                                </div>

                                {/* 发音 */}
                                <div className="audio-row">
                                    {currentVocab?.phonetic && (
                                        <span className="phonetic-text">{currentVocab.phonetic}</span>
                                    )}
                                    <button className="audio-btn-small" onClick={playWordAudio}>
                                        <img src="/icon/laba.png" alt="播放" className="audio-icon-small" />
                                    </button>
                                </div>
                            </div>

                            {/* 下划线输入区域 - 类似句子默写 */}
                            <div
                                className="underscore-input-container"
                                onClick={() => inputRef.current?.focus()}
                            >
                                {/* 透明输入框 */}
                                <input
                                    ref={inputRef}
                                    type="text"
                                    className="underscore-hidden-input"
                                    value={userInput}
                                    onChange={(e) => setUserInput(e.target.value)}
                                    onKeyDown={(e) => {
                                        // 播放打字音效
                                        if (shouldPlayTypingSound(e.nativeEvent)) {
                                            playTypingSound();
                                        }
                                        
                                        const isPhrase = currentVocab?.content.trim().includes(' ');

                                        // 空格键处理
                                        if (e.key === ' ' && !showSpellResult && userInput.trim()) {
                                            if (isPhrase) {
                                                // 如果是短语/句子，允许输入空格，不自动提交
                                                // 只有当输入长度超过目标长度时，才可能考虑阻止（但一般不阻止）
                                            } else {
                                                // 如果是单个单词，空格键作为提交键
                                                e.preventDefault();
                                                autoJudgeAndSubmit();
                                            }
                                        }
                                        // Enter键总是触发提交
                                        if (e.key === 'Enter' && !showSpellResult && userInput.trim()) {
                                            e.preventDefault();
                                            autoJudgeAndSubmit();
                                        }
                                    }}
                                    disabled={showSpellResult}
                                    autoComplete="off"
                                    autoCapitalize="off"
                                    spellCheck="false"
                                />

                                {/* 下划线字母显示 */}
                                <div className={`underscore-display-review ${showSpellResult ? (checkSpellAnswer() ? 'correct' : 'incorrect') : ''}`}>
                                    {currentVocab && Array.from({ length: currentVocab.content.length }, (_, i) => {
                                        const targetChar = currentVocab.content[i];
                                        const inputChar = userInput[i] || '';
                                        const isCorrect = inputChar.toLowerCase() === targetChar.toLowerCase();
                                        const hasInput = inputChar !== '';

                                        return (
                                            <span
                                                key={i}
                                                className={`underscore-char-review ${
                                                    showSpellResult
                                                        ? (hasInput ? (isCorrect ? 'correct' : 'incorrect') : 'empty')
                                                        : (hasInput ? 'filled' : 'empty')
                                                } ${i === userInput.length ? 'cursor' : ''}`}
                                            >
                                                {showSpellResult ? targetChar : (inputChar || '')}
                                            </span>
                                        );
                                    })}
                                </div>

                                {/* 结果提示区 */}
                                {showSpellResult && (
                                    <div className="result-actions-container">
                                        <div className={`result-badge-review ${checkSpellAnswer() ? 'correct' : 'incorrect'}`}>
                                            {checkSpellAnswer() ? '✓ 正确' : '✗ 错误'}
                                        </div>
                                    </div>
                                )}
                            </div>

                            {/* 例句显示（如果有） */}
                            {currentVocab?.example && (
                                <div className="spell-example-hint">
                                    <span className="example-label">例句：</span>
                                    <span className="example-text">{currentVocab.example}</span>
                                </div>
                            )}

                            {/* Tab提示 */}
                            {!showSpellResult && (
                                <div className="tab-hint-section">
                                    {showHint && getHintInfo() ? (
                                        <div className="hint-revealed">
                                            <span className="hint-label-small">答案：</span>
                                            <span className="hint-word-small">{getHintInfo()?.word}</span>
                                        </div>
                                    ) : (
                                        <div className="hint-prompt">按 Tab 键查看答案</div>
                                    )}
                                </div>
                            )}
                        </div>

                        {/* 右侧 - 详细信息（提交后或按Tab后显示） */}
                        <div className={`spell-right ${(showSpellResult || showHint) ? 'visible' : 'hidden'}`}>
                            {(showSpellResult || showHint) ? (
                                <>
                                    <div className="info-section">
                                        <div className="section-header">
                                            <img src="/icon/fanyi.png" alt="" className="section-icon" />
                                            <span>释义</span>
                                        </div>
                                        <div className="info-content">{currentVocab?.meaning || '暂无释义'}</div>
                                    </div>

                                    {currentVocab?.example && (
                                        <div className="info-section">
                                            <div className="section-header">
                                                <img src="/icon/liju.png" alt="" className="section-icon" />
                                                <span>例句</span>
                                            </div>
                                            <div className="info-content example-text">{currentVocab.example}</div>
                                            {currentVocab.example_translation && (
                                                <div className="info-content example-zh">{currentVocab.example_translation}</div>
                                            )}
                                        </div>
                                    )}

                                    {morphemes.length > 0 && (
                                        <div className="info-section">
                                            <div className="section-header">
                                                <img src="/icon/cigencizhui.png" alt="" className="section-icon" />
                                                <span>词根词缀</span>
                                            </div>
                                            <div className="morpheme-tags">
                                                {morphemes.map((m: any, i: number) => (
                                                    <span key={i} className="morpheme-tag-spell">
                                                        <span className="tag-name">{m.morpheme}</span>
                                                        <span className="tag-meaning">{m.meaning}</span>
                                                    </span>
                                                ))}
                                            </div>
                                        </div>
                                    )}

                                    {currentVocab?.context && (
                                        <div className="info-section">
                                            <div className="section-header">
                                                <span className="context-icon">📖</span>
                                                <span>上下文</span>
                                            </div>
                                            <div className="info-content context-text">{currentVocab.context}</div>
                                        </div>
                                    )}
                                </>
                            ) : (
                                <div className="spell-right-placeholder">
                                    <div className="placeholder-icon">📝</div>
                                    <div className="placeholder-text">提交答案后显示详细信息</div>
                                </div>
                            )}
                        </div>
                    </div>
                )}
            </div>

            {/* 底部操作按钮 */}
            <div className={`desktop-footer ${isFullscreen ? 'compact' : ''}`}>
                <div className="action-buttons">
                    <div className="action-item">
                        <button
                            className={`action-btn forget ${vibratingButton === 'forget' ? 'vibrating' : ''}`}
                            onClick={() => handleSubmitReview('forget')}
                            disabled={submitting || (mode === 'spell' && !showSpellResult)}
                            data-key="←"
                        >
                            <span className="btn-label">忘记</span>
                        </button>
                        <span className="btn-desc">明天复习</span>
                    </div>
                    <div className="action-item">
                        <button
                            className={`action-btn fuzzy ${vibratingButton === 'fuzzy' ? 'vibrating' : ''}`}
                            onClick={() => handleSubmitReview('fuzzy')}
                            disabled={submitting || (mode === 'spell' && !showSpellResult)}
                            data-key="↑"
                        >
                            <span className="btn-label">模糊</span>
                        </button>
                        <span className="btn-desc">3天后复习</span>
                    </div>
                    <div className="action-item">
                        <button
                            className={`action-btn know ${vibratingButton === 'know' ? 'vibrating' : ''}`}
                            onClick={() => handleSubmitReview('know')}
                            disabled={submitting || (mode === 'spell' && !showSpellResult)}
                            data-key="→"
                        >
                            <span className="btn-label">认识</span>
                        </button>
                        <span className="btn-desc">一周后复习</span>
                    </div>
                </div>
            </div>
        </div>
        </>
    );
};

export default ReviewSession;
