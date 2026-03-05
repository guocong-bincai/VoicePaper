import { useEffect, useState } from 'react';
import { getTitleProgress, equipTitle, unequipTitle } from '../services/api';
import type { TitleProgress } from '../types';
import { AchievementIcon } from './AchievementIcon';
import './TitleBadges.css';

interface TitleBadgesProps {
  onTitleChange?: () => void;  // 称号变化时的回调
}

export function TitleBadges({ onTitleChange }: TitleBadgesProps) {
  const [titles, setTitles] = useState<TitleProgress[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');

  useEffect(() => {
    loadTitles();
  }, []);

  const loadTitles = async () => {
    try {
      const data = await getTitleProgress();
      setTitles(data);
    } catch (error) {
      console.error('获取称号失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleEquipTitle = async (titleId: number) => {
    try {
      await equipTitle(titleId);
      await loadTitles();
      onTitleChange?.(); // 通知父组件称号已更换
      alert('称号佩戴成功！');
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      alert(err.response?.data?.error || '佩戴失败');
    }
  };

  // @ts-expect-error - 保留以供将来使用
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const _handleUnequipTitle = async () => {
    try {
      await unequipTitle();
      await loadTitles();
      alert('已取消佩戴');
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      alert(err.response?.data?.error || '取消失败');
    }
  };

  const categories = [
    { key: 'all', label: '全部', icon: '' },
    { key: 'reading', label: '阅读', icon: '/icon/chenghao/yuedu.png' },
    { key: 'dictation', label: '默写', icon: '/icon/chenghao/moxie.png' },
    { key: 'check_in', label: '签到', icon: '/icon/chenghao/qiandao.png' },
    { key: 'points', label: '积分', icon: '/icon/chenghao/jifen.png' },
    { key: 'special', label: '特殊', icon: '/icon/chenghao/teshu.png' },
  ];

  const filteredTitles = selectedCategory === 'all'
    ? titles
    : titles.filter(t => t.category === selectedCategory);

  // 普通称号的颜色变体
  const commonColorVariants = ['teal', 'coral', 'mint', 'brown', 'olive', 'cyan'];
  
  // 特定称号名称的颜色映射（覆盖默认的 titleId 计算）
  const titleNameColorMap: Record<string, string> = {
    '初心': 'coral',  // 粉色
    '笔吏': 'cyan',   // 青绿色
  };
  
  const getColorVariantClass = (titleId: number, rarity: string, titleName: string) => {
    if (rarity !== 'common') return '';
    // 优先检查特定称号名称的颜色映射
    if (titleNameColorMap[titleName]) {
      return `common-${titleNameColorMap[titleName]}`;
    }
    const variant = commonColorVariants[titleId % commonColorVariants.length];
    return `common-${variant}`;
  };

  const getRarityColor = (rarity: string) => {
    switch (rarity) {
      case 'legendary': return '#ff6b00';
      case 'epic': return '#a335ee';
      case 'rare': return '#0070dd';
      default: return '#9d9d9d';
    }
  };

  const getRarityLabel = (rarity: string) => {
    switch (rarity) {
      case 'legendary': return '传说';
      case 'epic': return '史诗';
      case 'rare': return '稀有';
      default: return '普通';
    }
  };

  if (loading) {
    return <div className="title-badges loading">加载中...</div>;
  }

  return (
    <div className="title-badges">
      <div className="title-header">
        <h2><img src="/icon/chenghao/chenghao.png" alt="" className="title-header-icon" /> 称号收集</h2>
        <p className="subtitle">
          已获得 {titles.filter(t => t.owned).length} / {titles.length} 个称号
        </p>
      </div>

      <div className="category-tabs">
        {categories.map(cat => (
          <button
            key={cat.key}
            className={`category-tab ${selectedCategory === cat.key ? 'active' : ''}`}
            onClick={() => setSelectedCategory(cat.key)}
          >
            {cat.icon && <img src={cat.icon} alt="" className="category-icon" />}
            {cat.label}
          </button>
        ))}
      </div>

      <div className="titles-grid">
        {filteredTitles.map(title => (
          <div
            key={title.title_id}
            className={`title-card ${title.owned ? 'owned' : 'locked'} ${title.is_equipped ? 'equipped' : ''} rarity-${title.rarity} ${getColorVariantClass(title.title_id, title.rarity, title.title_name)}`}
            style={{ borderColor: getRarityColor(title.rarity) }}
          >
            <div className="title-icon">
              <AchievementIcon 
                rarity={title.rarity as 'common' | 'rare' | 'epic' | 'legendary'} 
                size="medium"
                locked={!title.owned}
                colorVariant={title.title_id}
                titleName={title.title_name}
              />
            </div>
            <div className="title-info">
              <div className="title-name">
                {title.title_name}
                <span
                  className="rarity-badge"
                  style={{ backgroundColor: getRarityColor(title.rarity) }}
                >
                  {getRarityLabel(title.rarity)}
                </span>
              </div>
              <div className="title-description">{title.description}</div>
              <div className="title-condition">{title.condition_description}</div>
              
              {!title.owned && (
                <div className="title-progress">
                  <div className="progress-bar">
                    <div
                      className="progress-fill"
                      style={{ width: `${Math.min(title.progress, 100)}%` }}
                    ></div>
                  </div>
                  <div className="progress-text">
                    {title.current_value} / {title.condition_value} ({Math.floor(title.progress)}%)
                  </div>
                </div>
              )}

              {title.owned && (
                title.is_equipped ? (
                  <div className="equipped-badge">
                    <span className="equipped-icon">✓</span>
                    <span>正在穿戴</span>
                  </div>
                ) : (
                  <button
                    className="equip-button"
                    onClick={() => handleEquipTitle(title.title_id)}
                  >
                    穿戴称号
                  </button>
                )
              )}
            </div>
          </div>
        ))}
      </div>

      {filteredTitles.length === 0 && (
        <div className="empty-state">
          <p>该分类下暂无称号</p>
        </div>
      )}
    </div>
  );
}
