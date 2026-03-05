import React, { useState, useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import { addVocabulary, lookupWord, getPhonetic, getPronunciationAudio, getPrimaryMeaning, playPronunciation, translateToChineseAPI } from '../services/api';
import type { DictionaryResult } from '../services/api';
import { analyzeWord, type MorphemeAnalysis } from '../data/morphemes';
import { findCollocations, type Collocation } from '../data/collocations';
import './AddVocabularyPopup.css';

// 功能实现: 生词本添加成功音效
// 实现日期: 2025-01-27
const successSound = new Audio('/sounds/right.mp3');

// 词性英文转中文映射
const partOfSpeechMap: Record<string, string> = {
    'noun': '名词',
    'verb': '动词',
    'adjective': '形容词',
    'adverb': '副词',
    'pronoun': '代词',
    'preposition': '介词',
    'conjunction': '连词',
    'interjection': '感叹词',
    'determiner': '限定词',
    'article': '冠词',
    'numeral': '数词',
    'participle': '分词',
    'auxiliary': '助动词',
    'modal': '情态动词',
    'phrasal verb': '短语动词',
    'transitive verb': '及物动词',
    'intransitive verb': '不及物动词',
    'linking verb': '系动词',
    'abbreviation': '缩写',
    'prefix': '前缀',
    'suffix': '后缀',
    'exclamation': '感叹词',
};

// 将英文词性转换为中文
const toChinesePartOfSpeech = (pos: string): string => {
    const lowerPos = pos.toLowerCase().trim();
    return partOfSpeechMap[lowerPos] || pos;
};

interface AddVocabularyPopupProps {
    selectedText: string;
    position: { x: number; y: number };
    articleId?: number;
    context?: string;  // 上下文（完整句子）
    onClose: () => void;
    onSuccess?: () => void;
}

export const AddVocabularyPopup: React.FC<AddVocabularyPopupProps> = ({
    selectedText,
    position: _position,  // 不再使用，弹窗居中显示
    articleId,
    context,
    onClose,
    onSuccess
}) => {
    const [type, setType] = useState<'word' | 'phrase' | 'sentence'>('word');
    const [meaning, setMeaning] = useState('');
    const [phonetic, setPhonetic] = useState('');
    const [example, setExample] = useState('');
    const [exampleTranslation, setExampleTranslation] = useState('');
    const [loading, setLoading] = useState(false);
    const [lookingUp, setLookingUp] = useState(false);
    const [error, setError] = useState('');
    const [alreadyExists, setAlreadyExists] = useState(false);  // 单词已存在
    const [success, setSuccess] = useState(false);
    const [audioUrl, setAudioUrl] = useState('');
    const [dictResult, setDictResult] = useState<DictionaryResult | null>(null);
    const [morphemeAnalysis, setMorphemeAnalysis] = useState<MorphemeAnalysis | null>(null);
    const [collocations, setCollocations] = useState<Collocation[]>([]);
    const [selectedCollocation, setSelectedCollocation] = useState<Collocation | null>(null);
    const popupRef = useRef<HTMLDivElement>(null);

    // 根据选中文本长度自动判断类型，并查询词典/翻译
    useEffect(() => {
        const wordCount = selectedText.trim().split(/\s+/).length;
        if (wordCount === 1) {
            setType('word');
            // 自动查询单词释义和中文翻译
            lookupWordMeaning(selectedText.trim());
        } else if (wordCount <= 5) {
            setType('phrase');
            // 短语：翻译成中文
            translateText(selectedText.trim());
        } else {
            setType('sentence');
            // 句子：翻译成中文
            translateText(selectedText.trim());
        }
    }, [selectedText]);

    // 翻译成中文（用于短语和句子）
    const translateText = async (text: string) => {
        setLookingUp(true);
        try {
            const translation = await translateToChineseAPI(text);
            if (translation) {
                setMeaning(translation);
            }
        } catch (err) {
            console.warn('翻译失败:', err);
        } finally {
            setLookingUp(false);
        }
    };

    // 查询单词释义
    const lookupWordMeaning = async (word: string) => {
        setLookingUp(true);
        try {
            // 并行请求：英文词典 + 中文翻译
            const [dictResultData, chineseTranslation] = await Promise.all([
                lookupWord(word),
                translateToChineseAPI(word)
            ]);

            if (dictResultData) {
                setDictResult(dictResultData);
                // 自动填充音标
                const phoneticText = getPhonetic(dictResultData);
                if (phoneticText) {
                    setPhonetic(phoneticText);
                }
                // 获取发音URL
                const audio = getPronunciationAudio(dictResultData);
                if (audio) {
                    setAudioUrl(audio);
                }
                // 获取词性
                const primaryMeaning = getPrimaryMeaning(dictResultData);
                // 组合中文翻译和中文词性
                if (chineseTranslation) {
                    // 优先显示中文翻译，附带中文词性
                    const partOfSpeech = dictResultData.meanings?.[0]?.partOfSpeech || '';
                    const chinesePos = partOfSpeech ? toChinesePartOfSpeech(partOfSpeech) : '';
                    setMeaning(chinesePos ? `[${chinesePos}] ${chineseTranslation}` : chineseTranslation);
                } else if (primaryMeaning) {
                    setMeaning(primaryMeaning);
                }

                // 获取例句
                if (dictResultData.meanings && dictResultData.meanings.length > 0) {
                    for (const m of dictResultData.meanings) {
                        if (m.definitions && m.definitions.length > 0) {
                            for (const def of m.definitions) {
                                if (def.example) {
                                    setExample(def.example);
                                    // 翻译例句
                                    translateToChineseAPI(def.example).then(trans => {
                                        if (trans) {
                                            setExampleTranslation(trans);
                                        }
                                    });
                                    break;
                                }
                            }
                            if (example) break;
                        }
                    }
                }
            } else if (chineseTranslation) {
                // 如果词典没找到，至少显示中文翻译
                setMeaning(chineseTranslation);
            }

            // 分析词根词缀
            const analysis = analyzeWord(word);
            if (analysis.foundMorphemes.length > 0) {
                setMorphemeAnalysis(analysis);
            }

            // 检测固定搭配（如果有上下文）
            if (context) {
                const foundCollocations = findCollocations(word, context);
                if (foundCollocations.length > 0) {
                    setCollocations(foundCollocations);
                }
            }
        } catch (err) {
            console.warn('查询词典失败:', err);
        } finally {
            setLookingUp(false);
        }
    };

    // 播放发音
    const handlePlayAudio = () => {
        if (audioUrl) {
            playPronunciation(audioUrl);
        }
    };

    // 点击外部关闭
    useEffect(() => {
        const handleClickOutside = (e: MouseEvent) => {
            const target = e.target;
            if (popupRef.current && target instanceof Node && !popupRef.current.contains(target)) {
                onClose();
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, [onClose]);

    // 快捷键：ESC关闭，Enter保存
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                onClose();
            } else if (e.key === 'Enter' && !e.shiftKey && !loading) {
                // Enter键保存（Shift+Enter换行时不触发）
                e.preventDefault();
                handleQuickAdd();
            }
        };
        document.addEventListener('keydown', handleKeyDown);
        return () => document.removeEventListener('keydown', handleKeyDown);
    }, [onClose, loading]);

    // 阻止滚动穿透：弹窗打开时禁止外层滚动
    useEffect(() => {
        const originalStyle = document.body.style.overflow;
        document.body.style.overflow = 'hidden';
        return () => {
            document.body.style.overflow = originalStyle;
        };
    }, []);

    const handleQuickAdd = async () => {
        if (!selectedText.trim()) return;

        setLoading(true);
        setError('');
        console.log('📝 开始添加生词:', selectedText);

        try {
            // 准备词根词缀数据（如果有，且未选择固定搭配时）
            let morphemesJson: string | undefined;
            if (!selectedCollocation && morphemeAnalysis && morphemeAnalysis.foundMorphemes.length > 0) {
                // 将词根词缀分析结果转为简化的JSON格式
                const morphemesData = morphemeAnalysis.foundMorphemes.map(item => ({
                    type: item.position,
                    morpheme: item.morpheme.morpheme,
                    meaning: item.morpheme.meaning,
                    origin: item.morpheme.origin
                }));
                morphemesJson = JSON.stringify(morphemesData);
            }

            // 如果选择了固定搭配，则保存搭配短语；否则保存原始选中文本
            const contentToSave = selectedCollocation ? selectedCollocation.phrase : selectedText.trim();
            const meaningToSave = selectedCollocation ? selectedCollocation.meaning : meaning.trim();
            const typeToSave = selectedCollocation ? 'phrase' : type;

            console.log('Sending vocabulary data:', {
                type: typeToSave,
                content: contentToSave,
                meaning: meaningToSave
            });

            await addVocabulary({
                type: typeToSave,
                content: contentToSave,
                meaning: meaningToSave || undefined,
                phonetic: selectedCollocation ? undefined : (phonetic.trim() || undefined),  // 搭配短语不保存音标
                example: example.trim() || undefined,
                example_translation: exampleTranslation.trim() || undefined,
                context: context || undefined,
                article_id: articleId,
                morphemes: morphemesJson,
                source: articleId ? 'article' : 'manual', // 如果有 articleId，来源是 article；否则是 manual
            });

            console.log('✅ 添加成功');
            setSuccess(true);
            // 功能实现: 播放成功音效 ding~
            // 实现日期: 2025-01-27
            try {
                successSound.currentTime = 0;
                successSound.play();
            } catch (e) {
                console.warn('播放成功音效失败:', e);
            }
            setTimeout(() => {
                onSuccess?.();
                onClose();
            }, 1200);  // 延长显示时间让用户看到特效
        } catch (err: any) {
            console.error('❌ 添加失败:', err);
            const errorMsg = err.response?.data?.error || '添加失败';
            // 检测是否是"已存在"错误
            if (errorMsg.includes('已在生词本') || errorMsg.includes('已存在')) {
                setAlreadyExists(true);
            } else {
                setError(errorMsg);
                // 强制提示，确保用户知道发生了什么
                alert(`添加失败: ${errorMsg}\n请检查网络连接或稍后再试。`);
            }
        } finally {
            setLoading(false);
        }
    };

    const renderContent = () => {
        if (success) {
            return (
                <div
                    ref={popupRef}
                    className="add-vocabulary-popup success"
                >
                    {/* 彩色碎片动画 */}
                    <div className="confetti-burst">
                        {[...Array(20)].map((_, i) => (
                            <div
                                key={i}
                                className={`confetti-piece confetti-${i % 5}`}
                                style={{
                                    '--angle': `${(i * 18)}deg`,
                                    '--distance': `${60 + Math.random() * 40}px`,
                                    '--delay': `${i * 0.02}s`
                                } as React.CSSProperties}
                            />
                        ))}
                    </div>
                    <div className="success-content">
                        <div className="success-icon-wrapper">
                            <span className="success-icon">✓</span>
                        </div>
                        <span className="success-text">已加入生词本</span>
                    </div>
                </div>
            );
        }

        return (
            <div
                ref={popupRef}
                className="add-vocabulary-popup"
            >
                <div className="popup-header">
                    <span className="popup-title">添加到生词本</span>
                    <button className="close-btn" onClick={onClose} title="关闭">
                        <svg viewBox="0 0 24 24" width="20" height="20" stroke="currentColor" strokeWidth="3" fill="none">
                            <line x1="18" y1="6" x2="6" y2="18"></line>
                            <line x1="6" y1="6" x2="18" y2="18"></line>
                        </svg>
                    </button>
                </div>

                {/* 已存在提示 - 显示在顶部 */}
                {alreadyExists && (
                    <div className="already-exists-notice">
                        <span className="notice-icon">⚠️</span>
                        <span>该词已在生词本中</span>
                    </div>
                )}

                <div className="popup-content">
                    {/* 选中的文本 */}
                    <div className="selected-text">
                        <span className="text-label">选中内容</span>
                        <div className="text-content-row">
                            <span className="text-content">{selectedText}</span>
                            {/* 发音按钮 */}
                            {audioUrl && (
                                <button
                                    className="play-audio-btn"
                                    onClick={handlePlayAudio}
                                    title="播放发音"
                                >
                                    <img src="/icon/laba.png" alt="播放" className="icon-img" />
                                </button>
                            )}
                            {lookingUp && <span className="looking-up">查询中...</span>}
                        </div>
                    </div>

                    {/* 类型选择 */}
                    <div className="type-selector">
                        <button
                            className={`type-btn ${type === 'word' ? 'active' : ''}`}
                            onClick={() => setType('word')}
                        >
                            单词
                        </button>
                        <button
                            className={`type-btn ${type === 'phrase' ? 'active' : ''}`}
                            onClick={() => setType('phrase')}
                        >
                            短语
                        </button>
                        <button
                            className={`type-btn ${type === 'sentence' ? 'active' : ''}`}
                            onClick={() => setType('sentence')}
                        >
                            句子
                        </button>
                    </div>

                    {/* 音标（仅单词显示） */}
                    {type === 'word' && (
                        <div className="input-group phonetic-group">
                            <input
                                type="text"
                                placeholder="音标（自动识别）"
                                value={phonetic}
                                onChange={(e) => setPhonetic(e.target.value)}
                                className="phonetic-input"
                            />
                            {phonetic && <span className="auto-filled">✓ 已自动填充</span>}
                        </div>
                    )}

                    {/* 释义/翻译显示区域 */}
                    {meaning && !lookingUp && (
                        <div className="translation-display">
                            <span className="translation-label">
                                <img src="/icon/fanyi.png" alt="" className="label-icon" />
                                释义/翻译
                            </span>
                            <div className="translation-text">{meaning}</div>
                        </div>
                    )}

                    {/* 例句显示（仅单词） */}
                    {type === 'word' && example && (
                        <div className="example-display">
                            <span className="example-label">
                                <img src="/icon/liju.png" alt="" className="label-icon" />
                                例句
                            </span>
                            <div className="example-text">{example}</div>
                            {exampleTranslation && (
                                <div className="example-translation">{exampleTranslation}</div>
                            )}
                        </div>
                    )}

                    {/* 词根词缀分析（仅单词） */}
                    {type === 'word' && morphemeAnalysis && morphemeAnalysis.foundMorphemes.length > 0 && (
                        <div className="morpheme-display">
                            <span className="morpheme-label">
                                <img src="/icon/cigencizhui.png" alt="" className="label-icon" />
                                词根词缀
                            </span>
                            <div className="morpheme-list">
                                {morphemeAnalysis.foundMorphemes.map((item, index) => (
                                    <div key={index} className={`morpheme-item ${item.position}`}>
                                        <span className="morpheme-type">
                                            {item.position === 'prefix' ? '前缀' : item.position === 'root' ? '词根' : '后缀'}
                                        </span>
                                        <span className="morpheme-text">{item.morpheme.morpheme}</span>
                                        <span className="morpheme-meaning">{item.morpheme.meaning}</span>
                                        <span className="morpheme-origin">({item.morpheme.origin})</span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* 固定搭配提示（仅单词且有搭配时显示） */}
                    {type === 'word' && collocations.length > 0 && (
                        <div className="collocation-display">
                            <span className="collocation-label">
                                💡 检测到固定搭配
                            </span>
                            <div className="collocation-hint">点击可切换为添加短语</div>
                            <div className="collocation-list">
                                {collocations.map((col, index) => (
                                    <div
                                        key={index}
                                        className={`collocation-item ${selectedCollocation?.phrase === col.phrase ? 'selected' : ''}`}
                                        onClick={() => {
                                            if (selectedCollocation?.phrase === col.phrase) {
                                                // 取消选择，恢复为单词
                                                setSelectedCollocation(null);
                                                setType('word');
                                            } else {
                                                // 选择该搭配
                                                setSelectedCollocation(col);
                                                setType('phrase');
                                                setMeaning(col.meaning);
                                            }
                                        }}
                                    >
                                        <span className="collocation-phrase">{col.phrase}</span>
                                        <span className="collocation-meaning">{col.meaning}</span>
                                        {selectedCollocation?.phrase === col.phrase && (
                                            <span className="collocation-selected-badge">✓ 已选择</span>
                                        )}
                                    </div>
                                ))}
                            </div>
                            {selectedCollocation && (
                                <div className="collocation-selected-info">
                                    将添加短语 "<strong>{selectedCollocation.phrase}</strong>" 而非单词 "{selectedText}"
                                </div>
                            )}
                        </div>
                    )}

                    {/* 释义输入（可编辑） */}
                    <div className="input-group meaning-group">
                        <input
                            type="text"
                            placeholder={lookingUp ? "正在查询翻译..." : "释义/翻译（可编辑）"}
                            value={meaning}
                            onChange={(e) => setMeaning(e.target.value)}
                            className="meaning-input"
                        />
                        {meaning && <span className="auto-filled">✓ 自动翻译</span>}
                    </div>

                    {/* 词性和更多释义 */}
                    {dictResult && dictResult.meanings && dictResult.meanings.length > 1 && (
                        <div className="more-meanings">
                            <span className="more-meanings-label">更多释义：</span>
                            {dictResult.meanings.slice(0, 3).map((m, i) => (
                                <span key={i} className="meaning-tag" onClick={() => {
                                    if (m.definitions && m.definitions[0]) {
                                        const chinesePos = toChinesePartOfSpeech(m.partOfSpeech);
                                        setMeaning(`[${chinesePos}] ${m.definitions[0].definition}`);
                                    }
                                }}>
                                    {toChinesePartOfSpeech(m.partOfSpeech)}
                                </span>
                            ))}
                        </div>
                    )}

                    {/* 错误提示 */}
                    {error && <div className="error-message">{error}</div>}
                </div>

                <div className="popup-actions">
                    {alreadyExists ? (
                        <button
                            className="add-btn already-exists"
                            onClick={onClose}
                        >
                            该词已在生词本
                        </button>
                    ) : (
                        <button
                            className="add-btn primary-action"
                            onClick={handleQuickAdd}
                            disabled={loading}
                        >
                            {loading ? (
                                <span className="loading-spinner"></span>
                            ) : (
                                <>
                                    <img src="/icon/shengciben.png" alt="" className="btn-icon" />
                                    加入生词本
                                </>
                            )}
                        </button>
                    )}
                    <div className="shortcut-hints">
                        <span className="hint"><kbd>Enter</kbd> 保存</span>
                        <span className="hint"><kbd>Esc</kbd> 关闭</span>
                    </div>
                </div>
            </div>
        );
    };

    return createPortal(
        <div className="add-vocabulary-overlay">
            {renderContent()}
        </div>,
        document.body
    );
};

export default AddVocabularyPopup;
