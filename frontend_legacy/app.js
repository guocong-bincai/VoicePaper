// 功能实现: VoicePaper音频同步阅读器核心逻辑
// 实现方案: 基于时间轴数据实现音频与文本的精确同步
// 影响范围: index.html中的所有交互功能
// 实现日期: 2025-11-25

class VoicePaper {
    constructor() {
        // DOM元素
        this.audioPlayer = document.getElementById('audioPlayer');
        this.playPauseBtn = document.getElementById('playPauseBtn');
        this.rewindBtn = document.getElementById('rewindBtn');
        this.forwardBtn = document.getElementById('forwardBtn');
        this.progressSlider = document.getElementById('progressSlider');
        this.currentTimeDisplay = document.getElementById('currentTime');
        this.totalTimeDisplay = document.getElementById('totalTime');
        this.articleContent = document.getElementById('articleContent');

        // 数据
        this.timelineData = null;
        this.markdownContent = '';
        this.currentHighlightIndex = -1;
        this.currentHighlightParaIndex = -1; // 记录当前高亮的段落索引位置
        this.isUserInteracting = false; // 标记用户是否正在交互进度条
        this.currentArticle = null; // 当前文章配置
        this.isPlayRequestPending = false; // BUG修复: 防止播放请求冲突

        // 初始化
        this.init();
    }

    async init() {
        try {
            // 1. 加载文章清单
            await this.loadManifest();

            // 2. 加载数据
            await this.loadTimelineData();
            await this.loadMarkdownContent();

            // 3. 渲染Markdown
            this.renderMarkdown();

            // 4. 绑定事件
            this.bindEvents();

            console.log('✅ VoicePaper初始化成功');
        } catch (error) {
            console.error('❌ 初始化失败:', error);
            alert('加载失败: ' + error.message);
        }
    }

    // 加载文章清单并确定当前文章
    async loadManifest() {
        try {
            const response = await fetch('../data/manifest.json');
            const manifest = await response.json();

            // 获取URL参数中的id
            const urlParams = new URLSearchParams(window.location.search);
            const articleId = urlParams.get('id');

            if (articleId) {
                this.currentArticle = manifest.articles.find(a => a.id === articleId);
            }

            // 如果没有指定ID或找不到，默认使用第一个
            if (!this.currentArticle && manifest.articles.length > 0) {
                this.currentArticle = manifest.articles[0];
            }

            if (!this.currentArticle) {
                throw new Error('未找到任何文章配置');
            }

            console.log('📚 当前加载文章:', this.currentArticle.title);

            // 更新页面标题
            document.title = `${this.currentArticle.title} | VoicePaper`;

            // 更新播放器标题
            const trackTitle = document.querySelector('.track-title');
            if (trackTitle) trackTitle.textContent = this.currentArticle.title;

            // 更新音频源 (添加时间戳防止缓存问题)
            const audioSrc = `../data/${this.currentArticle.audio}`;
            console.log('🎵 设置音频源:', audioSrc);
            this.audioPlayer.src = audioSrc;
            this.audioPlayer.preload = 'auto';
            this.audioPlayer.load(); // 显式加载

        } catch (error) {
            console.error('❌ 清单加载失败:', error);
            throw error;
        }
    }

    // 加载时间轴数据
    async loadTimelineData() {
        try {
            const response = await fetch(`../data/${this.currentArticle.titles}`);
            this.timelineData = await response.json();
            console.log('✅ 时间轴数据加载成功:', this.timelineData.length, '条');
        } catch (error) {
            console.error('❌ 时间轴数据加载失败:', error);
            throw error;
        }
    }

    // 加载Markdown内容
    async loadMarkdownContent() {
        try {
            const response = await fetch(`../data/${this.currentArticle.markdown}`);
            this.markdownContent = await response.text();
            console.log('✅ Markdown内容加载成功');
        } catch (error) {
            console.error('❌ Markdown内容加载失败:', error);
            throw error;
        }
    }

