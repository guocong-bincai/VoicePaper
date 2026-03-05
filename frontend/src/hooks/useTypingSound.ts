import { useRef } from 'react';

// 音效路径
const typingSoundPath = '/sounds/typing.mp3';
const rightSoundPath = '/sounds/right.mp3';
const errorSoundPath = '/sounds/error.mp3';

// 正确和错误音效使用简单的 Audio 对象
const rightAudio = new Audio(rightSoundPath);
const errorAudio = new Audio(errorSoundPath);

// 打字音效使用 Web Audio API，性能更好
const PLAY_INTERVAL_TIME = 60; // 60ms 节流，避免过于频繁
let audioCtxRef: AudioContext | null = null;
let audioBuffer: AudioBuffer | null = null;

/**
 * 提前加载音频上下文和缓冲区
 */
async function loadAudioContext() {
  if (audioCtxRef) return;
  
  audioCtxRef = new AudioContext();
  await loadAudioBuffer(typingSoundPath);
}

async function loadAudioBuffer(url: string) {
  try {
    const response = await fetch(url);
    const arrayBuffer = await response.arrayBuffer();
    
    if (!audioCtxRef) return;
    
    const decodedAudioData = await audioCtxRef.decodeAudioData(arrayBuffer);
    if (decodedAudioData) {
      audioBuffer = decodedAudioData;
    }
  } catch (error) {
    console.warn('加载音效失败:', error);
  }
}

/**
 * 打字音效 Hook
 */
export function useTypingSound() {
  const lastPlayTimeRef = useRef(0);

  // 提前加载音频上下文
  if (!audioCtxRef) {
    loadAudioContext();
  }

  /**
   * 播放打字音效
   */
  const playTypingSound = () => {
    const now = Date.now();
    // 节流：如果距离上次播放不到 60ms，则跳过
    if (now - lastPlayTimeRef.current < PLAY_INTERVAL_TIME) return;
    if (!audioCtxRef || !audioBuffer) return;

    try {
      const source = audioCtxRef.createBufferSource();
      source.buffer = audioBuffer;
      source.connect(audioCtxRef.destination);
      source.start();
      lastPlayTimeRef.current = now;
      
      // 播放结束后释放资源
      source.onended = () => {
        source.disconnect();
      };
    } catch (error) {
      console.warn('播放打字音效失败:', error);
    }
  };

  /**
   * 检查是否应该播放打字音效
   */
  const shouldPlayTypingSound = (e: KeyboardEvent): boolean => {
    // 忽略组合键
    if (e.altKey || e.ctrlKey || e.metaKey) return false;

    // 只在输入字母、数字、空格、Backspace、撇号时播放
    if (/^[a-zA-Z0-9]$/.test(e.key) || ['Backspace', ' ', "'"].includes(e.key)) {
      return true;
    }

    return false;
  };

  /**
   * 播放正确音效
   */
  const playRightSound = () => {
    try {
      rightAudio.currentTime = 0; // 重置播放位置
      rightAudio.play();
    } catch (error) {
      console.warn('播放正确音效失败:', error);
    }
  };

  /**
   * 播放错误音效
   */
  const playErrorSound = () => {
    try {
      errorAudio.currentTime = 0; // 重置播放位置
      errorAudio.play();
    } catch (error) {
      console.warn('播放错误音效失败:', error);
    }
  };

  return {
    playTypingSound,
    shouldPlayTypingSound,
    playRightSound,
    playErrorSound,
  };
}

