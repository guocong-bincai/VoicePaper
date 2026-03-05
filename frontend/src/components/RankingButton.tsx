import { useEffect, useState } from 'react';
import { createPortal } from 'react-dom';
import { getMyRanking } from '../services/api';
import type { UserRanking } from '../types';
import './RankingButton.css';

interface RankingButtonProps {
  onShowModal?: () => void;
}

export function RankingButton({ onShowModal }: RankingButtonProps) {
  const [ranking, setRanking] = useState<UserRanking | null>(null);
  const [showModal, setShowModal] = useState(false);
  const [loading, setLoading] = useState(true);

  // 组件挂载时自动加载排名数据
  useEffect(() => {
    loadRanking();
  }, []);

  const loadRanking = async () => {
    try {
      setLoading(true);
      const data = await getMyRanking();
      setRanking(data);
    } catch (err) {
      console.error('获取排名失败:', err);
    } finally {
      setLoading(false);
    }
  };

  // BUG修复: 点击打开弹窗时刷新排名数据，确保显示最新的真实排名
  // 修复策略: 在打开弹窗时重新加载排名数据，避免显示过期的缓存数据
  // 影响范围: frontend/src/components/RankingButton.tsx:32-35
  // 修复日期: 2025-01-27
  const handleClick = async () => {
    // 打开弹窗前刷新排名数据，确保显示最新排名
    await loadRanking();
    setShowModal(true);
    onShowModal?.();
  };

  const handleClose = () => {
    setShowModal(false);
  };

  // 如果加载中，显示加载状态
  if (loading) {
    return (
      <button className="ranking-button ranking-loading" disabled>
        <img src="/icon/paihangbang.png" alt="排名" className="ranking-btn-icon" />
        <span>排名</span>
      </button>
    );
  }

  return (
    <>
      {/* 排名按钮 - 默认显示自己的排名 */}
      <button
        className="ranking-button"
        onClick={handleClick}
        title={ranking ? `全站排名 #${ranking.rank}` : '查看排名'}
      >
        <img src="/icon/paihangbang.png" alt="排名" className="ranking-btn-icon" />
        <span className="ranking-label">全站排名</span>
        {ranking ? (
          <span className="ranking-number">{ranking.rank}</span>
        ) : (
          <span>加载中</span>
        )}
      </button>

      {/* 排名弹窗 - 主要显示前5名列表 */}
      {showModal && createPortal(
        <div className="ranking-modal-overlay" onClick={handleClose}>
          <div className="ranking-modal-content" onClick={(e) => e.stopPropagation()}>
            <button className="modal-close-btn" onClick={handleClose}>✕</button>

            {ranking ? (
              <div className="ranking-modal-body">
                {/* 标题 */}
                <div className="modal-title-section">
                  <h2 className="modal-title">🏆 全站排行榜 · 你是第 {ranking.rank} 名</h2>
                </div>

                {/* 前5名排行 - 重点展示 */}
                <div className="top-rankings-section">
                  <div className="top-rankings-list">
                    {ranking.nearby_users.slice(0, 5).map((user) => {
                      let rankIcon = null;
                      if (user.rank >= 1 && user.rank <= 3) {
                        rankIcon = `/icon/qiandao/no${user.rank}.png`;
                      } else if (user.rank >= 4 && user.rank <= 5) {
                        rankIcon = `/icon/paihang/no${user.rank}.png`;
                      }

                      return (
                        <div key={user.rank} className={`top-ranking-item ${user.isCurrentUser ? 'current-user' : ''}`}>
                          <div className="ranking-position">
                            {rankIcon ? (
                              <img src={rankIcon} alt={`第${user.rank}名`} className="rank-medal-large" />
                            ) : (
                              <span className="rank-number-large">#{user.rank}</span>
                            )}
                          </div>
                          <div className="user-avatar-container">
                            <img
                              src={(user as any).avatar || `https://ui-avatars.com/api/?name=${encodeURIComponent(user.nickname || 'U')}&background=2563EB&color=fff&size=40&bold=true`}
                              alt={user.nickname}
                              className="user-ranking-avatar"
                            />
                          </div>
                          <div className="user-info">
                            <span className="user-name">{user.nickname}</span>
                            <span className="user-score-small">{user.score.toLocaleString()}</span>
                          </div>
                          <div className="user-score-large">{user.score.toLocaleString()}</div>
                        </div>
                      );
                    })}
                  </div>
                </div>

                {/* 你的排名信息（如果不在前5名） */}
                {ranking.rank > 5 && (
                  <div className="your-ranking-section">
                    <div className="divider">其他排名</div>
                    <div className="your-ranking-item">
                      <span className="your-rank">#{ranking.rank}</span>
                      <span className="your-name">你</span>
                      <span className="your-score">{ranking.score.toLocaleString()}</span>
                    </div>
                  </div>
                )}

                {/* 你的统计数据 */}
                <div className="your-stats">
                  <div className="stat-item">
                    <span className="stat-label">累积时长</span>
                    <span className="stat-value">
                      {ranking.total_duration}
                      <span className="stat-unit">分钟</span>
                    </span>
                  </div>
                  <div className="stat-item">
                    <span className="stat-label">累计积分</span>
                    <span className="stat-value">{ranking.total_points}</span>
                  </div>
                </div>
              </div>
            ) : (
              <div className="modal-error">无法加载排名数据</div>
            )}
          </div>
        </div>,
        document.body
      )}
    </>
  );
}