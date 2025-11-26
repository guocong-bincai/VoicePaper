import React, { useRef, useEffect, useState, useMemo } from 'react';
import ReactMarkdown from 'react-markdown';
import { useStore } from '../store/useStore';
import { useAudioHighlight } from '../hooks/useAudioHighlight';

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

export const ArticleViewer: React.FC = () => {
    const { currentArticle, currentTime } = useStore();
    const articleContentRef = useRef<HTMLDivElement>(null);
    
    // 使用音频高亮 Hook
    // currentTime 变化会触发组件重渲染，但得益于 React.memo，MarkdownContent 不会重绘 DOM
    // 从而保留了 Hook 添加的高亮 class
    useAudioHighlight(currentTime, articleContentRef);

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
        <div className="relative">
            <div className="article-body" id="articleContent" ref={articleContentRef}>
                <MarkdownContent content={currentArticle.content} />
            </div>
        </div>
    );
};
