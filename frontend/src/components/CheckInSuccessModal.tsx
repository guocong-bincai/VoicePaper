import { useEffect, useState } from 'react';
import { createPortal } from 'react-dom';
import confetti from 'canvas-confetti';
import { getCheckInRanking } from '../services/api';
import './CheckInSuccessModal.css';

interface CheckInSuccessModalProps {
  checkInResult: any;
  onClose: () => void;
}

interface RankingItem {
  rank: number;
  user_id: number;
  nickname: string;
  avatar: string;
  total_check_ins: number;
  max_continuous_check_ins: number;
  current_continuous_days: number;
  is_me: boolean;
}

export function CheckInSuccessModal({ checkInResult, onClose }: CheckInSuccessModalProps) {
  const [ranking, setRanking] = useState<RankingItem[]>([]);
  const [userRank, setUserRank] = useState<number>(0);
  const [loading, setLoading] = useState(true);
  const [sortBy, setSortBy] = useState<'total' | 'continuous'>('total');

  useEffect(() => {
    console.log('🎊 CheckInSuccessModal 组件已加载，checkInResult:', checkInResult);
    // 播放ding声音
    playDingSound();
    // 触发彩色亮片特效
    triggerConfetti();

    // 加载排行榜
    loadRanking();
  }, []);

  const playDingSound = () => {
    try {
      const audio = new Audio('/sounds/right.mp3');
      audio.play();
    } catch (error) {
      console.error('播放声音失败:', error);
    }
  };

  const triggerConfetti = () => {
    try {
      const count = 200;
      const defaults = {
        origin: { y: 0.5 },
        zIndex: 999999
      };

      function fire(particleRatio: number, opts: any) {
        confetti({
          ...defaults,
          ...opts,
          particleCount: Math.floor(count * particleRatio),
          zIndex: 999999
        });
      }

      // 第一波：超大爆炸
      fire(0.25, {
        spread: 26,
        startVelocity: 55,
        colors: ['#6366f1', '#22c55e', '#f59e0b', '#ec4899', '#8b5cf6', '#3b82f6']
      });
      fire(0.2, {
        spread: 60,
        colors: ['#6366f1', '#22c55e', '#f59e0b', '#ec4899', '#8b5cf6', '#3b82f6']
      });
      fire(0.35, {
        spread: 100,
        decay: 0.91,
        scalar: 0.8,
        colors: ['#6366f1', '#22c55e', '#f59e0b', '#ec4899', '#8b5cf6', '#3b82f6']
      });
      fire(0.1, {
        spread: 120,
        startVelocity: 25,
        decay: 0.92,
        scalar: 1.2,
        colors: ['#6366f1', '#22c55e', '#f59e0b', '#ec4899', '#8b5cf6', '#3b82f6']
      });
      fire(0.1, {
        spread: 120,
        startVelocity: 45,
        colors: ['#6366f1', '#22c55e', '#f59e0b', '#ec4899', '#8b5cf6', '#3b82f6']
      });

      // 持续喷射
      let duration = 2000;
      let animationEnd = Date.now() + duration;

      (function frame() {
        confetti({
          particleCount: 5,
          angle: 60,
          spread: 55,
          origin: { x: 0, y: 0.8 },
          zIndex: 999999,
          colors: ['#6366f1', '#22c55e', '#f59e0b', '#ec4899', '#8b5cf6', '#3b82f6']
        });
        confetti({
          particleCount: 5,
          angle: 120,
          spread: 55,
          origin: { x: 1, y: 0.8 },
          zIndex: 999999,
          colors: ['#6366f1', '#22c55e', '#f59e0b', '#ec4899', '#8b5cf6', '#3b82f6']
        });

        if (Date.now() < animationEnd) {
          requestAnimationFrame(frame);
        }
      }());
    } catch (error) {
      console.error('庆祝动画执行出错:', error);
    }
  };

  const loadRanking = async () => {
    try {
      setLoading(true);
      const data = await getCheckInRanking(20, sortBy);
      setRanking(data.items || []);
      setUserRank(data.user_rank || 0);
    } catch (error) {
      console.error('获取排行榜失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRanking();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sortBy]);

  const getRankIcon = (rank: number) => {
    if (rank === 1) return <img src="/icon/qiandao/no1.png" alt="1" className="rank-medal-img" />;
    if (rank === 2) return <img src="/icon/qiandao/no2.png" alt="2" className="rank-medal-img" />;
    if (rank === 3) return <img src="/icon/qiandao/no3.png" alt="3" className="rank-medal-img" />;
    return `${rank}`;
  };

  return createPortal(
    <div className="check-in-success-modal-overlay" onClick={onClose}>
      <div className="check-in-success-modal" onClick={(e) => e.stopPropagation()}>
        {/* 右上角关闭按钮 */}
        <button
          className="modal-close-x"
          onClick={onClose}
          title="关闭"
          aria-label="关闭"
        >
          ✕
        </button>

        {/* 可滚动内容区域 */}
        <div className="modal-content-scrollable">
          {/* 成功图标和标题 */}
          <div className="modal-header">
          <div className="success-icon-large">
            <div className="success-icon-ring-large"></div>
            <div className="success-icon-ring-large ring-2"></div>
            <span className="success-check">✓</span>
          </div>
          <h2 className="modal-title">签到成功！</h2>
          <div className="points-display-large">
            <span className="points-plus-large">+</span>
            <span className="points-number-large">{checkInResult.total_points}</span>
            <span className="points-unit-large">积分</span>
          </div>
        </div>

        {/* 用户统计卡片 */}
        <div className="user-stats-section">
          <div className="stat-card-large">
            <div className="stat-icon-large">
              <img src="/icon/qiandao/lianxuqiandao.png" alt="连续签到" className="stat-icon-img-large" />
            </div>
            <div className="stat-value-large">{checkInResult.continuous_days}</div>
            <div className="stat-label-large">连续签到</div>
          </div>
          <div className="stat-card-large">
            <div className="stat-icon-large">
              <img src="/icon/qiandao/leijiqiandao.png" alt="累计签到" className="stat-icon-img-large" />
            </div>
            <div className="stat-value-large">{checkInResult.total_check_ins || 0}</div>
            <div className="stat-label-large">累计签到</div>
          </div>
          <div className="stat-card-large stat-highlight">
            <div className="stat-icon-large">
              <img src="/icon/qiandao/good.png" alt="最长连续" className="stat-icon-img-large" />
            </div>
            <div className="stat-value-large">{checkInResult.max_continuous_days || 0}</div>
            <div className="stat-label-large">最长连续</div>
          </div>
        </div>

        {/* 排行榜 */}
        <div className="ranking-section">
          <div className="ranking-header">
            <h3 className="ranking-title">
              签到排行榜
              {userRank > 0 && (
                <span className="ranking-title-rank">· 你是第 {userRank} 名</span>
              )}
            </h3>
            <div className="ranking-tabs">
              <button
                className={`tab-btn ${sortBy === 'total' ? 'active' : ''}`}
                onClick={() => setSortBy('total')}
              >
                总签到
              </button>
              <button
                className={`tab-btn ${sortBy === 'continuous' ? 'active' : ''}`}
                onClick={() => setSortBy('continuous')}
              >
                连续签到
              </button>
            </div>
          </div>

          {loading ? (
            <div className="ranking-loading">加载中...</div>
          ) : (
            <div className="ranking-list">
              {ranking.map((item) => (
                <div
                  key={item.user_id}
                  className={`ranking-item ${item.is_me ? 'is-me' : ''}`}
                >
                  <div className="ranking-position">
                    {item.rank <= 3 ? (
                      <span className="rank-medal">{getRankIcon(item.rank)}</span>
                    ) : (
                      <span className="rank-number-large">{item.rank}</span>
                    )}
                  </div>
                  <div className="user-avatar-container">
                    <img
                      src={item.avatar || `https://ui-avatars.com/api/?name=${encodeURIComponent(item.nickname || 'U')}&background=2563EB&color=fff&size=40&bold=true`}
                      alt={item.nickname}
                      className="user-ranking-avatar"
                    />
                  </div>
                  <div className="user-info">
                    <span className="user-name">
                      {item.nickname || `用户${item.user_id}`}
                      {item.is_me && <span className="me-badge">我</span>}
                    </span>
                    <span className="user-score-small">
                      {sortBy === 'total' ? (
                        `总签到: ${item.total_check_ins}天 • 最长连续: ${item.max_continuous_check_ins}天`
                      ) : (
                        `最长连续: ${item.max_continuous_check_ins}天 • 总签到: ${item.total_check_ins}天`
                      )}
                    </span>
                  </div>
                  <div className="user-score-large">
                    {sortBy === 'total' ? item.total_check_ins : item.max_continuous_check_ins}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
        </div>
      </div>
    </div>,
    document.body
  );
}
