import React, { useRef, useEffect, useState, useCallback } from 'react';
import ReactMarkdown from 'react-markdown';
import { useStore } from '../store/useStore';
import { useAudioHighlight } from '../hooks/useAudioHighlight';
import { AddVocabularyPopup } from './AddVocabularyPopup';

// 使用 React.memo 包装 Markdown 渲染部分，防止因 currentTime 变化导致重渲染
// 这对于保持手动添加的 DOM 高亮样式至关重要
const MarkdownContent = React.memo(({ content }: { content: string }) => {
    return (
        <>
            <ReactMarkdown
                components={{
                    h1: ({node, ...props}) => <h1 {...props} />,
                    h2: ({node, ...props}) => <h2 {...props} />,
                    h3: ({node, ...props}) => <h3 {...props} />,
                    p: ({node, ...props}) => <p {...props} />,
                    blockquote: ({node, ...props}) => <blockquote {...props} />,
                    ul: ({node, ...props}) => <ul {...props} />,
                    ol: ({node, ...props}) => <ol {...props} />,
                    li: ({node, ...props}) => <li {...props} />,
                    strong: ({node, ...props}) => <strong {...props} />,
                }}
            >
                {content}
            </ReactMarkdown>

            <div className="article-footer">
                <p>End of Article</p>
                <div className="divider">***</div>
            </div>
        </>
    );
});

interface ArticleViewerProps {
    onForceScrollToHighlight?: (fn: () => void) => void; // 传递定位函数给父组件
    onAutoScrollChange?: (enabled: boolean) => void; // 传递自动滚动状态给父组件
}

export const ArticleViewer: React.FC<ArticleViewerProps> = ({ 
    onForceScrollToHighlight,
    onAutoScrollChange 
}) => {
    const { currentArticle, currentTime } = useStore();
    const articleContentRef = useRef<HTMLDivElement>(null);
    
    // 选词添加生词本相关状态
    const [showVocabularyPopup, setShowVocabularyPopup] = useState(false);
    const [selectedText, setSelectedText] = useState('');
    const [selectionPosition, setSelectionPosition] = useState({ x: 0, y: 0 });
    const [selectionContext, setSelectionContext] = useState('');
    
    // 使用音频高亮 Hook
    // currentTime 变化会触发组件重渲染，但得益于 React.memo，MarkdownContent 不会重绘 DOM
    // 从而保留了 Hook 添加的高亮 class
    const { forceScrollToHighlight, autoScrollEnabled } = useAudioHighlight(currentTime, articleContentRef);
    
    // 获取选中文本的上下文（完整句子）
    const getSelectionContext = useCallback((selection: Selection): string => {
        if (!selection.rangeCount) return '';
        
        const range = selection.getRangeAt(0);
        const container = range.commonAncestorContainer;
        
        // 找到包含选中文本的段落或句子
        let parentElement = container.nodeType === Node.TEXT_NODE 
            ? container.parentElement 
            : container as Element;
        
        // 向上查找到段落级别
        while (parentElement && !['P', 'LI', 'BLOCKQUOTE', 'H1', 'H2', 'H3'].includes(parentElement.tagName)) {
            parentElement = parentElement.parentElement;
        }
        
        if (parentElement) {
            const text = parentElement.textContent || '';
            // 如果段落太长，只取选中文本附近的内容
            if (text.length > 200) {
                const selectedStr = selection.toString();
                const index = text.indexOf(selectedStr);
                if (index !== -1) {
                    const start = Math.max(0, index - 50);
                    const end = Math.min(text.length, index + selectedStr.length + 50);
                    return '...' + text.slice(start, end) + '...';
                }
            }
            return text;
        }
        
        return '';
    }, []);
    
    // 处理文本选择
    const handleMouseUp = useCallback(() => {
        const selection = window.getSelection();
        if (!selection || selection.isCollapsed) {
            return;
        }
        
        const text = selection.toString().trim();
        if (text.length === 0 || text.length > 500) {
            return;
        }
        
        // 获取选区位置
        const range = selection.getRangeAt(0);
        const rect = range.getBoundingClientRect();
        
        // 获取上下文
        const context = getSelectionContext(selection);
        
        setSelectedText(text);
        setSelectionPosition({
            x: rect.left + rect.width / 2,
            y: rect.bottom
        });
        setSelectionContext(context);
        setShowVocabularyPopup(true);
    }, [getSelectionContext]);
    
    // 关闭弹窗
    const handleClosePopup = useCallback(() => {
        setShowVocabularyPopup(false);
        setSelectedText('');
        // 清除选区
        window.getSelection()?.removeAllRanges();
    }, []);

    // 将定位函数和自动滚动状态传递给父组件
    useEffect(() => {
        if (onForceScrollToHighlight) {
            onForceScrollToHighlight(forceScrollToHighlight);
        }
    }, [forceScrollToHighlight, onForceScrollToHighlight]);

    useEffect(() => {
        if (onAutoScrollChange) {
            onAutoScrollChange(autoScrollEnabled);
        }
    }, [autoScrollEnabled, onAutoScrollChange]);

    if (!currentArticle) {
        return null;
    }

    if (!currentArticle.content) {
        return (
            <div className="article-body" id="articleContent" ref={articleContentRef}>
                <p>正在加载文章内容...</p>
            </div>
        );
    }

    return (
        <>
            <div 
                className="article-body" 
                id="articleContent" 
                ref={articleContentRef}
                onMouseUp={handleMouseUp}
            >
            <MarkdownContent content={currentArticle.content} />
        </div>
            
            {/* 添加生词本弹窗 */}
            {showVocabularyPopup && selectedText && (
                <AddVocabularyPopup
                    selectedText={selectedText}
                    position={selectionPosition}
                    articleId={currentArticle.id}
                    context={selectionContext}
                    onClose={handleClosePopup}
                    onSuccess={() => {
                        console.log('✅ 生词添加成功:', selectedText);
                    }}
                />
            )}
        </>
    );
};
