import { useEffect, useRef, useState } from 'react';

interface TimelineSegment {
  text: string;
  time_begin: number;
  time_end: number;
  text_begin: number;
  text_end: number;
  index?: number;
  cleanText?: string;
}

export const useAudioHighlight = (currentTime: number, articleContentRef: React.RefObject<HTMLDivElement>) => {
  const [timelineData, setTimelineData] = useState<TimelineSegment[]>([]);
  const currentHighlightIndexRef = useRef<number>(-1);
  const currentHighlightParaIndexRef = useRef<number>(-1);
  const lastTimeRef = useRef<number>(0);

  // 加载时间轴数据
  useEffect(() => {
    const loadTimelineData = async () => {
      try {
        const response = await fetch('/timeline.json');
        const data = await response.json();
        // 预处理：为每个时间段创建映射
        data.forEach((item: TimelineSegment, index: number) => {
          item.index = index;
          item.cleanText = item.text.trim().replace(/\s+/g, ' ');
        });
        setTimelineData(data);
        console.log('✅ 时间轴数据加载成功:', data.length, '条');
      } catch (error) {
        console.error('❌ 时间轴数据加载失败:', error);
      }
    };
    loadTimelineData();
  }, []);

  // 根据音频时间查找对应的文本段
  const findCurrentSegment = (currentTime: number): number => {
    const currentTimeMs = currentTime * 1000;
    for (let i = 0; i < timelineData.length; i++) {
      const segment = timelineData[i];
      if (currentTimeMs >= segment.time_begin && currentTimeMs <= segment.time_end) {
        return i;
      }
    }
    return -1;
  };

  // 移除高亮
  const removeHighlight = () => {
    if (!articleContentRef.current) return;
    const highlights = articleContentRef.current.querySelectorAll('.highlight');
    highlights.forEach(element => {
      element.classList.remove('highlight');
    });
  };

  // 移除当前位置指示器
  const removeCurrentIndicator = () => {
    if (!articleContentRef.current) return;
    const indicators = articleContentRef.current.querySelectorAll('.current-indicator');
    indicators.forEach(indicator => {
      indicator.remove();
    });
  };

  // 滚动到高亮位置
  const scrollToHighlight = () => {
    if (!articleContentRef.current) return;
    const highlight = articleContentRef.current.querySelector('.highlight');
    if (highlight) {
      const rect = highlight.getBoundingClientRect();
      const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
      const targetTop = scrollTop + rect.top - (window.innerHeight / 3);
      window.scrollTo({
        top: targetTop,
        behavior: 'smooth'
      });
    }
  };

  // 在内容中高亮文本
  const highlightTextInContent = (text: string, segmentIndex: number, totalSegments: number) => {
    if (!articleContentRef.current) {
      console.warn('⚠️ articleContentRef.current 为空');
      return;
    }

    removeHighlight();
    removeCurrentIndicator();

    const containers = articleContentRef.current.querySelectorAll('.highlight-container');
    containers.forEach(c => c.classList.remove('highlight-container'));

    // 预处理目标文本：提取核心内容（只保留汉字、字母、数字）
    const cleanTarget = text.replace(/[^\u4e00-\u9fa5a-zA-Z0-9]/g, '').toLowerCase();
    if (cleanTarget.length < 5) {
      return;
    }

    // 先查找所有可能的段落元素
    const allElements = Array.from(articleContentRef.current.querySelectorAll('p, h1, h2, h3, h4, h5, h6, li, blockquote'));
    
    // 过滤出最底层的元素（不包含其他块级元素的元素）
    const paragraphs = allElements.filter(el => {
      // 检查是否包含其他块级元素
      const hasBlockChildren = el.querySelector('p, h1, h2, h3, h4, h5, h6, li, blockquote');
      return !hasBlockChildren;
    });

    const matches: Array<{ element: Element; index: number; score: number; matchPosition?: number }> = [];

    paragraphs.forEach((para, index) => {

      const paraText = para.textContent || "";
      const cleanPara = paraText.replace(/[^\u4e00-\u9fa5a-zA-Z0-9]/g, '').toLowerCase();

      if (cleanPara.length < 5) return;

      // 核心匹配逻辑
      if (cleanPara.includes(cleanTarget)) {
        const matchPosition = cleanPara.indexOf(cleanTarget);
        const matchRatio = matchPosition / Math.max(cleanPara.length, 1);
        const positionScore = matchRatio > 0.5 ? 0.1 : 0;
        matches.push({ element: para, index: index, score: 1.0 + positionScore, matchPosition: matchPosition });
      } else if (cleanTarget.includes(cleanPara)) {
        matches.push({ element: para, index: index, score: 1.0 });
      } else {
        const start = cleanTarget.substring(0, Math.min(30, cleanTarget.length));
        const end = cleanTarget.substring(Math.max(0, cleanTarget.length - 30));
        const hasStart = cleanPara.includes(start);
        const hasEnd = cleanPara.includes(end);
        
        if (hasStart || hasEnd) {
          let score = 0.5;
          if (hasStart && hasEnd) {
            score = 0.8;
          }
          matches.push({ element: para, index: index, score: score });
        }
      }
    });

    if (matches.length > 0) {
      // 排序：先按分数，再按索引
      matches.sort((a, b) => {
        if (Math.abs(a.score - b.score) > 0.1) {
          return b.score - a.score;
        }
        return a.index - b.index;
      });

      let bestMatches = matches;

      // 策略优化：多重匹配消歧
      if (matches.length > 1) {
        // 优先使用上下文 (Sequential Playback)
        if (currentHighlightParaIndexRef.current >= 0) {
          const subsequentMatches = matches.filter(m => m.index > currentHighlightParaIndexRef.current);
          if (subsequentMatches.length > 0) {
            const bestSubsequent = subsequentMatches.reduce((prev, curr) => {
              return curr.score > prev.score ? curr : prev;
            });
            bestMatches = [bestSubsequent];
          }
        }

        // 使用位置比率 (Ratio Heuristic)
        if (bestMatches.length > 1 && typeof segmentIndex === 'number' && typeof totalSegments === 'number') {
          const audioProgress = segmentIndex / totalSegments;
          const totalParas = paragraphs.length;

          bestMatches = [matches.reduce((prev, curr) => {
            const prevRatio = prev.index / totalParas;
            const currRatio = curr.index / totalParas;
            const prevDiff = Math.abs(prevRatio - audioProgress);
            const currDiff = Math.abs(currRatio - audioProgress);
            
            if (Math.abs(prevDiff - currDiff) < 0.05) {
              return curr.index > prev.index ? curr : prev;
            }
            return currDiff < prevDiff ? curr : prev;
          })];
        }
      }

      // "填补空缺" 逻辑
      if (bestMatches.length >= 2) {
        const firstIndex = bestMatches[0].index;
        const lastIndex = bestMatches[bestMatches.length - 1].index;

        if (lastIndex - firstIndex < 5) {
          for (let i = firstIndex + 1; i < lastIndex; i++) {
            const exists = bestMatches.find(m => m.index === i);
            if (!exists) {
              const gapPara = paragraphs[i];
              if (gapPara && gapPara.textContent && gapPara.textContent.trim().length > 0) {
                bestMatches.push({ element: gapPara, index: i, score: 0.5 });
              }
            }
          }
          bestMatches.sort((a, b) => a.index - b.index);
        }
      }

      if (bestMatches.length > 0) {
        const bestMatch = bestMatches.length > 1 
          ? bestMatches.reduce((prev, curr) => {
              if (Math.abs(curr.score - prev.score) > 0.1) {
                return curr.score > prev.score ? curr : prev;
              }
              return curr.index > prev.index ? curr : prev;
            })
          : bestMatches[0];

        currentHighlightParaIndexRef.current = bestMatch.index;

        // 高亮最佳匹配及其相邻的匹配
        bestMatches.forEach((match) => {
          match.element.classList.add('highlight');

          const parent = match.element.parentElement;
          if (parent && parent !== articleContentRef.current) {
            parent.classList.add('highlight-container');
          }

          // 只在最佳匹配上显示指示器
          if (match.index === bestMatch.index) {
            const indicator = document.createElement('div');
            indicator.className = 'current-indicator';
            indicator.innerHTML = '▶';
            (match.element as HTMLElement).style.position = 'relative';
            match.element.insertBefore(indicator, match.element.firstChild);
          }
        });
        scrollToHighlight();
      }
    }
  };

  // 当音频时间更新时，更新高亮
  useEffect(() => {
    if (timelineData.length === 0 || !articleContentRef.current) return;

    // 如果时间向后跳转（用户拖动了进度条），重置高亮索引
    if (currentTime < lastTimeRef.current - 1) {
      currentHighlightIndexRef.current = -1;
      currentHighlightParaIndexRef.current = -1;
      removeHighlight();
      removeCurrentIndicator();
    }
    lastTimeRef.current = currentTime;

    const segmentIndex = findCurrentSegment(currentTime);
    if (segmentIndex !== -1 && segmentIndex !== currentHighlightIndexRef.current) {
      const segment = timelineData[segmentIndex];
      if (segment) {
        highlightTextInContent(segment.text, segmentIndex, timelineData.length);
        currentHighlightIndexRef.current = segmentIndex;
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentTime, timelineData]);

  return {
    timelineData,
    removeHighlight,
  };
};

