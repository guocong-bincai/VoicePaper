import React, { useEffect, useRef, useState } from 'react';
import { useStore } from '../store/useStore';

export const AudioPlayer: React.FC = () => {
    const { currentArticle, isPlaying, setIsPlaying, currentTime, setCurrentTime, duration, setDuration } = useStore();
    const audioRef = useRef<HTMLAudioElement>(null);
    const progressSliderRef = useRef<HTMLInputElement>(null);
    const progressFillRef = useRef<HTMLDivElement>(null);
    const [isUserInteracting, setIsUserInteracting] = useState(false);

    useEffect(() => {
        if (currentArticle && audioRef.current && currentArticle.audio_path) {
            // 直接使用本地音频文件路径（相对于 public 目录）
            audioRef.current.src = `/${currentArticle.audio_path}`;
            audioRef.current.load();
            if (isPlaying) {
                audioRef.current.play().catch(e => console.error("Auto-play failed:", e));
            }
        }
    }, [currentArticle, isPlaying]);

    useEffect(() => {
        if (audioRef.current) {
            if (isPlaying) {
                audioRef.current.play().catch(e => {
                    console.error("Play failed:", e);
                    setIsPlaying(false);
                });
            } else {
                audioRef.current.pause();
            }
        }
    }, [isPlaying, setIsPlaying]);

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

    const handleLoadedMetadata = () => {
        if (audioRef.current) {
            const total = audioRef.current.duration;
            setDuration(total);
            // 初始化进度条
            if (progressFillRef.current) {
                progressFillRef.current.style.width = '0%';
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

    const formatTime = (time: number) => {
        if (isNaN(time)) return "00:00";
        const mins = Math.floor(time / 60);
        const secs = Math.floor(time % 60);
        return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    };

    if (!currentArticle || !currentArticle.audio_path) return null;

    return (
        <>
            <audio
                ref={audioRef}
                id="audioPlayer"
                onTimeUpdate={handleTimeUpdate}
                onLoadedMetadata={handleLoadedMetadata}
                onEnded={() => setIsPlaying(false)}
            />

            <div className="bottom-player">
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
                                disabled={currentArticle.status !== 'completed'}
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
                                    value={duration > 0 ? (currentTime / duration) * 100 : 0}
                                    step="0.1"
                                    onChange={(e) => {
                                        handleSeekInput(e);
                                        handleSeekChange(e);
                                    }}
                                    onMouseDown={handleSeekStart}
                                    onTouchStart={handleSeekStart}
                                    disabled={currentArticle.status !== 'completed'}
                                />
                                <div className="progress-fill" ref={progressFillRef}></div>
                            </div>
                            <span id="totalTime">{formatTime(duration)}</span>
                        </div>
                    </div>

                    <div className="player-extra">
                        <button className="icon-btn small">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                <path d="M3 18v-6a9 9 0 0 1 18 0v6"/>
                                <path d="M21 19a2 2 0 0 1-2 2h-1a2 2 0 0 1-2-2v-3a2 2 0 0 1 2-2h3zM3 19a2 2 0 0 0 2 2h1a2 2 0 0 0 2-2v-3a2 2 0 0 0-2-2H3z"/>
                            </svg>
                        </button>
                    </div>
                </div>
            </div>
        </>
    );
};
