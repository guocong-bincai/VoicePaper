import React, { useState, useEffect, useRef, useLayoutEffect, useCallback } from 'react';
import { useStore } from '../store/useStore';
import { getWords, getSentences, saveDictationRecord, getArticle, getDictationProgress, saveDictationProgress } from '../services/api';
import type { Word, Sentence } from '../types';
import { useTypingSound } from '../hooks/useTypingSound';
import confetti from 'canvas-confetti';
import { AddVocabularyPopup } from './AddVocabularyPopup';
import { FloatingBackButton } from './FloatingBackButton';
import './DictationPage.css';

// 听写内容缓存
const dictationDataCache: Record<number, { words: Word[], sentences: Sentence[] }> = {};

interface DictationPageProps {
  articleId: number;
  onBack: () => void; // 返回上一层级
}

type PracticeMode = 'word' | 'sentence';
type PracticeState = 'idle' | 'listening' | 'typing' | 'checking' | 'result';

// 功能实现: 参考earthworm实现三种输入模式
// 实现方案: Input(正常输入) → Fix(等待修复) → Fix_Input(修复输入)
// 影响范围: frontend/src/components/DictationPage.tsx
// 实现日期: 2025-12-10
// 注：InputMode类型定义预留给未来earthworm风格优化
// const InputMode = {
//   Input: 'input',
//   Fix: 'fix',
//   Fix_Input: 'fix-input'
// } as const;
// type InputModeType = typeof InputMode[keyof typeof InputMode];