    // 渲染Markdown
    renderMarkdown() {
        // 配置marked.js选项
        marked.setOptions({
            breaks: true,        // 识别单个换行符
            gfm: true,          // 启用GitHub风格的Markdown
            headerIds: false,   // 不生成header ID
            mangle: false       // 不混淆邮箱地址
        });

        // 使用marked.js渲染Markdown
        const htmlContent = marked.parse(this.markdownContent);

        // 处理渲染后的内容，添加段落标识
        const tempDiv = document.createElement('div');
        tempDiv.innerHTML = htmlContent;

        // 为每个段落添加数据属性，便于后续匹配
        let textOffset = 0;
        const allNodes = tempDiv.querySelectorAll('p, h1, h2, h3, h4, h5, h6, li, blockquote');
        allNodes.forEach((node, index) => {
            const text = node.textContent;
            node.setAttribute('data-text-start', textOffset);
            node.setAttribute('data-text-end', textOffset + text.length);
            node.setAttribute('data-para-index', index);
            textOffset += text.length;
        });

        this.articleContent.innerHTML = tempDiv.innerHTML;

        // 为每个时间段的文本添加标记
        this.wrapTextSegments();
    }

    // 为文本段落添加标记,用于高亮
    wrapTextSegments() {
        // 为每个时间段创建映射
        this.timelineData.forEach((item, index) => {
            item.index = index;
            // 清理文本：去除多余空格
            item.cleanText = item.text.trim().replace(/\s+/g, ' ');
        });
    }

    // 根据音频时间查找对应的文本段
    findCurrentSegment(currentTime) {
        // 将秒转换为毫秒
        const currentTimeMs = currentTime * 1000;

        for (let i = 0; i < this.timelineData.length; i++) {
            const segment = this.timelineData[i];
            if (currentTimeMs >= segment.time_begin && currentTimeMs <= segment.time_end) {
                return i;
            }
        }
        return -1;
    }

    // 高亮当前文本段 - 适配新逻辑
    highlightSegment(index) {
        // 这个方法在新逻辑中被 updateProgress 直接调用 highlightTextInContent 替代了
        // 但为了保持兼容性，我们可以留空或者转发
    }

