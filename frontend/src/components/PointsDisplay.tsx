import { useEffect, useState } from 'react';
import { getMyPoints } from '../services/api';
import type { UserPoints } from '../types';
import { RankingDisplay } from './RankingDisplay';
import './PointsDisplay.css';

export function PointsDisplay() {
  const [points, setPoints] = useState<UserPoints | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadPoints();
  }, []);

  const loadPoints = async () => {
    try {
      const data = await getMyPoints();
      setPoints(data);
    } catch (error) {
      console.error('获取积分信息失败:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="points-display loading">加载中...</div>;
  }

  if (!points) {
    return null;
  }

  // 计算等级进度
  const levelThresholds = [0, 100, 500, 1000, 3000, 5000, 10000, 20000, 30000, 50000];
  const currentLevelMin = levelThresholds[points.level - 1] || 0;
  const nextLevelMin = levelThresholds[points.level] || currentLevelMin;
  const progressInLevel = points.total_points - currentLevelMin;
  const levelRange = nextLevelMin - currentLevelMin;
  const levelProgress = levelRange > 0 ? (progressInLevel / levelRange) * 100 : 100;

  return (
    <div className="points-display">
      <div className="points-card">
        <div className="points-header">
          <div className="level-badge">
            <span className="level-number">Lv.{points.level}</span>
            <span className="level-name">
              {(points as any).display_icon && <span className="display-icon">{(points as any).display_icon}</span>}
              {(points as any).display_name || points.level_name}
            </span>
          </div>
          <div className="points-main">
            <div className="points-value">
              <span className="points-number">{points.current_points.toLocaleString()}</span>
              <span className="points-label">积分</span>
            </div>
          </div>
        </div>

        <div className="level-progress-bar">
          <div className="progress-fill" style={{ width: `${levelProgress}%` }}></div>
        </div>
        <div className="level-progress-text">
          {points.level < 10 ? (
            <span>距离下一级还需 {(nextLevelMin - points.total_points).toLocaleString()} 积分</span>
          ) : (
            <span>已达到最高等级！</span>
          )}
        </div>

        <div className="points-stats">
          <div className="stat-item">
            <img src="/icon/yuedu.png" alt="阅读" className="stat-icon" style={{ width: '28px', height: '28px' }} onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; (e.target as HTMLImageElement).insertAdjacentHTML('afterend', '<span style="font-size:24px">📚</span>'); }} />
            <span className="stat-value">{points.total_articles_read}</span>
            <span className="stat-label">阅读</span>
          </div>
          <div className="stat-item">
            <img src="/icon/moxie.png" alt="默写" className="stat-icon" style={{ width: '28px', height: '28px' }} onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; (e.target as HTMLImageElement).insertAdjacentHTML('afterend', '<span style="font-size:24px">✍️</span>'); }} />
            <span className="stat-value">{points.total_dictations_completed}</span>
            <span className="stat-label">默写</span>
          </div>
          <div className="stat-item">
            <img src="/icon/lianxuqiandao.png" alt="连续签到" className="stat-icon" style={{ width: '28px', height: '28px' }} onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; (e.target as HTMLImageElement).insertAdjacentHTML('afterend', '<span style="font-size:24px">📅</span>'); }} />
            <span className="stat-value">{points.continuous_check_ins}</span>
            <span className="stat-label">连续签到</span>
          </div>
        </div>
      </div>

      {/* 全站排名 */}
      <div style={{ marginTop: '24px' }}>
        <RankingDisplay />
      </div>
    </div>
  );
}
