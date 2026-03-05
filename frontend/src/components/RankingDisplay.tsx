import { useEffect, useState } from 'react';
import { getMyRanking } from '../services/api';
import type { UserRanking } from '../types';
import './RankingDisplay.css';

interface RankingDisplayProps {
  userId?: number;
}

export function RankingDisplay({ userId }: RankingDisplayProps) {
  const [ranking, setRanking] = useState<UserRanking | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadRanking();
  }, [userId]);

  const loadRanking = async () => {
    try {
      setLoading(true);
      const data = await getMyRanking();
      console.log('✅ 排名数据:', data);
      setRanking(data);
      setError('');
    } catch (err: any) {
      const errorMsg = err.response?.data?.details || err.message || '获取排名信息失败';
      console.error('❌ 获取排名信息失败:', errorMsg, err);
      setError(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="ranking-display loading">加载排名中...</div>;
  }

  if (error) {
    return (
      <div className="ranking-display error" style={{ padding: '20px', textAlign: 'center', color: '#dc2626' }}>
        <p>❌ {error}</p>
        <button onClick={loadRanking} style={{ marginTop: '10px', padding: '8px 16px', cursor: 'pointer' }}>
          重试
        </button>
      </div>
    );
  }

  if (!ranking) {
    return <div className="ranking-display error">无法加载排名数据</div>;
  }

  // 计算排名进度百分比
  const percentile = Math.round((ranking.rank / ranking.total_users) * 100);
  const topPercentage = 100 - percentile;

  return (
    <div className="ranking-display">
      <div className="ranking-card">
        {/* 排名头部 */}
        <div className="ranking-header">
          <div className="ranking-badge">
            <img src="/icon/paihangbang.png" alt="排名" className="ranking-icon" />
            <span className="ranking-number">#{ranking.rank}</span>
          </div>
          <div className="ranking-main">
            <div className="ranking-info">
              <span className="ranking-label">全站排名</span>
              <span className="ranking-percentage">超越了 {topPercentage}% 的用户</span>
            </div>
          </div>
        </div>

        {/* 排名指标 */}
        <div className="ranking-metrics">
          <div className="metric-item">
            <div className="metric-label">累积时长</div>
            <div className="metric-value">
              <span className="metric-number">{ranking.total_duration}</span>
              <span className="metric-unit">分钟</span>
            </div>
          </div>
          <div className="metric-divider"></div>
          <div className="metric-item">
            <div className="metric-label">累计积分</div>
            <div className="metric-value">
              <span className="metric-number">{ranking.total_points.toLocaleString()}</span>
              <span className="metric-unit">分</span>
            </div>
          </div>
          <div className="metric-divider"></div>
          <div className="metric-item">
            <div className="metric-label">综合分数</div>
            <div className="metric-value">
              <span className="metric-number">{ranking.score.toLocaleString()}</span>
            </div>
          </div>
        </div>

        {/* 排名公式说明 */}
        <div className="ranking-formula">
          <div className="formula-title">排名计算公式</div>
          <div className="formula-content">
            综合分数 = 累积时长 × 0.5 + 累计积分 × 0.5
          </div>
        </div>

        {/* 周围排名用户 */}
        <div className="nearby-rankings">
          <div className="nearby-title">周围排名</div>
          <div className="nearby-list">
            {ranking.nearby_users.map((user) => {
              // 为前5名显示特殊的排名图标
              const getRankIcon = (rank: number) => {
                if (rank >= 1 && rank <= 3) {
                  return `/icon/qiandao/no${rank}.png`;
                }
                if (rank >= 4 && rank <= 5) {
                  return `/icon/paihang/no${rank}.png`;
                }
                return null;
              };

              const rankIcon = getRankIcon(user.rank);

              return (
                <div
                  key={user.rank}
                  className={`nearby-item ${user.isCurrentUser ? 'current-user' : ''}`}
                >
                  {rankIcon ? (
                    <img src={rankIcon} alt={`第${user.rank}名`} className="rank-medal" />
                  ) : (
                    <div className="nearby-rank">#{user.rank}</div>
                  )}
                  <div className="user-avatar-mini">
                    <img
                      src={(user as any).avatar || `https://ui-avatars.com/api/?name=${encodeURIComponent(user.nickname || 'U')}&background=2563EB&color=fff&size=32&bold=true`}
                      alt={user.nickname}
                    />
                  </div>
                  <div className="nearby-name">{user.nickname}</div>
                  <div className="nearby-score">{user.score.toLocaleString()}</div>
                </div>
              );
            })}
          </div>
        </div>

        {/* 刷新按钮 */}
        <button className="ranking-refresh-btn" onClick={loadRanking}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M23 4v6h-6"/>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
          </svg>
          刷新排名
        </button>
      </div>
    </div>
  );
}
