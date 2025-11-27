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
      // 清除可能的内联样式
      (element as HTMLElement).style.marginLeft = '';
      (element as HTMLElement).style.width = '';
      (element as HTMLElement).style.paddingLeft = '';
    });
  };

  // 动态修正高亮框对齐
  const fixHighlightAlignment = (element: HTMLElement) => {
    if (!articleContentRef.current) return;
    
    // 使用 requestAnimationFrame 确保在下一帧计算，获取准确的样式
    requestAnimationFrame(() => {
      if (!articleContentRef.current) return;
      
      // 先移除可能的内联样式以获取原始位置
      element.style.marginLeft = '';
      element.style.width = '';
      element.style.paddingLeft = '';
      
      const containerRect = articleContentRef.current.getBoundingClientRect();
      const elementRect = element.getBoundingClientRect();
      
      // 计算左侧偏移量
      const leftOffset = elementRect.left - containerRect.left;
      
      // 如果有偏移（说明有缩进），进行修正
      // 允许 5px 的误差，避免微小抖动
      if (leftOffset > 5) {
        element.style.marginLeft = `-${leftOffset}px`;
        element.style.width = `calc(100% + ${leftOffset}px)`;
        // 关键：增加 padding-left 以保持文字位置不变
        // 24px 是基础 padding (从 CSS .highlight { padding: 20px 24px } 中获取)
        element.style.paddingLeft = `${leftOffset + 24}px`;
      }
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
  const highlightTextInContent = (text: string, segmentIndex: number, totalSegments: number, segment?: TimelineSegment) => {
    if (!articleContentRef.current) {
      console.warn('⚠️ articleContentRef.current 为空');
      return;
    }

    removeHighlight();
    removeCurrentIndicator();

    const containers = articleContentRef.current.querySelectorAll('.highlight-container');
    containers.forEach(c => c.classList.remove('highlight-container'));

    // **方案1：优先使用 text_begin 和 text_end 直接定位（最准确）**
    if (segment && segment.text_begin > 0 && segment.text_end > segment.text_begin) {
      // 获取所有文本节点
      const walker = document.createTreeWalker(
        articleContentRef.current,
        NodeFilter.SHOW_TEXT,
        null
      );
      
      let charCount = 0;
      const targetStart = segment.text_begin;
      const targetEnd = segment.text_end;
      const matchedElements: HTMLElement[] = [];
      
      let node: Node | null;
      while ((node = walker.nextNode())) {
        const nodeText = node.textContent || '';
        const nodeStart = charCount;
        const nodeEnd = charCount + nodeText.length;
        
        // 检查这个文本节点是否与目标范围重叠
        if (nodeEnd > targetStart && nodeStart < targetEnd) {
          // 找到包含这个文本节点的父元素（段落）
          let parent = node.parentElement;
          while (parent && parent !== articleContentRef.current) {
            // 检查是否是块级元素
            if (['P', 'H1', 'H2', 'H3', 'H4', 'H5', 'H6', 'LI', 'BLOCKQUOTE'].includes(parent.tagName)) {
              if (!matchedElements.includes(parent)) {
                matchedElements.push(parent);
                parent.classList.add('highlight');
                
                // 添加指示器到第一个匹配的元素
                if (matchedElements.length === 1) {
                  const indicator = document.createElement('div');
                  indicator.className = 'current-indicator';
                  indicator.innerHTML = '▶';
                  (parent as HTMLElement).style.position = 'relative';
                  parent.insertBefore(indicator, parent.firstChild);
                }
                
                // 动态修正对齐
                fixHighlightAlignment(parent as HTMLElement);
              }
              break;
            }
            parent = parent.parentElement;
          }
        }
        
        charCount = nodeEnd;
      }
      
      if (matchedElements.length > 0) {
        scrollToHighlight();
        return; // 成功使用字符位置定位，直接返回
      }
    }

    // **方案2：文本匹配（作为备选方案）**
    // 预处理目标文本：提取核心内容（只保留汉字、字母、数字）
    const cleanTarget = text.trim().replace(/[^\u4e00-\u9fa5a-zA-Z0-9]/g, '').toLowerCase();
    if (cleanTarget.length < 3) {  // 降低最小长度要求，从 5 改为 3
      console.log(`⚠️ 文本太短，跳过: "${text}" (长度: ${cleanTarget.length})`);
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

      // 核心匹配逻辑 - 更宽松的匹配策略
      // 1. 完全包含匹配（最高优先级）
      if (cleanPara.includes(cleanTarget)) {
        const matchPosition = cleanPara.indexOf(cleanTarget);
        const matchRatio = matchPosition / Math.max(cleanPara.length, 1);
        const positionScore = matchRatio > 0.5 ? 0.1 : 0;
        matches.push({ element: para, index: index, score: 1.0 + positionScore, matchPosition: matchPosition });
      } 
      // 2. 目标文本包含段落（段落是目标的一部分）
      else if (cleanTarget.includes(cleanPara)) {
        matches.push({ element: para, index: index, score: 1.0 });
      } 
      // 3. 部分匹配：检查开头和结尾
      else {
        const minMatchLength = Math.max(5, Math.min(cleanTarget.length * 0.3, 20)); // 至少匹配 30% 或 5 个字符
        const start = cleanTarget.substring(0, Math.min(minMatchLength, cleanTarget.length));
        const end = cleanTarget.substring(Math.max(0, cleanTarget.length - minMatchLength));
        const hasStart = cleanPara.includes(start);
        const hasEnd = cleanPara.includes(end);
        
        if (hasStart || hasEnd) {
          let score = 0.5;
          if (hasStart && hasEnd) {
            score = 0.8;
          }
          matches.push({ element: para, index: index, score: score });
        }
        // 4. 模糊匹配：检查是否有足够的共同字符
        else {
          const commonChars = cleanTarget.split('').filter(char => cleanPara.includes(char)).length;
          const similarity = commonChars / Math.max(cleanTarget.length, cleanPara.length);
          if (similarity > 0.4) {  // 如果相似度超过 40%
            matches.push({ element: para, index: index, score: similarity * 0.6 });
          }
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
      
      // 调试信息
      if (segmentIndex % 10 === 0) {  // 每 10 个片段打印一次，避免日志过多
        console.log(`🔍 片段 ${segmentIndex}: 搜索 "${text.substring(0, 30)}..." 找到 ${matches.length} 个匹配`);
      }

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
          
          // 动态修正对齐
          fixHighlightAlignment(match.element as HTMLElement);
        });
        scrollToHighlight();
      }
    } else {
      // 如果没有找到匹配，记录调试信息
      if (segmentIndex % 20 === 0) {  // 每 20 个片段打印一次
        console.warn(`⚠️ 片段 ${segmentIndex}: 未找到匹配文本 "${text.substring(0, 50)}..."`);
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
        // 传递完整的 segment 对象，以便使用 text_begin/text_end
        highlightTextInContent(segment.text, segmentIndex, timelineData.length, segment);
        currentHighlightIndexRef.current = segmentIndex;
      }
    } else if (segmentIndex === -1 && currentTime > 0) {
      // 如果找不到对应的片段，可能是时间戳不匹配
      if (Math.floor(currentTime) % 5 === 0) {  // 每 5 秒打印一次
        console.warn(`⚠️ 时间 ${currentTime.toFixed(2)}s (${(currentTime * 1000).toFixed(0)}ms) 未找到对应片段`);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentTime, timelineData]);

  return {
    timelineData,
    removeHighlight,
  };
};