export const DictationPage: React.FC<DictationPageProps> = ({ articleId, onBack }) => {
  const { theme, setTheme } = useStore();
  const [isFullscreen, setIsFullscreen] = useState(false);

  // 进入全屏模式
  const enterFullscreen = useCallback(() => {
    console.log('📺 enterFullscreen called');
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
    console.log('📺 exitFullscreen called');
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

  // 立即设置背景色，不等useEffect - 最激进的方法
  useLayoutEffect(() => {
    const bgColor = theme === 'dark' ? '#1F2937' : '#F9F8F6';
    console.log('🎨 DictationPage: 设置背景色为', bgColor);

    // 立即强制设置所有可能的父元素背景色
    document.body.style.setProperty('background', bgColor, 'important');
    document.body.style.setProperty('background-color', bgColor, 'important');
    document.documentElement.style.setProperty('background', bgColor, 'important');
    document.documentElement.style.setProperty('background-color', bgColor, 'important');

    const rootElement = document.getElementById('root');
    if (rootElement) {
      rootElement.style.setProperty('background', bgColor, 'important');
      rootElement.style.setProperty('background-color', bgColor, 'important');
    }
  }, [theme]);

  const [mode, setMode] = useState<PracticeMode>('sentence'); // 默认句子模式（如果没有单词就只显示句子）
  const [words, setWords] = useState<Word[]>(dictationDataCache[articleId]?.words || []);
  const [sentences, setSentences] = useState<Sentence[]>(dictationDataCache[articleId]?.sentences || []);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [userInput, setUserInput] = useState('');
  const [practiceState, setPracticeState] = useState<PracticeState>('idle');
  const [isCorrect, setIsCorrect] = useState<boolean | null>(null);
  const [score, setScore] = useState(0);
  const [attempts, setAttempts] = useState(0);
  const [showAnswer, setShowAnswer] = useState(false);
  const [startTime, setStartTime] = useState<number | null>(null);
  const [showHint, setShowHint] = useState(false); // 显示提示信息
  const [articleTitle, setArticleTitle] = useState<string>(''); // 文章标题
  const [wordErrors, setWordErrors] = useState<boolean[]>([]); // 每个单词的错误状态
  const [userWordsArray, setUserWordsArray] = useState<string[]>([]); // 句子模式：每个单词的输入数组
  const [dataLoaded, setDataLoaded] = useState(!!dictationDataCache[articleId]); // 数据是否加载完成 (如果有缓存则初始为true)
  const [initialDataLoaded, setInitialDataLoaded] = useState(!!dictationDataCache[articleId]); // 初始数据（单词+句子）是否加载完成

  // 保留原有的修复模式状态（兼容现有代码）
  const [isFixingErrors, setIsFixingErrors] = useState(false); // 是否在修复错误模式
  const [currentFixingIndex, setCurrentFixingIndex] = useState(-1); // 当前正在修复的单词索引
  const [fixingInput, setFixingInput] = useState(''); // 修复时单个单词的输入

  // 功能实现: 生词本弹窗状态
  // 实现方案: 点击生词本按钮时显示AddVocabularyPopup弹窗
  // 影响范围: frontend/src/components/DictationPage.tsx
  // 实现日期: 2025-01-27
  const [showVocabularyPopup, setShowVocabularyPopup] = useState(false);

  // 功能实现: 新的状态管理，基于earthworm三种模式
  // 注：以下状态暂未使用，预留给未来earthworm风格优化
  // const [inputMode, setInputMode] = useState('input');
  // const [currentWordIndex, setCurrentWordIndex] = useState(0);
  // const [currentEditWordIndex, setCurrentEditWordIndex] = useState(-1);

  // const audioRef = useRef<HTMLAudioElement>(null); // 暂未使用
  const successAudioRef = useRef<HTMLAudioElement>(null);
  const errorAudioRef = useRef<HTMLAudioElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // 打字音效 Hook
  const {
    playTypingSound,
    shouldPlayTypingSound,
    playRightSound: playCorrectSoundEffect,
    playErrorSound: playErrorSoundEffect
  } = useTypingSound();

  // 强制设置背景色，彻底消除黑边
  useEffect(() => {
    const bgColor = theme === 'dark' ? '#1F2937' : '#F9F8F6';

    // 保存原始样式
    const originalStyles = {
      bodyBg: document.body.style.background,
      bodyBgColor: document.body.style.backgroundColor,
      htmlBg: document.documentElement.style.background,
      htmlBgColor: document.documentElement.style.backgroundColor,
      bodyOverflow: document.body.style.overflow,
      htmlOverflow: document.documentElement.style.overflow,
      bodyMargin: document.body.style.margin,
      bodyPadding: document.body.style.padding,
      htmlMargin: document.documentElement.style.margin,
      htmlPadding: document.documentElement.style.padding,
    };

    // 强制设置 body 和 html 的背景色
    document.body.style.setProperty('background', bgColor, 'important');
    document.body.style.setProperty('background-color', bgColor, 'important');
    document.documentElement.style.setProperty('background', bgColor, 'important');
    document.documentElement.style.setProperty('background-color', bgColor, 'important');

    // 设置其他必要样式
    document.body.style.overflow = 'hidden';
    document.documentElement.style.overflow = 'hidden';
    document.body.style.margin = '0';
    document.body.style.padding = '0';
    document.documentElement.style.margin = '0';
    document.documentElement.style.padding = '0';

    // 强制设置 #root 的背景色
    const rootElement = document.getElementById('root');
    let originalRootStyles: any = {};

    if (rootElement) {
      originalRootStyles = {
        bg: rootElement.style.background,
        bgColor: rootElement.style.backgroundColor,
        maxWidth: rootElement.style.maxWidth,
        margin: rootElement.style.margin,
        padding: rootElement.style.padding,
      };

      rootElement.style.setProperty('background', bgColor, 'important');
      rootElement.style.setProperty('background-color', bgColor, 'important');
      rootElement.style.maxWidth = '100vw';
      rootElement.style.margin = '0';
      rootElement.style.padding = '0';
    }

    return () => {
      // 恢复原始样式
      document.body.style.background = originalStyles.bodyBg;
      document.body.style.backgroundColor = originalStyles.bodyBgColor;
      document.documentElement.style.background = originalStyles.htmlBg;
      document.documentElement.style.backgroundColor = originalStyles.htmlBgColor;
      document.body.style.overflow = originalStyles.bodyOverflow;
      document.documentElement.style.overflow = originalStyles.htmlOverflow;
      document.body.style.margin = originalStyles.bodyMargin;
      document.body.style.padding = originalStyles.bodyPadding;
      document.documentElement.style.margin = originalStyles.htmlMargin;
      document.documentElement.style.padding = originalStyles.htmlPadding;

      if (rootElement) {
        rootElement.style.background = originalRootStyles.bg || '';
        rootElement.style.backgroundColor = originalRootStyles.bgColor || '';
        rootElement.style.maxWidth = originalRootStyles.maxWidth || '';
        rootElement.style.margin = originalRootStyles.margin || '';
        rootElement.style.padding = originalRootStyles.padding || '';
      }
    };
  }, [theme]);

  // 加载文章标题
  useEffect(() => {
    const loadArticleTitle = async () => {
      try {
        const article = await getArticle(articleId);
        setArticleTitle(article.title);
      } catch (error) {
        console.error('加载文章标题失败:', error);
      }
    };
    loadArticleTitle();
  }, [articleId]);

  // 加载进度数据（在初始数据加载完成后）
  useEffect(() => {
    if (initialDataLoaded) {
    loadData();
    }
  }, [articleId, mode, initialDataLoaded]); // eslint-disable-line react-hooks/exhaustive-deps

  // 智能模式切换：数据加载完成后，如果当前模式没有数据，自动切换到有数据的模式
  useEffect(() => {
    if (dataLoaded) {
      const hasWords = (words?.length || 0) > 0;
      const hasSentences = (sentences?.length || 0) > 0;

      if (mode === 'word' && !hasWords && hasSentences) {
        // 单词模式但没有单词，切换到句子模式
        setMode('sentence');
      } else if (mode === 'sentence' && !hasSentences && hasWords) {
        // 句子模式但没有句子，切换到单词模式
        setMode('word');
      }
    }
  }, [dataLoaded, words, sentences]); // eslint-disable-line react-hooks/exhaustive-deps

  // 自动开始朗读 - 当数据加载完成且处于idle状态时
  useEffect(() => {
    if (practiceState === 'idle') {
      // 检查是否有可用的单词或句子（添加空值检查）
      const hasData = mode === 'word' ? (words?.length || 0) > 0 : (sentences?.length || 0) > 0;
      if (hasData) {
        // 延迟一点点，让UI先渲染
        const timer = setTimeout(() => {
          playPronunciation();
        }, 100);
        return () => clearTimeout(timer);
      }
    }
  }, [words, sentences, currentIndex, practiceState, mode]); // eslint-disable-line react-hooks/exhaustive-deps

  // 快捷键支持：空格键播放声音、Tab键显示提示、Ctrl+切换题目
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Esc键：重新开始（清空当前输入）- 输入状态可用
      if (e.code === 'Escape' && practiceState === 'typing' && !showAnswer) {
        e.preventDefault();
        setUserInput('');
        setFixingInput('');
        setUserWordsArray([]);
        setWordErrors([]);
        setIsFixingErrors(false);
        setCurrentFixingIndex(-1);
        inputRef.current?.focus();
      }

      // Ctrl + / : 播放发音（所有模式）
      if (e.ctrlKey && e.code === 'Slash') {
        e.preventDefault();
        if (!showAnswer || (showAnswer && isCorrect)) {
          playPronunciation();
        }
      }

      // Tab键切换提示显示（typing状态时都可用）
      if (e.code === 'Tab' && practiceState === 'typing') {
        e.preventDefault();
        setShowHint(prev => !prev);
      }

      // R键：答错后重新再来
      if (e.code === 'KeyR' && showAnswer && !isCorrect && !isFixingErrors) {
        e.preventDefault();
        retryItem();
      }

      // Ctrl + , : 上一题（任何状态，但不能在修复错误时）
      if (e.ctrlKey && e.code === 'Comma' && !isFixingErrors && currentIndex > 0) {
        e.preventDefault();
        previousItem();
      }

      // Ctrl + . : 下一题（任何状态，但不能在修复错误时）
      if (e.ctrlKey && e.code === 'Period' && !isFixingErrors) {
        e.preventDefault();
        nextItem();
      }

      // 左箭头键：上一题（idle状态时可用）
      if (e.code === 'ArrowLeft' && practiceState === 'idle' && currentIndex > 0) {
        e.preventDefault();
        previousItem();
      }

      // 右箭头键：下一题（idle状态或答对后可用）
      if (e.code === 'ArrowRight' && (practiceState === 'idle' || (showAnswer && isCorrect))) {
        e.preventDefault();
        nextItem();
      }

      // 功能实现: Ctrl + S 快捷键打开生词本弹窗
      // 实现日期: 2025-01-27
      if (e.ctrlKey && e.code === 'KeyS') {
        e.preventDefault();
        setShowVocabularyPopup(true);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [practiceState, showAnswer, isCorrect, currentIndex, isFixingErrors]); // eslint-disable-line react-hooks/exhaustive-deps

  // BUG修复: 初始化时同时加载单词和句子数据，用于正确显示按钮
  // 修复日期: 2025-12-11
  // 初始加载：同时获取单词和句子数据
  useEffect(() => {
    const loadInitialData = async () => {
      try {
        console.log('📥 开始加载单词和句子数据...');
        const [wordsData, sentencesData] = await Promise.all([
          getWords(articleId),
          getSentences(articleId)
        ]);

        console.log('获取到的单词数据:', wordsData);
        console.log('获取到的句子数据:', sentencesData);

        setWords(wordsData || []);
        setSentences(sentencesData || []);
        
        // 更新缓存
        dictationDataCache[articleId] = {
          words: wordsData || [],
          sentences: sentencesData || []
        };

        // 根据数据决定默认模式：有单词优先单词，否则句子
        const hasWords = (wordsData?.length || 0) > 0;
        const hasSentences = (sentencesData?.length || 0) > 0;

        if (hasWords) {
          setMode('word');
          console.log('✅ 默认单词模式');
        } else if (hasSentences) {
          setMode('sentence');
          console.log('✅ 默认句子模式');
        }

        setInitialDataLoaded(true);
      } catch (error) {
        console.error('❌ 加载初始数据失败:', error);
        setInitialDataLoaded(true);
      }
    };

    loadInitialData();
  }, [articleId]);

  const loadData = async () => {
    try {
      if (!dataLoaded) setDataLoaded(false); // 只有未加载时才显示loading

      // 加载当前模式的进度
      if (mode === 'word') {
        // 尝试加载学习进度
        try {
          const progress = await getDictationProgress(articleId, 'word');
          console.log('✅ 获取到的进度数据:', progress);

          if (progress && !progress.completed && progress.current_index > 0) {
            // 从上次的位置继续
            setCurrentIndex(progress.current_index);
            setScore(progress.score || 0);
            console.log(`📍 从第 ${progress.current_index + 1} 个单词继续学习`);
          } else {
            // 从头开始
            setCurrentIndex(0);
            setScore(0);
            console.log('🆕 从头开始学习');
          }
        } catch (error: any) {
          console.log('⚠️ 未找到学习进度，从头开始');
          if (error.response?.status === 401) {
            console.log('💡 提示：登录后可以保存学习进度');
          } else {
            console.error('❌ 获取进度失败:', error);
          }
          setCurrentIndex(0);
          setScore(0);
        }
      } else {
        // 尝试加载学习进度
        try {
          const progress = await getDictationProgress(articleId, 'sentence');
          console.log('✅ 获取到的进度数据:', progress);

          if (progress && !progress.completed && progress.current_index > 0) {
            // 从上次的位置继续
            setCurrentIndex(progress.current_index);
            setScore(progress.score || 0);
            console.log(`📍 从第 ${progress.current_index + 1} 个句子继续学习`);
          } else {
            // 从头开始
            setCurrentIndex(0);
            setScore(0);
            console.log('🆕 从头开始学习');
          }
        } catch (error: any) {
          console.log('⚠️ 未找到学习进度，从头开始');
          if (error.response?.status === 401) {
            console.log('💡 提示：登录后可以保存学习进度');
          } else {
            console.error('❌ 获取进度失败:', error);
          }
          setCurrentIndex(0);
          setScore(0);
        }
      }
      setUserInput('');
      setPracticeState('idle');
      setIsCorrect(null);
      setAttempts(0);
      setShowAnswer(false);
      setDataLoaded(true); // 数据加载完成
    } catch (error) {
      console.error('加载数据失败:', error);
      setDataLoaded(false); // 加载失败，保持 false
    }
  };

  // 获取当前练习项
  const getCurrentItem = (): Word | Sentence | null => {
    // BUG修复: 添加数组有效性检查
    // 修复策略: 在访问数组前先检查数组是否存在且有元素
    // 影响范围: frontend/src/components/DictationPage.tsx:314-320
    // 修复日期: 2025-12-08
    if (mode === 'word') {
      if (!words || words.length === 0) return null;
      return words[currentIndex] || null;
    } else {
      if (!sentences || sentences.length === 0) return null;
      return sentences[currentIndex] || null;
    }
  };

  // 功能实现: earthworm风格的辅助函数
  // 实现日期: 2025-12-10

  // 获取句子的单词数组（只包含字母数字的单词）
  const getAnswerWords = (sentence: string): string[] => {
    return sentence.split(' ').filter(w => /[a-zA-Z0-9]/.test(w));
  };

  // 查找上一个错误单词的索引（从currentFixingIndex之前）
  const findPrevErrorIndex = (): number => {
    for (let i = currentFixingIndex - 1; i >= 0; i--) {
      if (wordErrors[i]) return i;
    }
    return -1;
  };

  // 判断当前单词是否为空
  const isCurrentWordEmpty = (): boolean => {
    return !userWordsArray[currentFixingIndex]?.trim();
  };

  // BUG修复: 判断当前是否在最后一个单词（用于判断空格是否提交）
  // 修复策略: 只有当用户输入的单词数量已等于答案单词数量时，空格才触发提交
  // 影响范围: frontend/src/components/DictationPage.tsx
  // 修复日期: 2025-12-11
  const isLastWord = (): boolean => {
    const currentItem = getCurrentItem();
    if (!currentItem || mode !== 'sentence') return false;
    // 只有当用户已经在输入最后一个单词时，空格才提交
    const answerWords = getAnswerWords(currentItem.text);
    const userWords = userInput.trim().split(/\s+/).filter(w => w.trim());
    return userWords.length >= answerWords.length;
  };

  // 播放发音（使用Web Speech API或音频文件）- 只播放单词
  const playPronunciation = () => {
    const item = getCurrentItem();
    if (!item) return;

    // 直接进入输入状态，不再展示单独的listening页面
    setPracticeState('typing');
    setStartTime(Date.now()); // 记录开始时间
    
    // 确保输入框获得焦点
    setTimeout(() => {
      if (inputRef.current) {
        inputRef.current.focus();
      }
    }, 100);

    // 使用Web Speech API播放发音
    if ('speechSynthesis' in window) {
      // 先停止之前的播放
      window.speechSynthesis.cancel();
      
      const wordUtterance = new SpeechSynthesisUtterance(item.text);
      wordUtterance.lang = 'en-US';
      wordUtterance.rate = 0.8; // 稍慢一点，便于听写

      window.speechSynthesis.speak(wordUtterance);
    }
  };

  // 播放例句
  const playExample = (exampleText: string) => {
    if ('speechSynthesis' in window) {
      const exampleUtterance = new SpeechSynthesisUtterance(exampleText);
      exampleUtterance.lang = 'en-US';
      exampleUtterance.rate = 0.8;
      window.speechSynthesis.speak(exampleUtterance);
    }
  };

  // 检查答案
  const checkAnswer = async (inputOverride?: string) => {
    const item = getCurrentItem();
    const actualInput = inputOverride !== undefined ? inputOverride : userInput;
    if (!item || !actualInput.trim()) return;

    setPracticeState('checking');

    // 如果是修复模式重新提交，不增加尝试次数
    const currentAttempts = isFixingErrors ? attempts : attempts + 1;
    if (!isFixingErrors) {
      setAttempts(currentAttempts);
    }

    // 计算花费时间（秒）
    const timeSpent = startTime ? Math.floor((Date.now() - startTime) / 1000) : 0;

    let correct = false;
    let calculatedScore = 0;
    let errors: boolean[] = []; // 定义在外部作用域

    if (mode === 'sentence') {
      // 句子模式：逐个单词比对
      const answerWords = item.text.split(' ').filter(w => /[a-zA-Z0-9]/.test(w));
      const userWords = actualInput.trim().split(/\s+/).filter(w => w.trim()); // 过滤空字符串

      console.log('=== 验证答案 ===');
      console.log('标准答案单词:', answerWords);
      console.log('用户输入单词:', userWords);
      console.log('是否修复模式:', isFixingErrors);

      let correctCount = 0;

      answerWords.forEach((answerWord, index) => {
        const userWord = userWords[index] || '';
        // BUG修复: 使用统一的标准化逻辑（只保留字母和数字）
        const normalizedAnswer = answerWord.toLowerCase().replace(/[^a-z0-9]/g, '');
        const normalizedUser = userWord.toLowerCase().replace(/[^a-z0-9]/g, '');

        const isCorrect = normalizedAnswer === normalizedUser;
        errors.push(!isCorrect);
        if (isCorrect) correctCount++;

        console.log(`单词${index}: "${answerWord}" vs "${userWord}" => ${isCorrect ? '✓' : '✗'} (normalized: "${normalizedAnswer}" vs "${normalizedUser}")`);
      });

      setWordErrors(errors);
      correct = errors.every(e => !e);
      calculatedScore = correct ? 100 : Math.max(0, Math.floor((correctCount / answerWords.length) * 100));

      console.log('错误数组:', errors);
      console.log('是否全部正确:', correct);
      console.log('==================');
    } else {
      // 单词模式：整体比较
      const normalizedInput = actualInput.trim().toLowerCase().replace(/[.,!?;:]/g, '');
      const normalizedAnswer = item.text.toLowerCase().replace(/[.,!?;:]/g, '');
      correct = normalizedInput === normalizedAnswer;
      calculatedScore = correct ? 100 : Math.max(0, 100 - (currentAttempts - 1) * 20);
    }

    setIsCorrect(correct);
    setShowAnswer(true);

    // 保存记录到后端
    try {
      const recordData: any = {
        article_id: articleId,
        dictation_type: mode,
        user_answer: actualInput.trim(),
        is_correct: correct,
        score: calculatedScore,
        attempt_count: currentAttempts,
        time_spent: timeSpent,
      };

      if (mode === 'word') {
        recordData.word_id = (item as Word).id;
      } else {
        recordData.sentence_id = (item as Sentence).id;
      }

      await saveDictationRecord(recordData);
    } catch (error) {
      console.error('保存默写记录失败:', error);
      // 即使保存失败，也不影响用户体验
    }

    if (correct) {
      // 播放成功音效
      playSuccessSound();
      setScore(prev => prev + 1);

      // 延迟后进入下一题
      setTimeout(() => {
        nextItem();
      }, 1500);
    } else {
      // 播放错误音效
      playErrorSound();

      // 句子模式：进入修复模式（仅在首次错误时）
      if (mode === 'sentence' && !isFixingErrors) {
        setShowAnswer(true);
        setIsFixingErrors(true);

        // 找到第一个错误单词的索引
        const firstErrorIndex = errors.findIndex(e => e);
        if (firstErrorIndex !== -1) {
          setCurrentFixingIndex(firstErrorIndex);

          // 初始化单词数组（用数组保存每个单词的输入）
          const answerWords = item.text.split(' ').filter(w => /[a-zA-Z0-9]/.test(w));
          const userWords = actualInput.trim().split(/\s+/);

          // 确保数组长度匹配
          const wordsArray = answerWords.map((_, index) => userWords[index] || '');
          setUserWordsArray(wordsArray);

          setFixingInput(''); // 清空单个单词的输入
          setPracticeState('typing');
        }
      } else if (mode === 'sentence' && isFixingErrors) {
        // 修复后仍然有错误，继续修复
        const firstErrorIndex = errors.findIndex(e => e);
        if (firstErrorIndex !== -1) {
          // 重新初始化单词数组（用当前的 actualInput）
          const answerWords = item.text.split(' ').filter(w => /[a-zA-Z0-9]/.test(w));
          const userWords = actualInput.trim().split(/\s+/);
          const wordsArray = answerWords.map((_, index) => userWords[index] || '');
          setUserWordsArray(wordsArray);

          setCurrentFixingIndex(firstErrorIndex);
          setFixingInput('');
          setPracticeState('typing');
        }
      }
    }
  };

  // 上一题
  const previousItem = async () => {
    const totalItems = mode === 'word' ? (words?.length || 0) : (sentences?.length || 0);

    if (currentIndex > 0) {
      const prevIndex = currentIndex - 1;

      // 保存进度
      try {
        await saveDictationProgress({
          article_id: articleId,
          dictation_type: mode,
          current_index: prevIndex,
          total_items: totalItems,
          score: score,
          completed: false,
        });
        console.log('✅ 进度已保存:', { 当前索引: prevIndex, 总数: totalItems, 得分: score });
      } catch (error: any) {
        if (error.response?.status === 401) {
          console.log('⚠️ 未登录，无法保存进度');
        } else {
          console.error('❌ 保存进度失败:', error);
        }
      }

      setCurrentIndex(prevIndex);
      setUserInput('');
      setPracticeState('idle');
      setIsCorrect(null);
      setShowAnswer(false);
      setShowHint(false);
      setAttempts(0);
      setWordErrors([]);
      setIsFixingErrors(false);
      setCurrentFixingIndex(-1);
      setFixingInput('');
      setUserWordsArray([]);
      setStartTime(null);
    }
  };

  // 下一题
  const nextItem = async () => {
    const maxIndex = mode === 'word' ? (words?.length || 1) - 1 : (sentences?.length || 1) - 1;
    const totalItems = mode === 'word' ? (words?.length || 0) : (sentences?.length || 0);

    if (currentIndex < maxIndex) {
      const nextIndex = currentIndex + 1;

      // 保存进度
      try {
        await saveDictationProgress({
          article_id: articleId,
          dictation_type: mode,
          current_index: nextIndex,
          total_items: totalItems,
          score: score,
          completed: false,
        });
        console.log('✅ 进度已保存:', { 当前索引: nextIndex, 总数: totalItems, 得分: score });
      } catch (error: any) {
        if (error.response?.status === 401) {
          console.log('⚠️ 未登录，无法保存进度');
        } else {
          console.error('❌ 保存进度失败:', error);
        }
        // 不影响用户继续练习
      }

      setCurrentIndex(nextIndex);
      setUserInput('');
      setPracticeState('idle');
      setIsCorrect(null);
      setShowAnswer(false);
      setShowHint(false); // 重置提示显示状态
      setAttempts(0);
      setWordErrors([]); // 清空错误标记
      setIsFixingErrors(false);
      setCurrentFixingIndex(-1);
      setFixingInput('');
      setUserWordsArray([]);
      setStartTime(null);
    } else {
      // 练习完成
      try {
        await saveDictationProgress({
          article_id: articleId,
          dictation_type: mode,
          current_index: currentIndex,
          total_items: totalItems,
          score: score,
          completed: true, // 标记为已完成
        });
        console.log('🎉 练习完成，进度已保存');
      } catch (error: any) {
        if (error.response?.status === 401) {
          console.log('⚠️ 未登录，无法保存完成状态');
        } else {
          console.error('❌ 保存完成状态失败:', error);
        }
      }
      setPracticeState('result');
    }
  };

  // 庆祝动效 - 答对时喷射彩色粒子（超强版 + 超高z-index）
  const triggerCelebration = () => {
    console.log('🎉🎉🎉 triggerCelebration 函数开始执行！');
    console.log('confetti 类型:', typeof confetti);

    try {
      // 超级明显的测试版本 - 巨大的粒子，超高z-index
      const count = 200;
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

      console.log('🎆 开始超级庆祝动画');

      // 第一波：超大爆炸
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

      console.log('✅ 超级庆祝动画执行完成');

      // 持续喷射
      let duration = 3000;
      let animationEnd = Date.now() + duration;

      (function frame() {
        confetti({
          particleCount: 7,
          angle: 60,
          spread: 55,
          origin: { x: 0, y: 0.8 },
          zIndex: 999999,
          colors: ['#ff0000', '#00ff00', '#0000ff', '#ffff00', '#ff00ff', '#00ffff']
        });
        confetti({
          particleCount: 7,
          angle: 120,
          spread: 55,
          origin: { x: 1, y: 0.8 },
          zIndex: 999999,
          colors: ['#ff0000', '#00ff00', '#0000ff', '#ffff00', '#ff00ff', '#00ffff']
        });

        if (Date.now() < animationEnd) {
          requestAnimationFrame(frame);
        } else {
          console.log('🎉 持续庆祝动画结束');
        }
      }());

    } catch (error) {
      console.error('❌ 庆祝动画执行出错:', error);
    }
  };

  // 震动反馈（物理震动 + 视觉震动）
  const triggerVibration = (pattern: number | number[], animationType: 'success' | 'error') => {
    console.log('🎯 触发震动:', animationType, '模式:', pattern);

    // 1. 物理震动（移动设备）
    if ('vibrate' in navigator) {
      navigator.vibrate(pattern);
      console.log('✅ 物理震动已触发');
    } else {
      console.log('❌ 浏览器不支持物理震动');
    }

    // 2. 视觉震动（所有设备）- 针对练习区域震动
    const practiceArea = document.querySelector('.practice-item-no-card');
    console.log('🔍 找到元素:', practiceArea ? '是' : '否');

    if (practiceArea) {
      const className = `vibrate-${animationType}`;
      practiceArea.classList.add(className);
      console.log('✅ 添加CSS类:', className);

      setTimeout(() => {
        practiceArea.classList.remove(className);
        console.log('✅ 移除CSS类:', className);
      }, 600);
    } else {
      console.warn('⚠️ 未找到 .practice-item-no-card 元素');
    }
  };

  // 跳过当前题
  // 播放成功音效 - 使用竞品音效
  const playSuccessSound = () => {
    console.log('🎵 playSuccessSound 被调用');
    playCorrectSoundEffect(); // 使用新的打字音效 Hook
    triggerVibration([50, 30, 50], 'success'); // 答对：两次短震动

    // 检查 confetti 是否可用
    if (typeof confetti === 'undefined') {
      console.error('❌ confetti 库未加载！');
    } else {
      console.log('✅ confetti 库已加载，准备调用 triggerCelebration');
      triggerCelebration(); // 🎉 庆祝动画
    }
  };

  // 播放错误音效 - 使用竞品音效
  const playErrorSound = () => {
    playErrorSoundEffect(); // 使用新的打字音效 Hook
    triggerVibration([100, 50, 100, 50, 100], 'error'); // 答错：三次震动
  };


  const skipItem = () => {
    setShowAnswer(true);
    setIsCorrect(false);
    playErrorSound();
    setTimeout(() => {
      nextItem();
    }, 1000);
  };


  // 跳到下一个错误单词
  const goToNextError = () => {
    if (!isFixingErrors) return;

    const currentItem = getCurrentItem();
    if (!currentItem) return;

    console.log('goToNextError - current userWordsArray:', userWordsArray);
    console.log('goToNextError - current fixingInput:', fixingInput);
    console.log('goToNextError - current fixingIndex:', currentFixingIndex);

    // ★ 关键修复：先手动更新当前修复位置的单词到数组
    const updatedWordsArray = [...userWordsArray];
    if (currentFixingIndex >= 0 && currentFixingIndex < updatedWordsArray.length) {
      updatedWordsArray[currentFixingIndex] = fixingInput.trim();
      setUserWordsArray(updatedWordsArray);
    }

    console.log('goToNextError - updated userWordsArray:', updatedWordsArray);

    // 重新验证所有单词，找出真正还有错误的单词（使用更新后的数组）
    const answerWords = currentItem.text.split(' ').filter(w => /[a-zA-Z0-9]/.test(w));
    const updatedErrors = answerWords.map((answerWord, index) => {
      const userWord = updatedWordsArray[index] || ''; // 使用更新后的数组
      // BUG修复: 使用统一的标准化逻辑
      const normalizedAnswer = answerWord.toLowerCase().replace(/[^a-z0-9]/g, '');
      const normalizedUser = userWord.trim().toLowerCase().replace(/[^a-z0-9]/g, '');
      return normalizedAnswer !== normalizedUser;
    });

    console.log('Revalidated errors:', updatedErrors);
    setWordErrors(updatedErrors);

    // 找下一个真正的错误（从当前位置之后开始找）
    const nextErrorIndex = updatedErrors.findIndex((e, i) => e && i > currentFixingIndex);

    if (nextErrorIndex === -1) {
      // 检查是否还有之前的错误（从头开始找）
      const anyErrorIndex = updatedErrors.findIndex(e => e);

      if (anyErrorIndex === -1) {
        // 没有任何错误了，重新提交
        // 用空格连接所有单词，处理多余空格
        const updatedInput = updatedWordsArray.join(' ').replace(/\s+/g, ' ').trim();
        console.log('✅ All errors fixed, submitting:', updatedInput);
        console.log('Answer words:', answerWords);
        console.log('User words array:', updatedWordsArray);

        setIsFixingErrors(false);
        setCurrentFixingIndex(-1);
        setFixingInput('');

        // BUG修复: 先更新userInput，确保checkAnswer使用最新值
        // 修复策略: 同步更新后再异步调用checkAnswer
        // 影响范围: frontend/src/components/DictationPage.tsx:827-842
        // 修复日期: 2025-12-10
        setUserInput(updatedInput);

        // 延长等待时间，确保React状态完全更新
        setTimeout(() => {
          checkAnswer();
        }, 100);
      } else {
        // 还有之前的错误，跳到那个位置
        console.log('Going back to earlier error at:', anyErrorIndex);
        setCurrentFixingIndex(anyErrorIndex);
        setFixingInput(updatedWordsArray[anyErrorIndex] || '');
      }
    } else {
      // 跳到下一个错误
      console.log('Next error at index:', nextErrorIndex);
      setCurrentFixingIndex(nextErrorIndex);
      setFixingInput(updatedWordsArray[nextErrorIndex] || '');
    }
  };

  // 验证当前单词并移动到下一个错误（修复模式回车键）
  const validateAndMoveToNextError = () => {
    if (!isFixingErrors) return;

    const currentItem = getCurrentItem();
    if (!currentItem) return;

    console.log('=== validateAndMoveToNextError ===');
    console.log('Current fixingInput:', fixingInput);
    console.log('Current fixingIndex:', currentFixingIndex);

    // 1. 更新当前单词到数组
    const updatedWordsArray = [...userWordsArray];
    if (currentFixingIndex >= 0 && currentFixingIndex < updatedWordsArray.length) {
      updatedWordsArray[currentFixingIndex] = fixingInput.trim();
      setUserWordsArray(updatedWordsArray);
    }

    // 2. 验证当前单词是否正确
    const answerWords = currentItem.text.split(' ').filter(w => /[a-zA-Z0-9]/.test(w));
    const correctWord = answerWords[currentFixingIndex];

    // BUG修复: 统一使用严格的标准化逻辑
    // 修复策略: 去除所有标点、引号、空格，只保留字母和数字，统一小写比较
    // 影响范围: frontend/src/components/DictationPage.tsx:862-920
    // 修复日期: 2025-12-10
    const normalizedAnswer = correctWord
      .toLowerCase()
      .replace(/[^a-z0-9]/g, ''); // 只保留字母和数字
    const normalizedUser = fixingInput
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9]/g, ''); // 只保留字母和数字

    const isCurrentWordCorrect = normalizedAnswer === normalizedUser;

    console.log('Current word validation:');
    console.log('  Answer:', correctWord, '→', normalizedAnswer);
    console.log('  User:', fixingInput.trim(), '→', normalizedUser);
    console.log('  Is correct:', isCurrentWordCorrect);

    if (!isCurrentWordCorrect) {
      // 3. 如果当前单词不正确，清空输入让用户重新填写
      console.log('❌ Current word incorrect, clearing input');
      setFixingInput(''); // 清空当前输入
      updatedWordsArray[currentFixingIndex] = ''; // 清空数组中的值
      setUserWordsArray(updatedWordsArray);
      playErrorSound(); // 播放错误音效
      return;
    }

    // 4. 如果当前单词正确，继续处理（和原来的 goToNextError 逻辑一样）
    console.log('✅ Current word correct, moving to next error or submitting');

    // 重新验证所有单词（使用统一的标准化逻辑）
    const updatedErrors = answerWords.map((answerWord, index) => {
      const userWord = updatedWordsArray[index] || '';
      const normalizedAns = answerWord.toLowerCase().replace(/[^a-z0-9]/g, '');
      const normalizedUsr = userWord.trim().toLowerCase().replace(/[^a-z0-9]/g, '');
      return normalizedAns !== normalizedUsr;
    });

    console.log('Revalidated all errors:', updatedErrors);
    setWordErrors(updatedErrors);

    // 找下一个错误
    const nextErrorIndex = updatedErrors.findIndex((e, i) => e && i > currentFixingIndex);
    const anyErrorIndex = updatedErrors.findIndex(e => e);

    if (anyErrorIndex === -1) {
      // 没有错误了，提交整体答案
      const updatedInput = updatedWordsArray.join(' ').replace(/\s+/g, ' ').trim();
      console.log('✅ All words correct, submitting:', updatedInput);

      // BUG修复: 先更新状态，再延长等待确保状态完全同步
      setIsFixingErrors(false);
      setCurrentFixingIndex(-1);
      setFixingInput('');
      setUserInput(updatedInput);

      // 直接传入updatedInput，不依赖state更新
      setTimeout(() => {
        console.log('Submitting with direct input:', updatedInput);
        checkAnswer(updatedInput);
      }, 100);
    } else if (nextErrorIndex === -1 && anyErrorIndex !== -1) {
      // 后面没有错误了，但前面还有，跳回前面
      console.log('Going back to earlier error at:', anyErrorIndex);
      setCurrentFixingIndex(anyErrorIndex);
      setFixingInput(updatedWordsArray[anyErrorIndex] || '');
    } else {
      // 跳到下一个错误
      console.log('Next error at index:', nextErrorIndex);
      setCurrentFixingIndex(nextErrorIndex);
      setFixingInput(updatedWordsArray[nextErrorIndex] || '');
    }
  };

  // 更新修复输入
  const updateFixingInput = (value: string) => {
    console.log('updateFixingInput called:', value);
    setFixingInput(value);

    // 实时更新到数组
    const updatedWordsArray = [...userWordsArray];
    if (currentFixingIndex >= 0 && currentFixingIndex < updatedWordsArray.length) {
      updatedWordsArray[currentFixingIndex] = value;
      setUserWordsArray(updatedWordsArray);
      console.log('Updated userWordsArray:', updatedWordsArray);

      // 实时验证当前单词是否正确
      const currentItem = getCurrentItem();
      if (currentItem && mode === 'sentence') {
        const answerWords = currentItem.text.split(' ').filter(w => /[a-zA-Z0-9]/.test(w));
        const correctWord = answerWords[currentFixingIndex];

        if (correctWord) {
          // BUG修复: 使用统一的标准化逻辑
          const normalizedAnswer = correctWord.toLowerCase().replace(/[^a-z0-9]/g, '');
          const normalizedUser = value.trim().toLowerCase().replace(/[^a-z0-9]/g, '');

          if (normalizedAnswer === normalizedUser) {
            // 当前单词输入正确，更新错误状态
            const updatedErrors = [...wordErrors];
            updatedErrors[currentFixingIndex] = false;
            setWordErrors(updatedErrors);
            console.log('Word fixed correctly:', value);
          }
        }
      }
    }
  };

  // 重新再来 - 重新尝试当前单词
  const retryItem = () => {
    setUserInput('');
    setShowAnswer(false);
    setIsCorrect(null);
    setAttempts(0);
    setWordErrors([]); // 清空错误标记
    setIsFixingErrors(false);
    setCurrentFixingIndex(-1);
    setFixingInput('');
    setUserWordsArray([]);
    setStartTime(Date.now());
    setPracticeState('typing');
    setTimeout(() => inputRef.current?.focus(), 100);
  };

  // 重新开始
  const restart = async () => {
    // 重置进度到初始状态
    try {
      await saveDictationProgress({
        article_id: articleId,
        dictation_type: mode,
        current_index: 0,
        total_items: mode === 'word' ? (words?.length || 0) : (sentences?.length || 0),
        score: 0,
        completed: false,
      });
      console.log('🔄 进度已重置');
    } catch (error: any) {
      if (error.response?.status === 401) {
        console.log('⚠️ 未登录，无法保存进度重置');
      } else {
        console.error('❌ 重置进度失败:', error);
      }
    }

    setCurrentIndex(0);
    setUserInput('');
    setPracticeState('idle');
    setIsCorrect(null);
    setScore(0);
    setAttempts(0);
    setShowAnswer(false);
    setWordErrors([]); // 清空错误标记
    setIsFixingErrors(false);
    setCurrentFixingIndex(-1);
    setFixingInput('');
    setUserWordsArray([]);
    setStartTime(null);
  };

  // 安全地获取当前项和总数（数据未加载时返回安全值）
  const totalItems = mode === 'word' ? (words?.length || 0) : (sentences?.length || 0);
  const currentItem = dataLoaded ? getCurrentItem() : null;
  const progress = totalItems > 0 ? ((currentIndex + 1) / totalItems) * 100 : 0;
  const isCompleted = practiceState === 'result';

  return (
    <div className={`dictation-page-overlay ${theme === 'dark' ? 'dark' : ''}`}>
      <FloatingBackButton show={true} onBack={onBack} />
      <div className="dictation-page">
      {/* 音效元素 */}
      <audio ref={successAudioRef} preload="auto">
        <source src="/sounds/success.mp3" type="audio/mpeg" />
        {/* 如果音效文件不存在，使用Web Audio API生成 */}
      </audio>
      <audio ref={errorAudioRef} preload="auto">
        <source src="/sounds/error.mp3" type="audio/mpeg" />
        {/* 如果音效文件不存在，使用Web Audio API生成 */}
      </audio>

      {/* 顶部导航 */}
      <div className="dictation-header">
        <div className="header-left">
        </div>
        <div className="dictation-title-container">
          <div className="dictation-title-text">
            {articleTitle && <p className="article-title-subtitle">{articleTitle}</p>}
          </div>
        </div>
        <div className="header-right">
          {/* 全屏沉浸按钮 - 放在模式切换前面 */}
          <button
            className="fullscreen-btn"
            onClick={isFullscreen ? exitFullscreen : enterFullscreen}
            title={isFullscreen ? '退出全屏' : '沉浸式练习（全屏）'}
          >
            {isFullscreen ? (
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3"/>
              </svg>
            ) : (
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/>
              </svg>
            )}
          </button>
          {/* 只有当有多种模式可选时才显示切换按钮 */}
          {((words?.length || 0) > 0 || (sentences?.length || 0) > 0) && (
            <div className="dictation-mode-toggle">
              {/* 只有当有单词数据时才显示单词按钮 */}
              {(words?.length || 0) > 0 && (
                <button
                  className={`mode-btn ${mode === 'word' ? 'active' : ''}`}
                  onClick={() => {
                    setMode('word');
                    setCurrentIndex(0);
                    setUserInput('');
                    setPracticeState('idle');
                    setDataLoaded(false); // 切换模式时重置加载状态
                  }}
                >
                  单词
                </button>
              )}
              {/* 只有当有句子数据时才显示句子按钮 */}
              {(sentences?.length || 0) > 0 && (
                <button
                  className={`mode-btn ${mode === 'sentence' ? 'active' : ''}`}
                  onClick={() => {
                    setMode('sentence');
                    setCurrentIndex(0);
                    setUserInput('');
                    setPracticeState('idle');
                    setDataLoaded(false); // 切换模式时重置加载状态
                  }}
                >
                  句子
                </button>
              )}
            </div>
          )}
          <button
            className="theme-toggle-btn"
            onClick={() => {
              console.log('🌓 theme toggle clicked, current theme:', theme);
              setTheme(theme === 'light' ? 'dark' : 'light');
            }}
            title={theme === 'light' ? '切换到深色模式' : '切换到浅色模式'}
          >
            {theme === 'light' ? (
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
              </svg>
            ) : (
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
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

      {/* 进度条 */}
      <div className="dictation-progress">
        <div className="progress-bar">
          <div
            className="progress-fill"
            style={{ width: `${progress}%` }}
          ></div>
        </div>
        <div className="progress-text">
          {currentIndex + 1} / {totalItems}
        </div>
      </div>

      {/* 主内容区 */}
      <div className="dictation-content">
        {isCompleted ? (
          // 完成页面
          <div className="dictation-result">
            <div className="result-icon">
              <svg viewBox="0 0 100 100">
                <circle cx="50" cy="50" r="45" fill="none" stroke="var(--accent-color)" strokeWidth="4"/>
                <path d="M30 50 L45 65 L70 35" fill="none" stroke="var(--accent-color)" strokeWidth="4" strokeLinecap="round"/>
              </svg>
            </div>
            <h2>练习完成！</h2>
            <div className="result-stats">
              <div className="stat-item">
                <div className="stat-value">{score}</div>
                <div className="stat-label">正确</div>
              </div>
              <div className="stat-item">
                <div className="stat-value">{totalItems - score}</div>
                <div className="stat-label">错误</div>
              </div>
              <div className="stat-item">
                <div className="stat-value">{Math.round((score / totalItems) * 100)}%</div>
                <div className="stat-label">正确率</div>
              </div>
            </div>
            <div className="result-actions">
              <button className="action-btn primary" onClick={restart}>
                重新开始
              </button>
              <button className="action-btn secondary" onClick={onBack}>
                返回阅读
              </button>
            </div>
          </div>
        ) : dataLoaded && currentItem ? (
          // 练习页面（确保数据完全加载后再渲染）
          <div className="dictation-practice">
            {/* 功能实现: 加入生词本按钮 */}
            {/* 实现方案: 悬浮在右上角，点击弹出AddVocabularyPopup */}
            {/* 影响范围: frontend/src/components/DictationPage.tsx */}
            {/* 实现日期: 2025-01-27 */}
            <button 
              className="add-to-vocabulary-btn"
              onClick={() => {
                console.log('📚 点击加入生词本', { 
                  currentItem, 
                  articleId, 
                  mode,
                  hasPopup: !!showVocabularyPopup 
                });
                setShowVocabularyPopup(true);
              }}
              title="加入生词本"
            >
              <img src="/icon/shengciben.png" alt="生词本" />
              <span>生词本</span>
            </button>

            {/* 当前题目 */}
            <div className="practice-item-no-card">
              {practiceState === 'listening' && (
                <div className="item-listening">
                  <div className="listening-animation">
                    <div className="sound-wave"></div>
                    <div className="sound-wave"></div>
                    <div className="sound-wave"></div>
                  </div>
                  <p className="listening-status">正在朗读{mode === 'word' ? '单词' : '句子'}...</p>

                  {/* 功能实现: 朗读时显示内容和翻译 */}
                  {/* 实现方案: 在listening状态下展示英文原文和中文翻译 */}
                  {/* 影响范围: frontend/src/components/DictationPage.tsx:1192-1201 */}
                  {/* 实现日期: 2025-12-10 */}
                  {currentItem && (
                    <div className="listening-content">
                      {mode === 'word' ? (
                        // 单词模式：显示音标和释义
                        <>
                          {(currentItem as Word).phonetic && (
                            <div className="listening-phonetic">{(currentItem as Word).phonetic}</div>
                          )}
                          {(currentItem as Word).meaning && (
                            <div className="listening-meaning">{(currentItem as Word).meaning}</div>
                          )}
                        </>
                      ) : (
                        // 句子模式：显示英文句子和中文翻译
                        <>
                          <div className="listening-sentence">{currentItem.text}</div>
                          {(currentItem as Sentence).translation && (
                            <div className="listening-translation">{(currentItem as Sentence).translation}</div>
                          )}
                        </>
                      )}
                    </div>
                  )}
                </div>
              )}

              {/* idle状态的简单显示 */}
              {practiceState === 'idle' && (
                <div className="item-idle">
                  <p>准备好后，按回车键开始听写</p>
                </div>
              )}

              {(practiceState === 'typing' || practiceState === 'checking') && (
                <div className="item-typing">
                  {/* 下划线提示区域 - 始终显示 */}
                  {currentItem && (
                    <div className="word-hint">
                      {/* 单词意思显示（仅单词）- 放在最上面 */}
                      {mode === 'word' && (currentItem as Word).meaning && (
                        <div className="word-meaning">
                          {(currentItem as Word).meaning}
                        </div>
                      )}
                      {/* 音标显示（仅单词）- 放在单词意思下面 */}
                      {mode === 'word' && (currentItem as Word).phonetic && (
                        <div className="word-phonetic">
                          {(currentItem as Word).phonetic}
                        </div>
                      )}
                      {/* 句子翻译显示（在横线上方） */}
                      {mode === 'sentence' && currentItem && (
                        <div className="sentence-translation-hint">
                          {(currentItem as Sentence).translation || '正在加载翻译...'}
                        </div>
                      )}
                      {/* 输入区域容器 - 相对定位，用于放置透明input */}
                      <div className="input-display-container" onClick={() => inputRef.current?.focus()}>
                        {mode === 'word' ? (
                          // 单词模式：按字母显示
                          <div className="underscore-display">
                            {currentItem && Array.from({ length: currentItem.text.length }, (_, i) => (
                              <span key={i} className="underscore-char">
                                {userInput[i] || '_'}
                              </span>
                            ))}
                          </div>
                        ) : (
                          // 句子模式：按单词显示
                          <div
                            className="sentence-words-display"
                            style={{
                              fontSize: (() => {
                                if (!currentItem) return '32px';
                                const sentenceLength = currentItem.text.length;
                                if (sentenceLength > 80) return '20px';
                                if (sentenceLength > 60) return '24px';
                                if (sentenceLength > 40) return '28px';
                                return '32px'; // 默认
                              })(),
                            }}
                          >
                            {(() => {
                              if (!currentItem) return null;
                              if (isFixingErrors) {
                                // 修复模式：只显示和编辑当前错误的单词
                                return currentItem.text.split(' ').map((part, wordIndex) => {
                                  const isWord = /[a-zA-Z0-9]/.test(part);

                                  if (!isWord) {
                                    return (
                                      <div key={wordIndex} className="sentence-punctuation">
                                        {part}
                                      </div>
                                    );
                                  }

                                  const actualWordIndex = currentItem.text.split(' ')
                                    .slice(0, wordIndex)
                                    .filter(w => /[a-zA-Z0-9]/.test(w)).length;

                                  const isCurrentFixing = actualWordIndex === currentFixingIndex;
                                  const hasError = wordErrors[actualWordIndex];
                                  const isCorrectWord = !hasError;

                                  let className = 'sentence-word-box';
                                  let displayText = '';

                                  if (isCurrentFixing) {
                                    // 当前正在修复的单词 - 显示实时输入
                                    className += ' active';
                                    displayText = fixingInput || '　'.repeat(Math.max(part.length, 4));
                                  } else if (isCorrectWord) {
                                    // 正确的单词 - 显示用户输入
                                    className += ' completed';
                                    displayText = userWordsArray[actualWordIndex] && userWordsArray[actualWordIndex].trim()
                                      ? userWordsArray[actualWordIndex]
                                      : part;
                                  } else {
                                    // 其他错误单词 - 显示已修复的内容或空格 + 抖动动画
                                    className += ' error shake';
                                    displayText = userWordsArray[actualWordIndex] && userWordsArray[actualWordIndex].trim()
                                      ? userWordsArray[actualWordIndex]
                                      : '　'.repeat(Math.max(part.length, 4));
                                  }

                                  return (
                                    <div key={wordIndex} className={className}>
                                      {displayText}
                                    </div>
                                  );
                                });
                              } else {
                                // 正常输入模式
                                const userWords = userInput.split(' ');
                                const currentInputIndex = userInput.endsWith(' ')
                                  ? userWords.length - 1
                                  : Math.max(0, userWords.filter(w => w.trim()).length - 1);

                                return currentItem.text.split(' ').map((part, wordIndex) => {
                                  const isWord = /[a-zA-Z0-9]/.test(part);

                                  if (!isWord) {
                                    return (
                                      <div key={wordIndex} className="sentence-punctuation">
                                        {part}
                                      </div>
                                    );
                                  }

                                  const actualWordIndex = currentItem.text.split(' ')
                                    .slice(0, wordIndex)
                                    .filter(w => /[a-zA-Z0-9]/.test(w)).length;

                                  const userWord = userWords[actualWordIndex] || '';
                                  const hasError = showAnswer && !isCorrect && wordErrors[actualWordIndex];
                                  const isActive = !showAnswer && actualWordIndex === currentInputIndex;
                                  const isCompleted = userWord && actualWordIndex < currentInputIndex;

                                  let className = 'sentence-word-box';
                                  if (hasError) {
                                    className += ' error';
                                  } else if (showAnswer && isCorrect) {
                                    className += ' completed';
                                  } else if (showAnswer && !isCorrect && !wordErrors[actualWordIndex]) {
                                    className += ' completed';
                                  } else if (isActive) {
                                    className += ' active';
                                  } else if (isCompleted) {
                                    className += ' completed';
                                  } else {
                                    className += ' empty';
                                  }

                                  return (
                                    <div key={wordIndex} className={className}>
                                      {userWord || '　'.repeat(Math.max(part.length, 4))}
                                    </div>
                                  );
                                });
                              }
                            })()}
                          </div>
                        )}

                        {/* 透明输入框覆盖在横线上方 */}
                        <input
                          ref={inputRef}
                          type="text"
                          value={isFixingErrors ? fixingInput : userInput}
                          onChange={(e) => {
                            if (isFixingErrors) {
                              updateFixingInput(e.target.value);
                            } else {
                              setUserInput(e.target.value);
                            }
                          }}
                          onKeyDown={(e) => {
                            // 功能实现: earthworm风格的键盘输入处理
                            // 实现方案: 空格智能提交、退格回到上一个错误单词
                            // 影响范围: frontend/src/components/DictationPage.tsx onKeyDown
                            // 实现日期: 2025-12-10

                            // 播放打字音效
                            if (shouldPlayTypingSound(e as any)) {
                              playTypingSound();
                            }

                            // 禁止方向键移动
                            if (['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight'].includes(e.key)) {
                              e.preventDefault();
                              return;
                            }

                            // 【句子模式 + 正常输入】空格 + 在最后一个单词 → 提交答案
                            if (mode === 'sentence' && !isFixingErrors && e.key === ' ' && isLastWord() && practiceState === 'typing') {
                              e.preventDefault();
                              e.stopPropagation();
                              checkAnswer();
                              return;
                            }

                            // 修复模式：空格跳到下一个错误
                            if (isFixingErrors && e.key === ' ') {
                              e.preventDefault();
                              goToNextError();
                              return;
                            }

                            // 修复模式：退格键 + 当前单词为空 → 回到上一个错误单词
                            if (isFixingErrors && e.key === 'Backspace' && isCurrentWordEmpty()) {
                              e.preventDefault();
                              const prevIndex = findPrevErrorIndex();
                              if (prevIndex !== -1) {
                                setCurrentFixingIndex(prevIndex);
                                setFixingInput(userWordsArray[prevIndex] || '');
                              }
                              return;
                            }

                            // 修复模式：回车验证当前单词并跳转
                            if (isFixingErrors && e.key === 'Enter') {
                              e.preventDefault();
                              validateAndMoveToNextError();
                              return;
                            }

                            // 正常模式：回车提交答案
                            if (e.key === 'Enter' && practiceState === 'typing' && !isFixingErrors) {
                              e.preventDefault();
                              checkAnswer();
                            } else if (e.key === ' ' && mode === 'word' && practiceState === 'typing') {
                              // 单词模式：空格播放声音
                              e.preventDefault();
                              playPronunciation();
                            }
                            // 句子模式：空格正常输入（除非在最后一个单词）
                          }}
                          disabled={practiceState === 'checking'}
                          className="overlay-input"
                          maxLength={mode === 'word' && currentItem ? currentItem.text.length : undefined}
                          autoFocus
                        />
                      </div>
                    </div>
                  )}

                  {/* 答案提示区域 - 仅显示翻译和提示信息 */}
                  {showAnswer && !isCorrect && currentItem && (
                    <div className="answer-tips">
                      <div className="answer-text incorrect">
                        ✗ {mode === 'sentence' ? '部分单词有误（红色标记）' : '答案错误'}
                      </div>

                      {mode === 'word' && (currentItem as Word).meaning && (
                        <div className="word-meaning">{(currentItem as Word).meaning}</div>
                      )}
                      {mode === 'word' && (currentItem as Word).example && (
                        <>
                          <div className="word-example">例句：{(currentItem as Word).example}</div>
                          {(currentItem as Word).example_translation && (
                            <div className="word-example-translation">翻译：{(currentItem as Word).example_translation}</div>
                          )}
                        </>
                      )}
                      {mode === 'sentence' && (currentItem as Sentence).translation && (
                        <div className="word-example-translation">翻译：{(currentItem as Sentence).translation}</div>
                      )}
                    </div>
                  )}

                  {/* 答对时的提示 */}
                  {showAnswer && isCorrect && (
                    <div className="answer-tips">
                      <div className="answer-text correct">
                        ✓ 正确
                      </div>
                    </div>
                  )}

                  {/* 提示信息显示 - 单词模式 */}
                  {showHint && mode === 'word' && currentItem && (currentItem as Word).meaning && (
                    <div className="hint-display">
                      <div className="hint-header">
                        <span className="hint-icon">💡</span>
                        <span className="hint-title">提示信息</span>
                      </div>
                      <div className="hint-word">
                        <strong>单词：</strong>{currentItem?.text || ''}
                        {(currentItem as Word).phonetic && (
                          <span className="hint-word-phonetic"> {(currentItem as Word).phonetic}</span>
                        )}
                      </div>
                      {(currentItem as Word).meaning && (
                        <div className="hint-meaning">
                          <strong>释义：</strong>{(currentItem as Word).meaning}
                        </div>
                      )}
                      {(currentItem as Word).example && (
                        <>
                          <div className="hint-example">
                            <strong>例句：</strong>{(currentItem as Word).example}
                            <button
                              className="play-example-btn"
                              onClick={() => playExample((currentItem as Word).example!)}
                              title="播放例句"
                            >
                              <svg viewBox="0 0 24 24" fill="currentColor">
                                <path d="M8 5v14l11-7z"/>
                              </svg>
                            </button>
                          </div>
                          {(currentItem as Word).example_translation && (
                            <div className="hint-translation">
                              <strong>翻译：</strong>{(currentItem as Word).example_translation}
                            </div>
                          )}
                        </>
                      )}
                    </div>
                  )}

                  {/* 提示信息显示 - 句子模式 */}
                  {showHint && mode === 'sentence' && currentItem && (currentItem as Sentence).translation && (
                    <div className="hint-display">
                      <div className="hint-header">
                        <span className="hint-icon">💡</span>
                        <span className="hint-title">提示信息</span>
                      </div>
                      <div className="hint-word">
                        <strong>句子：</strong>{currentItem?.text || ''}
                      </div>
                      <div className="hint-translation">
                        <strong>翻译：</strong>{(currentItem as Sentence).translation}
                      </div>
                    </div>
                  )}

                  <div className="input-actions">
                    {practiceState === 'typing' && !showAnswer && (
                      <>
                        <button className="action-btn secondary" onClick={() => playPronunciation()} title="重新播放 (Ctrl+/)">
                          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M23 4v6h-6"></path>
                            <path d="M1 20v-6h6"></path>
                            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
                          </svg>
                          <span>重播</span>
                        </button>

                        {currentItem && ((mode === 'word' && (currentItem as Word).meaning) || (mode === 'sentence' && (currentItem as Sentence).translation)) && (
                          <button
                            className={`action-btn secondary ${showHint ? 'active' : ''}`}
                            onClick={() => setShowHint(prev => !prev)}
                            title="显示/隐藏提示 (Tab)"
                          >
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                              <path d="M9 18h6"></path>
                              <path d="M10 22h4"></path>
                              <path d="M12 2v1"></path>
                              <path d="M12 15a7 7 0 0 1-7-7c0-3.87 3.13-7 7-7s7 3.13 7 7a7 7 0 0 1-7 7z"></path>
                            </svg>
                            <span>{showHint ? '隐藏' : '提示'}</span>
                          </button>
                        )}

                        <button className="action-btn secondary" onClick={() => {
                          // 重新开始：清空当前输入
                          setUserInput('');
                          setFixingInput('');
                          setUserWordsArray([]);
                          setWordErrors([]);
                          setIsFixingErrors(false);
                          setCurrentFixingIndex(-1);
                          inputRef.current?.focus();
                        }} title="重新开始 (Esc)">
                          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"></path>
                            <path d="M3 3v5h5"></path>
                          </svg>
                          <span>重新开始</span>
                        </button>

                        <button className="action-btn secondary" onClick={skipItem} title="跳过当前题">
                          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <polygon points="5 4 15 12 5 20 5 4"></polygon>
                            <line x1="19" y1="5" x2="19" y2="19"></line>
                          </svg>
                          <span>跳过</span>
                        </button>

                        <button
                          className="action-btn primary"
                          onClick={() => checkAnswer()}
                          disabled={!userInput.trim()}
                          title="提交答案 (Enter)"
                        >
                          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <polyline points="20 6 9 17 4 12"></polyline>
                          </svg>
                          <span>提交</span>
                        </button>
                      </>
                    )}
                  </div>

                  {/* 功能实现: 快捷键提示区域 - 移到 input-actions 外面 */}
                  {/* 实现日期: 2025-12-10 */}
                  {practiceState === 'typing' && !showAnswer && (
                    <div className="keyboard-shortcuts-hint">
                      <span className="shortcut-item">
                        <kbd>Enter</kbd> {isFixingErrors ? '验证' : '提交'}
                      </span>
                      {isFixingErrors ? (
                        <>
                          <span className="shortcut-item">
                            <kbd>Space</kbd> 下一个
                          </span>
                          <span className="shortcut-item">
                            <kbd>Backspace</kbd> 上一个
                          </span>
                        </>
                      ) : (
                        <span className="shortcut-item">
                          <kbd>Space</kbd> 末词提交
                        </span>
                      )}
                      <span className="shortcut-item">
                        <kbd>Esc</kbd> 重开
                      </span>
                      <span className="shortcut-item">
                        <kbd>Tab</kbd> 提示
                      </span>
                      <span className="shortcut-item">
                        <kbd>Ctrl /</kbd> 重播
                      </span>
                      {!isFixingErrors && currentIndex > 0 && (
                        <span className="shortcut-item">
                          <kbd>Ctrl ,</kbd> 上题
                        </span>
                      )}
                      {!isFixingErrors && (
                        <span className="shortcut-item">
                          <kbd>Ctrl .</kbd> 下题
                        </span>
                      )}
                      <span className="shortcut-item">
                        <kbd>Ctrl S</kbd> 生词本
                      </span>
                    </div>
                  )}

                  <div className="input-actions">
                    {showAnswer && !isCorrect && mode === 'word' && (
                      <>
                        <button className="action-btn secondary" onClick={retryItem}>
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                          </svg>
                          重新再来
                        </button>
                        <button className="action-btn primary" onClick={nextItem}>
                          下一题
                        </button>
                      </>
                    )}

                    {/* 句子模式：答错后只能修改或重新开始 */}
                    {showAnswer && !isCorrect && mode === 'sentence' && (
                      <button className="action-btn secondary" onClick={retryItem}>
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                          <path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                        </svg>
                        放弃，重新开始
                      </button>
                    )}
                  </div>
                </div>
              )}

            </div>
          </div>
        ) : !dataLoaded ? (
          <div className="dictation-empty">
            <div className="loading-spinner"></div>
            <p>正在加载{mode === 'word' ? '单词' : '句子'}数据...</p>
          </div>
        ) : (
          <div className="dictation-empty">
            <p>暂无{mode === 'word' ? '单词' : '句子'}数据</p>
          </div>
        )}
      </div>
      </div>

      {/* 功能实现: 生词本弹窗 */}
      {/* 实现方案: 复用AddVocabularyPopup组件 */}
      {/* 影响范围: frontend/src/components/DictationPage.tsx */}
      {/* 实现日期: 2025-01-27 */}
      {showVocabularyPopup && (
        currentItem ? (
          <AddVocabularyPopup
            selectedText={currentItem.text}
            position={{ x: 0, y: 0 }}
            articleId={articleId}
            context={mode === 'word' ? (currentItem as Word).example : undefined}
            onClose={() => setShowVocabularyPopup(false)}
            onSuccess={() => setShowVocabularyPopup(false)}
          />
        ) : (
          (() => {
            console.error('❌ 尝试打开生词本但 currentItem 为空');
            setShowVocabularyPopup(false);
            return null;
          })()
        )
      )}
    </div>
  );
};