    // 在内容中高亮文本 - 优化匹配算法
    // BUG修复: 提高文本匹配精度，实现更准确的音频-文本同步
    // 修复策略: 1) 使用更精确的文本匹配 2) 优先匹配更靠后的位置 3) 考虑句子级别匹配
    // 影响范围: frontend/app.js:186-321
    // 修复日期: 2025-11-25
    highlightTextInContent(text, segmentIndex, totalSegments) {
        this.removeHighlight();
        this.removeCurrentIndicator();

        const containers = this.articleContent.querySelectorAll('.highlight-container');
        containers.forEach(c => c.classList.remove('highlight-container'));

        // 1. 预处理目标文本：提取核心内容（只保留汉字、字母、数字）
        // 这样可以忽略音标 []、标点、Markdown符号等所有干扰
        const cleanTarget = text.replace(/[^\u4e00-\u9fa5a-zA-Z0-9]/g, '').toLowerCase();

        if (cleanTarget.length < 5) return;

        const paragraphs = Array.from(this.articleContent.querySelectorAll('p, h1, h2, h3, h4, h5, h6, li, blockquote, div'));

        const matches = [];

        paragraphs.forEach((para, index) => {
            // 跳过包含子块级元素的容器，只关注最底层的文本节点容器
            if (para.querySelector('p, h1, h2, h3, h4, h5, h6, li')) return;

            const paraText = para.textContent || "";
            // 2. 同样的规则处理段落文本
            const cleanPara = paraText.replace(/[^\u4e00-\u9fa5a-zA-Z0-9]/g, '').toLowerCase();

            if (cleanPara.length < 5) return;

            // 3. 核心匹配逻辑：基于纯净文本的包含关系
            // 情况A: 段落包含目标 (段落比目标长，或者差不多)
            if (cleanPara.includes(cleanTarget)) {
                // BUG修复: 计算匹配位置，优先选择更靠后的匹配（因为音频是向前播放的）
                const matchPosition = cleanPara.indexOf(cleanTarget);
                const matchRatio = matchPosition / Math.max(cleanPara.length, 1); // 匹配位置在段落中的比例
                // 如果匹配在段落后半部分，提高分数（因为音频更可能读到这里）
                const positionScore = matchRatio > 0.5 ? 0.1 : 0;
                matches.push({ element: para, index: index, score: 1.0 + positionScore, matchPosition: matchPosition });
            }
            // 情况B: 目标包含段落 (目标比段落长，比如音频读了一大段)
            else if (cleanTarget.includes(cleanPara)) {
                matches.push({ element: para, index: index, score: 1.0 });
            }
            // 情况C: 模糊匹配 (处理跨段落或只有部分重叠的情况)
            else {
                // 取目标的前30个有效字符（增加长度以提高匹配精度）
                const start = cleanTarget.substring(0, Math.min(30, cleanTarget.length));
                // 取目标的后30个有效字符
                const end = cleanTarget.substring(Math.max(0, cleanTarget.length - 30));

                // BUG修复: 同时检查开头和结尾，如果都匹配则提高分数
                const hasStart = cleanPara.includes(start);
                const hasEnd = cleanPara.includes(end);
                
                if (hasStart || hasEnd) {
                    let score = 0.5;
                    // 如果开头和结尾都匹配，说明匹配度更高
                    if (hasStart && hasEnd) {
                        score = 0.8;
                    }
                    matches.push({ element: para, index: index, score: score });
                }
            }
        });

        if (matches.length > 0) {
            // BUG修复: 先按分数排序，再按索引排序，优先选择分数高且位置靠后的匹配
            matches.sort((a, b) => {
                if (Math.abs(a.score - b.score) > 0.1) {
                    return b.score - a.score; // 分数高的优先
                }
                return a.index - b.index; // 分数相同时，按索引排序
            });

            let bestMatches = matches;

            // 策略优化：多重匹配消歧
            if (matches.length > 1) {
                // 1. 优先使用上下文 (Sequential Playback) - 这是最重要的策略
                // 如果当前有高亮段落，且存在位于其后的匹配项，优先考虑这些
                if (this.currentHighlightParaIndex >= 0) {
                    const subsequentMatches = matches.filter(m => m.index > this.currentHighlightParaIndex);
                    if (subsequentMatches.length > 0) {
                        // BUG修复: 在后续匹配中，优先选择分数最高的，而不是第一个
                        // 这样可以更准确地匹配到实际播放位置
                        const bestSubsequent = subsequentMatches.reduce((prev, curr) => {
                            return curr.score > prev.score ? curr : prev;
                        });
                        bestMatches = [bestSubsequent];
                    } else {
                        // 如果没有后续匹配，可能循环了或者逻辑异常，回退到比率匹配
                        bestMatches = matches; // 暂时重置，让下面的逻辑处理
                    }
                }

                // 2. 使用位置比率 (Ratio Heuristic) - 适用于 Seek 和无上下文情况
                // 如果上面的逻辑没有锁定唯一匹配，或者我们处于 Seek 模式（currentHighlightParaIndex == -1）
                if (bestMatches.length > 1 && typeof segmentIndex === 'number' && typeof totalSegments === 'number') {
                    const audioProgress = segmentIndex / totalSegments;
                    const totalParas = paragraphs.length;

                    // BUG修复: 优先选择位置靠后的匹配（因为音频是向前播放的）
                    // 找到与当前音频进度最接近且位置靠后的段落位置
                    bestMatches = [matches.reduce((prev, curr) => {
                        const prevRatio = prev.index / totalParas;
                        const currRatio = curr.index / totalParas;
                        const prevDiff = Math.abs(prevRatio - audioProgress);
                        const currDiff = Math.abs(currRatio - audioProgress);
                        
                        // 如果差值相近（小于0.05），优先选择位置靠后的
                        if (Math.abs(prevDiff - currDiff) < 0.05) {
                            return curr.index > prev.index ? curr : prev;
                        }
                        return currDiff < prevDiff ? curr : prev;
                    })];
                }
            }

            // 4. "填补空缺" (Fill the Gap) 逻辑
            // 如果匹配了第 5 个和第 7 个段落，那么第 6 个段落很有可能也应该被高亮
            if (bestMatches.length >= 2) {
                const firstIndex = bestMatches[0].index;
                const lastIndex = bestMatches[bestMatches.length - 1].index;

                // 如果跨度不太大（比如中间只隔了不到 5 个段落），就填补中间的
                if (lastIndex - firstIndex < 5) {
                    for (let i = firstIndex + 1; i < lastIndex; i++) {
                        // 检查这个索引是否已经在匹配列表中
                        const exists = bestMatches.find(m => m.index === i);
                        if (!exists) {
                            // 获取对应的元素
                            const gapPara = paragraphs[i];
                            // 只有当它不是空元素时才添加
                            if (gapPara && gapPara.textContent.trim().length > 0) {
                                console.log('🔧 自动填补中间段落:', i);
                                bestMatches.push({ element: gapPara, index: i, score: 0.5 });
                            }
                        }
                    }
                    // 重新排序
                    bestMatches.sort((a, b) => a.index - b.index);
                }
            }

            if (bestMatches.length > 0) {
                // BUG修复: 如果有多个匹配，优先选择最靠后的（因为音频是向前播放的）
                // 这样可以确保高亮跟随音频播放进度，而不是停留在前面的匹配
                const bestMatch = bestMatches.length > 1 
                    ? bestMatches.reduce((prev, curr) => {
                        // 优先选择分数高的，分数相同时选择位置靠后的
                        if (Math.abs(curr.score - prev.score) > 0.1) {
                            return curr.score > prev.score ? curr : prev;
                        }
                        return curr.index > prev.index ? curr : prev;
                    })
                    : bestMatches[0];

                // 记录当前高亮的段落索引（使用最佳匹配的索引）
                this.currentHighlightParaIndex = bestMatch.index;

                // 高亮最佳匹配及其相邻的匹配（如果有）
                bestMatches.forEach((match, i) => {
                    match.element.classList.add('highlight');

                    const parent = match.element.parentElement;
                    if (parent && parent !== this.articleContent) {
                        parent.classList.add('highlight-container');
                    }

                    // 只在最佳匹配上显示指示器
                    if (match.index === bestMatch.index) {
                        const indicator = document.createElement('div');
                        indicator.className = 'current-indicator';
                        indicator.innerHTML = '▶';
                        match.element.style.position = 'relative';
                        match.element.insertBefore(indicator, match.element.firstChild);
                    }
                });
                this.scrollToHighlight();
            }
        }
    }

