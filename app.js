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

        // 初始化
        this.init();
    }

    async init() {
        try {
            // 加载数据
            await this.loadTimelineData();
            await this.loadMarkdownContent();
            
            // 渲染Markdown
            this.renderMarkdown();
            
            // 绑定事件
            this.bindEvents();
            
            console.log('✅ VoicePaper初始化成功');
        } catch (error) {
            console.error('❌ 初始化失败:', error);
            alert('加载失败,请检查文件路径');
        }
    }

    // 加载时间轴数据
    async loadTimelineData() {
        try {
            const response = await fetch('1925118537643336511_202511251800_337964855271998_337969297908224/content-1925118537643336511_202511251800_337964855271998_337969297908224.titles');
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
            const response = await fetch('1.md');
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

    // 在内容中高亮文本 - 终极匹配算法
    highlightTextInContent(text) {
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
                matches.push({ element: para, index: index, score: 1.0 });
            }
            // 情况B: 目标包含段落 (目标比段落长，比如音频读了一大段)
            else if (cleanTarget.includes(cleanPara)) {
                matches.push({ element: para, index: index, score: 1.0 });
            }
            // 情况C: 模糊匹配 (处理跨段落或只有部分重叠的情况)
            else {
                // 取目标的前20个有效字符
                const start = cleanTarget.substring(0, 20);
                // 取目标的后20个有效字符
                const end = cleanTarget.substring(Math.max(0, cleanTarget.length - 20));
                
                if (cleanPara.includes(start) || cleanPara.includes(end)) {
                    matches.push({ element: para, index: index, score: 0.5 });
                }
            }
        });

        if (matches.length > 0) {
            matches.sort((a, b) => a.index - b.index);
            
            // 过滤逻辑：如果是多重匹配，只保留得分高的，或者连续的
            const bestMatches = matches; // 这里简化策略，只要匹配上都高亮，宁可多亮不可少亮

            if (bestMatches.length > 0) {
                bestMatches.forEach((match, i) => {
                    match.element.classList.add('highlight');
                    
                    const parent = match.element.parentElement;
                    if (parent && parent !== this.articleContent) {
                        parent.classList.add('highlight-container');
                    }

                    if (i === 0) {
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
        this.playPauseBtn.addEventListener('click', () => this.togglePlayPause());

        // 后退10秒
        this.rewindBtn.addEventListener('click', () => this.skip(-10));

        // 前进10秒
        this.forwardBtn.addEventListener('click', () => this.skip(10));

        // 进度条拖动
        this.progressSlider.addEventListener('input', (e) => this.seekTo(e.target.value));

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
    togglePlayPause() {
        const playIcon = this.playPauseBtn.querySelector('.play-icon');
        const pauseIcon = this.playPauseBtn.querySelector('.pause-icon');

        if (this.audioPlayer.paused) {
            this.audioPlayer.play();
            // 切换图标
            if (playIcon) playIcon.style.display = 'none';
            if (pauseIcon) pauseIcon.style.display = 'block';
            
            // 添加正在播放的状态样式
            const statusEl = document.querySelector('.track-status');
            if (statusEl) statusEl.textContent = '正在朗读...';
            
            const iconEl = document.querySelector('.track-icon');
            if (iconEl) iconEl.classList.add('playing');
        } else {
            this.audioPlayer.pause();
            // 切换图标
            if (playIcon) playIcon.style.display = 'block';
            if (pauseIcon) pauseIcon.style.display = 'none';
            
            const statusEl = document.querySelector('.track-status');
            if (statusEl) statusEl.textContent = '已暂停';
            
            const iconEl = document.querySelector('.track-icon');
            if (iconEl) iconEl.classList.remove('playing');
        }
    }

    // 跳转(前进/后退)
    skip(seconds) {
        const duration = this.audioPlayer.duration;
        const currentTime = this.audioPlayer.currentTime;
        
        // 确保 duration 有效
        if (!isFinite(duration) || duration === 0) {
            console.warn('⚠️ 无法跳转：音频时长无效', duration);
            return;
        }
        
        console.log(`⏩ 跳转前: ${currentTime.toFixed(2)}s, 目标偏移: ${seconds}s`);
        
        let newTime = currentTime + seconds;
        
        // 边界检查
        if (newTime < 0) newTime = 0;
        if (newTime > duration) newTime = duration;
        
        // 执行跳转
        this.audioPlayer.currentTime = newTime;
        console.log(`✅ 跳转后: ${this.audioPlayer.currentTime.toFixed(2)}s`);
    }

    // 跳转到指定位置
    seekTo(percentage) {
        const time = (percentage / 100) * this.audioPlayer.duration;
        this.audioPlayer.currentTime = time;
    }

    // 更新进度
    updateProgress() {
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
                 this.highlightTextInContent(segment.text);
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

