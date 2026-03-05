import React, { useEffect, useRef, useState, useCallback } from 'react';
import { useStore } from '../store/useStore';
import { getReadingProgress, saveReadingProgress, quickSaveReadingProgress } from '../services/api';

export const AudioPlayer: React.FC = () => {
    const { currentArticle, isPlaying, setIsPlaying, currentTime, setCurrentTime, duration, setDuration, user } = useStore();
    const audioRef = useRef<HTMLAudioElement>(null);
    const progressSliderRef = useRef<HTMLInputElement>(null);
    const progressFillRef = useRef<HTMLDivElement>(null);
    const [isUserInteracting, setIsUserInteracting] = useState(false);
    const [hasRestoredProgress, setHasRestoredProgress] = useState(false);
    const [isMinimized, setIsMinimized] = useState(false); // 播放器折叠状态
    const lastSaveTimeRef = useRef<number>(0);
    const saveIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
    const lastArticleIdRef = useRef<number | null>(null);

    const [playbackRate, setPlaybackRate] = useState(1);
    const playbackRates = [0.75, 1, 1.25, 1.5, 2];

    const togglePlaybackRate = () => {
        const currentIndex = playbackRates.indexOf(playbackRate);
        const nextIndex = (currentIndex + 1) % playbackRates.length;
        setPlaybackRate(playbackRates[nextIndex]);
    };

    useEffect(() => {
        if (audioRef.current) {
            audioRef.current.playbackRate = playbackRate;
        }
    }, [playbackRate]);

    // 保存阅读进度
    const saveProgress = useCallback(async (final: boolean = false) => {
        if (!currentArticle || !user || !audioRef.current) return;
        
        const current = audioRef.current.currentTime;
        const total = audioRef.current.duration;
        
        if (isNaN(total) || total <= 0) return;

        try {
            if (final) {
                // 最终保存（暂停、切换文章、播放结束时）
                await saveReadingProgress({
                    article_id: currentArticle.id,
                    current_time: current,
                    duration: total
                });
                console.log('✅ 阅读进度已保存:', Math.round((current / total) * 100) + '%');
            } else {
                // 快速保存（定期自动保存）
                await quickSaveReadingProgress({
                    article_id: currentArticle.id,
                    current_time: current,
                    duration: total
                });
            }
            lastSaveTimeRef.current = current;
        } catch (error) {
            console.warn('⚠️ 保存阅读进度失败:', error);
        }
    }, [currentArticle, user]);

    // 恢复阅读进度
    const restoreProgress = useCallback(async () => {
        if (!currentArticle || !user || !audioRef.current) return;
        
        try {
            const progress = await getReadingProgress(currentArticle.id);
            if (progress && progress.current_time > 0 && progress.duration > 0) {
                // 如果进度小于95%，恢复到上次位置
                if (progress.progress < 95) {
                    audioRef.current.currentTime = progress.current_time;
                    setCurrentTime(progress.current_time);
                    console.log('✅ 已恢复阅读进度:', Math.round(progress.progress) + '%');
                }
            }
        } catch (error) {
            console.warn('⚠️ 获取阅读进度失败:', error);
        }
    }, [currentArticle, user, setCurrentTime]);

    // 当文章变化时，重置状态并恢复进度
    useEffect(() => {
        if (currentArticle && audioRef.current && currentArticle.audio_url) {
            // 检测文章是否真的变化了
            if (lastArticleIdRef.current !== currentArticle.id) {
                // 文章变化了，保存上一篇的进度
                if (lastArticleIdRef.current !== null && user) {
                    saveProgress(true);
                }
                
                lastArticleIdRef.current = currentArticle.id;
                setHasRestoredProgress(false);
                
                // 使用从API获取的audio_url
            const newSrc = currentArticle.audio_url;
                // 只有当 src 真正改变时才重新加载
                const currentSrc = audioRef.current.src;
                const newSrcBase = newSrc.split('?')[0];
                if (!currentSrc.endsWith(newSrcBase) && currentSrc !== newSrc) {
            audioRef.current.src = newSrc;
            audioRef.current.load();
            if (isPlaying) {
                audioRef.current.play().catch(e => console.error("Auto-play failed:", e));
            }
        }
            }
        }
    }, [currentArticle, user, saveProgress, isPlaying]);

    // 播放状态变化时
    useEffect(() => {
        if (audioRef.current) {
            if (isPlaying) {
                audioRef.current.play().catch(e => {
                    console.error("Play failed:", e);
                    setIsPlaying(false);
                });
            } else {
                audioRef.current.pause();
                // 暂停时保存进度
                if (user && currentArticle) {
                    saveProgress(true);
                }
            }
        }
    }, [isPlaying, setIsPlaying, user, currentArticle, saveProgress]);

    // 定期自动保存进度（每10秒）
    useEffect(() => {
        if (isPlaying && user && currentArticle) {
            saveIntervalRef.current = setInterval(() => {
                if (audioRef.current && !isNaN(audioRef.current.duration)) {
                    const timeDiff = audioRef.current.currentTime - lastSaveTimeRef.current;
                    // 只有当播放了超过5秒才保存
                    if (Math.abs(timeDiff) >= 5) {
                        saveProgress(false);
                    }
                }
            }, 10000); // 每10秒检查一次
        }

        return () => {
            if (saveIntervalRef.current) {
                clearInterval(saveIntervalRef.current);
                saveIntervalRef.current = null;
            }
        };
    }, [isPlaying, user, currentArticle, saveProgress]);

    // 页面卸载时保存进度
    useEffect(() => {
        const handleBeforeUnload = () => {
            if (user && currentArticle && audioRef.current) {
                // 使用 sendBeacon 保存进度（不会被页面关闭中断）
                const data = {
                    article_id: currentArticle.id,
                    current_time: audioRef.current.currentTime,
                    duration: audioRef.current.duration
                };
                const token = localStorage.getItem('token');
                if (token && !isNaN(data.duration) && data.duration > 0) {
                    navigator.sendBeacon(
                        '/api/v1/reading-progress/quick',
                        new Blob([JSON.stringify(data)], { type: 'application/json' })
                    );
                }
            }
        };

        window.addEventListener('beforeunload', handleBeforeUnload);
        return () => {
            window.removeEventListener('beforeunload', handleBeforeUnload);
        };
    }, [user, currentArticle]);

    const togglePlay = () => {
        setIsPlaying(!isPlaying);
    };

    const handleTimeUpdate = () => {
        if (audioRef.current && !isUserInteracting) {
            const current = audioRef.current.currentTime;
            const total = audioRef.current.duration;
            setCurrentTime(current);
            
            if (total > 0) {
                const percentage = (current / total) * 100;
                if (progressSliderRef.current) {
                    progressSliderRef.current.value = percentage.toString();
                }
                if (progressFillRef.current) {
                    progressFillRef.current.style.width = `${percentage}%`;
                }
            }
        }
    };

    const handleLoadedMetadata = async () => {
        if (audioRef.current) {
            const total = audioRef.current.duration;
            setDuration(total);
            // 初始化进度条
            if (progressFillRef.current) {
                progressFillRef.current.style.width = '0%';
            }
            
            // 音频加载完成后恢复进度
            if (!hasRestoredProgress && user && currentArticle) {
                setHasRestoredProgress(true);
                await restoreProgress();
            }
        }
    };

    const handleSeekInput = (e: React.ChangeEvent<HTMLInputElement>) => {
        const percentage = parseFloat(e.target.value);
        if (audioRef.current && duration > 0) {
            const time = (percentage / 100) * duration;
            if (progressFillRef.current) {
                progressFillRef.current.style.width = `${percentage}%`;
            }
            setCurrentTime(time);
        }
    };

    const handleSeekChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const percentage = parseFloat(e.target.value);
        if (audioRef.current && duration > 0) {
            const time = (percentage / 100) * duration;
            audioRef.current.currentTime = time;
            setCurrentTime(time);
        }
        setIsUserInteracting(false);
    };

    const handleSeekStart = () => {
        setIsUserInteracting(true);
    };

    const skip = (seconds: number) => {
        if (audioRef.current) {
            const newTime = Math.max(0, Math.min(audioRef.current.currentTime + seconds, duration));
            audioRef.current.currentTime = newTime;
            setCurrentTime(newTime);
        }
    };

    const handleEnded = () => {
        setIsPlaying(false);
        // 播放结束时保存进度（标记为100%完成）
        if (user && currentArticle) {
            saveProgress(true);
        }
    };

    const formatTime = (time: number) => {
        if (isNaN(time)) return "00:00";
        const mins = Math.floor(time / 60);
        const secs = Math.floor(time % 60);
        return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    };

    if (!currentArticle || !currentArticle.audio_url) return null;

    const progressPercent = duration > 0 ? (currentTime / duration) * 100 : 0;

    return (
        <>
            <audio
                ref={audioRef}
                id="audioPlayer"
                onTimeUpdate={handleTimeUpdate}
                onLoadedMetadata={handleLoadedMetadata}
                onEnded={handleEnded}
            />

            {/* 迷你播放器 - 折叠状态 */}
            <div className={`mini-player ${isMinimized ? 'visible' : ''}`}>
                <div className="mini-player-content">
                    {/* 迷你进度环 */}
                    <div className="mini-progress-ring">
                        <svg viewBox="0 0 44 44">
                            <circle className="ring-bg" cx="22" cy="22" r="18" />
                            <circle 
                                className="ring-progress" 
                                cx="22" cy="22" r="18"
                                style={{ 
                                    strokeDasharray: `${2 * Math.PI * 18}`,
                                    strokeDashoffset: `${2 * Math.PI * 18 * (1 - progressPercent / 100)}`
                                }}
                            />
                        </svg>
                        <button
                            type="button"
                            onClick={togglePlay}
                            className="mini-play-btn"
                            title="播放/暂停"
                            disabled={currentArticle.online !== '1'}
                        >
                            <svg viewBox="0 0 24 24" fill="currentColor" style={{ display: isPlaying ? 'none' : 'block' }}>
                                <path d="M8 5v14l11-7z"/>
                            </svg>
                            <svg viewBox="0 0 24 24" fill="currentColor" style={{ display: isPlaying ? 'block' : 'none' }}>
                                <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
                            </svg>
                        </button>
                    </div>
                    {/* 展开按钮 */}
                    <button 
                        className="mini-expand-btn"
                        onClick={() => setIsMinimized(false)}
                        title="展开播放器"
                    >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <path d="M7 14l5-5 5 5"/>
                        </svg>
                    </button>
                </div>
            </div>

            {/* 完整播放器 - 展开状态 */}
            <div className={`bottom-player ${isMinimized ? 'minimized' : ''}`}>
                <div className="player-content">
                    <div className="track-info">
                        <div className="track-icon">🎧</div>
                        <div className="track-text">
                            <div className="track-title">{currentArticle.title}</div>
                            <div className="track-status">{isPlaying ? '正在朗读...' : '点击播放'}</div>
                        </div>
                    </div>

                    <div className="player-controls-wrapper">
                        <div className="controls-main">
                            <button
                                type="button"
                                onClick={() => skip(-10)}
                                className="icon-btn"
                                title="后退10秒"
                            >
                                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                    <path d="M11 17l-5-5 5-5M18 17l-5-5 5-5"/>
                                </svg>
                            </button>

                            <button
                                type="button"
                                onClick={togglePlay}
                                className="play-btn"
                                title="播放/暂停"
                                disabled={currentArticle.online !== '1'}
                            >
                                <svg className="play-icon" viewBox="0 0 24 24" fill="currentColor" style={{ display: isPlaying ? 'none' : 'block' }}>
                                    <path d="M8 5v14l11-7z"/>
                                </svg>
                                <svg className="pause-icon" viewBox="0 0 24 24" fill="currentColor" style={{ display: isPlaying ? 'block' : 'none' }}>
                                    <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
                                </svg>
                            </button>

                            <button
                                type="button"
                                onClick={() => skip(10)}
                                className="icon-btn"
                                title="前进10秒"
                            >
                                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                    <path d="M13 17l5-5-5-5M6 17l5-5-5-5"/>
                                </svg>
                            </button>
                        </div>

                        <div className="progress-wrapper">
                            <span id="currentTime">{formatTime(currentTime)}</span>
                            <div className="progress-bar-container">
                                <input
                                    type="range"
                                    id="progressSlider"
                                    ref={progressSliderRef}
                                    min="0"
                                    max="100"
                                    value={progressPercent}
                                    step="0.1"
                                    onChange={(e) => {
                                        handleSeekInput(e);
                                        handleSeekChange(e);
                                    }}
                                    onMouseDown={handleSeekStart}
                                    onTouchStart={handleSeekStart}
                                    disabled={currentArticle.online !== '1'}
                                />
                                <div className="progress-fill" ref={progressFillRef}></div>
                            </div>
                            <span id="totalTime">{formatTime(duration)}</span>
                        </div>
                    </div>

                    <div className="player-extra">
                        {/* 倍速按钮 */}
                        <button
                            className="icon-btn speed-btn"
                            onClick={togglePlaybackRate}
                            title="切换倍速"
                            style={{ fontSize: '0.8rem', fontWeight: 'bold', width: 'auto', minWidth: '36px' }}
                        >
                            {playbackRate}x
                        </button>

                        {/* 收起按钮 */}
                        <button 
                            className="icon-btn minimize-btn"
                            onClick={() => setIsMinimized(true)}
                            title="收起播放器"
                        >
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                <path d="M7 10l5 5 5-5"/>
                            </svg>
                        </button>
                    </div>
                </div>
            </div>
        </>
    );
};