    // 移除高亮
    removeHighlight() {
        const highlights = this.articleContent.querySelectorAll('.highlight');
        highlights.forEach(element => {
            element.classList.remove('highlight');
        });
        // 注意：不清除 currentHighlightParaIndex，保持上下文以便下次匹配
    }

    // 移除当前位置指示器
    removeCurrentIndicator() {
        const indicators = this.articleContent.querySelectorAll('.current-indicator');
        indicators.forEach(indicator => {
            indicator.remove();
        });
    }

    // 滚动到高亮位置 - 优化体验
    scrollToHighlight() {
        const highlight = this.articleContent.querySelector('.highlight');
        if (highlight) {
            // 计算位置，使高亮块位于屏幕中间偏上位置，阅读体验更好
            const rect = highlight.getBoundingClientRect();
            const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
            const targetTop = scrollTop + rect.top - (window.innerHeight / 3); // 位于视口 1/3 处

            window.scrollTo({
                top: targetTop,
                behavior: 'smooth'
            });
        }
    }

    // 绑定事件
    bindEvents() {
        // 播放/暂停按钮
        this.playPauseBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            this.togglePlayPause();
        });

        // 监听音频状态事件，确保UI与实际状态同步
        this.audioPlayer.addEventListener('play', () => {
            this.isPlayRequestPending = false; // BUG修复: 确保播放事件触发时重置标志位
            this.updatePlayState(true);
        });
        this.audioPlayer.addEventListener('pause', () => {
            this.isPlayRequestPending = false; // BUG修复: 确保暂停事件触发时重置标志位
            this.updatePlayState(false);
        });
        this.audioPlayer.addEventListener('waiting', () => {
            const statusEl = document.querySelector('.track-status');
            if (statusEl) statusEl.textContent = '缓冲中...';
        });
        this.audioPlayer.addEventListener('playing', () => {
            const statusEl = document.querySelector('.track-status');
            if (statusEl) statusEl.textContent = '正在朗读...';
        });
        this.audioPlayer.addEventListener('error', (e) => {
            console.error("音频播放出错:", this.audioPlayer.error);
            const statusEl = document.querySelector('.track-status');
            if (statusEl) statusEl.textContent = '播放出错';
            alert('音频加载失败，请检查网络或文件是否存在');
        });

        // 后退10秒
        this.rewindBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            this.skip(-10);
        });

        // 前进10秒
        this.forwardBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            this.skip(10);
        });

        // 进度条交互优化
        // 1. 开始拖动/点击
        const startInteraction = () => {
            this.isUserInteracting = true;
        };

        // 2. 结束拖动/点击
        const endInteraction = () => {
            this.isUserInteracting = false;
        };

        this.progressSlider.addEventListener('mousedown', startInteraction);
        this.progressSlider.addEventListener('touchstart', startInteraction);

        this.progressSlider.addEventListener('mouseup', endInteraction);
        this.progressSlider.addEventListener('touchend', endInteraction);

        // 3. 拖动中：只更新视觉，不seek
        this.progressSlider.addEventListener('input', (e) => {
            const percentage = e.target.value;
            // 更新视觉进度条
            const progressFill = document.querySelector('.progress-fill');
            if (progressFill) {
                progressFill.style.width = `${percentage}%`;
            }
            // 更新时间显示
            const duration = this.audioPlayer.duration;
            if (duration > 0) {
                const time = (percentage / 100) * duration;
                this.currentTimeDisplay.textContent = this.formatTime(time);
            }
        });

        // 4. 拖动结束/点击松开：执行seek
        this.progressSlider.addEventListener('change', (e) => {
            this.seekTo(e.target.value);
        });

        // 音频时间更新
        this.audioPlayer.addEventListener('timeupdate', () => this.updateProgress());

        // 音频加载完成
        this.audioPlayer.addEventListener('loadedmetadata', () => this.updateTotalTime());

        // 音频播放结束
        this.audioPlayer.addEventListener('ended', () => this.onPlaybackEnded());

        // 键盘快捷键
        document.addEventListener('keydown', (e) => this.handleKeyPress(e));
    }

    // 播放/暂停切换
    // BUG修复: 防止播放请求冲突导致的AbortError
    // 修复策略: 添加状态标志位，确保播放请求完成前不会重复调用
    // 影响范围: frontend/app.js:445-458
    // 修复日期: 2025-11-25
    async togglePlayPause() {
        // 如果正在处理播放请求，忽略新的请求
        if (this.isPlayRequestPending) {
            return;
        }

        if (this.audioPlayer.paused) {
            this.isPlayRequestPending = true;
            try {
                await this.audioPlayer.play();
                // UI更新将由 'play'/'playing' 事件监听器处理
            } catch (error) {
                // AbortError是预期的，当播放请求被中断时会出现，可以安全忽略
                if (error.name === 'AbortError') {
                    console.log('播放请求被中断（这是正常的）');
                } else {
                    console.error("播放请求失败:", error);
                    // 可以在这里处理自动播放策略限制等问题
                }
            } finally {
                // 无论成功或失败，都重置标志位
                this.isPlayRequestPending = false;
            }
        } else {
            // 暂停操作不需要等待，直接执行
            this.audioPlayer.pause();
            // UI更新将由 'pause' 事件监听器处理
        }
    }

    // 更新播放状态UI
    updatePlayState(isPlaying) {
        const playIcon = this.playPauseBtn.querySelector('.play-icon');
        const pauseIcon = this.playPauseBtn.querySelector('.pause-icon');
        const statusEl = document.querySelector('.track-status');
        const iconEl = document.querySelector('.track-icon');

        if (isPlaying) {
            if (playIcon) playIcon.style.display = 'none';
            if (pauseIcon) pauseIcon.style.display = 'block';
            if (statusEl) statusEl.textContent = '正在朗读...';
            if (iconEl) iconEl.classList.add('playing');
        } else {
            if (playIcon) playIcon.style.display = 'block';
            if (pauseIcon) pauseIcon.style.display = 'none';
            if (statusEl) statusEl.textContent = '点击播放'; // 暂停时显示引导文案
            if (iconEl) iconEl.classList.remove('playing');
        }
    }

    // 跳转(前进/后退)
    skip(seconds) {
        const duration = this.audioPlayer.duration;
        let currentTime = this.audioPlayer.currentTime;

        // 确保 duration 有效
        if (!isFinite(duration) || duration === 0) {
            console.warn('⚠️ 无法跳转：音频时长无效', duration);
            return;
        }

        // 确保 currentTime 有效
        if (!isFinite(currentTime)) {
            currentTime = 0;
        }

        console.log(`⏩ 跳转前: ${currentTime.toFixed(2)}s, 目标偏移: ${seconds}s`);

        let newTime = currentTime + seconds;

        // 边界检查
        if (newTime < 0) newTime = 0;
        if (newTime > duration) newTime = duration;

        // 执行跳转
        this.audioPlayer.currentTime = newTime;
        console.log(`✅ 跳转后: ${this.audioPlayer.currentTime.toFixed(2)}s`);

        // 重置高亮位置，因为用户跳转了，需要重新匹配
        this.currentHighlightParaIndex = -1;
        this.currentHighlightIndex = -1;
    }

    // 跳转到指定位置
    seekTo(percentage) {
        const duration = this.audioPlayer.duration;
        if (!isFinite(duration) || duration === 0) return;

        const time = (percentage / 100) * duration;
        this.audioPlayer.currentTime = time;

        // 更新视觉进度条（因为在拖动时 updateProgress 被暂停了）
        const progressFill = document.querySelector('.progress-fill');
        if (progressFill) {
            progressFill.style.width = `${percentage}%`;
        }

        // 更新时间显示
        this.currentTimeDisplay.textContent = this.formatTime(time);

        // 重置高亮位置，因为用户跳转了，需要重新匹配
        this.currentHighlightParaIndex = -1;
        this.currentHighlightIndex = -1;
    }

    // 更新进度
    updateProgress() {
        // 如果用户正在交互，暂停自动更新进度条位置，避免冲突
        if (this.isUserInteracting) return;

        const currentTime = this.audioPlayer.currentTime;
        const duration = this.audioPlayer.duration;

        // 更新进度条
        if (duration > 0) {
            const percentage = (currentTime / duration) * 100;
            this.progressSlider.value = percentage;

            // 更新自定义进度条的视觉宽度
            const progressFill = document.querySelector('.progress-fill');
            if (progressFill) {
                progressFill.style.width = `${percentage}%`;
            }
        }

        // 更新时间显示
        this.currentTimeDisplay.textContent = this.formatTime(currentTime);

        // 更新文本高亮
        const segmentIndex = this.findCurrentSegment(currentTime);
        if (segmentIndex !== -1 && segmentIndex !== this.currentHighlightIndex) {
             const segment = this.timelineData[segmentIndex];
             if (segment) {
                 // BUG修复: 添加调试日志，帮助定位匹配问题
                 if (segmentIndex % 10 === 0 || segmentIndex - this.currentHighlightIndex > 5) {
                     console.log(`🎯 匹配段 ${segmentIndex}/${this.timelineData.length}: "${segment.text.substring(0, 30)}..."`);
                 }
                 // 传递 segmentIndex 和 totalSegments 用于消歧
                 this.highlightTextInContent(segment.text, segmentIndex, this.timelineData.length);
                 this.currentHighlightIndex = segmentIndex;
             }
        }
    }

    // 更新总时长
    updateTotalTime() {
        this.totalTimeDisplay.textContent = this.formatTime(this.audioPlayer.duration);
    }

    // 播放结束
    onPlaybackEnded() {
        const playIcon = this.playPauseBtn.querySelector('.play-icon');
        const pauseIcon = this.playPauseBtn.querySelector('.pause-icon');

        if (playIcon) playIcon.style.display = 'block';
        if (pauseIcon) pauseIcon.style.display = 'none';

        this.removeHighlight();
        this.currentHighlightIndex = -1;
        this.currentHighlightParaIndex = -1; // 重置段落索引

        const statusEl = document.querySelector('.track-status');
        if (statusEl) statusEl.textContent = '播放结束';
    }

    // 格式化时间(秒 -> MM:SS)
    formatTime(seconds) {
        if (isNaN(seconds)) return '00:00';

        const mins = Math.floor(seconds / 60);
        const secs = Math.floor(seconds % 60);
        return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
    }

    // 键盘快捷键
    handleKeyPress(e) {
        switch(e.code) {
            case 'Space':
                e.preventDefault();
                this.togglePlayPause();
                break;
            case 'ArrowLeft':
                e.preventDefault();
                this.skip(-10);
                break;
            case 'ArrowRight':
                e.preventDefault();
                this.skip(10);
                break;
        }
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    new VoicePaper();
});