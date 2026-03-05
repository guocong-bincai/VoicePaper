import { useEffect, useState } from 'react';
import { checkIn, getCheckInStatus } from '../services/api';
import type { CheckInStatus } from '../types';
import { CheckInSuccessModal } from './CheckInSuccessModal';
import './CheckInButton.css';

export function CheckInButton() {
  const [status, setStatus] = useState<CheckInStatus | null>(null);
  const [loading, setLoading] = useState(false);
  const [showResult, setShowResult] = useState(false);
  const [checkInResult, setCheckInResult] = useState<any>(null);

  useEffect(() => {
    loadStatus();
  }, []);

  const loadStatus = async () => {
    try {
      const data = await getCheckInStatus();
      setStatus(data);
    } catch (error) {
      console.error('获取签到状态失败:', error);
    }
  };

  const handleCheckIn = async () => {
    if (loading || status?.today_checked_in) return;

    setLoading(true);
    try {
      const result = await checkIn();
      setCheckInResult(result);
      setShowResult(true);
      
      // 刷新状态
      await loadStatus();

      // 3秒后自动关闭结果显示
      setTimeout(() => {
        setShowResult(false);
      }, 3000);
    } catch (error: any) {
      alert(error.response?.data?.error || '签到失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  if (!status) {
    return null;
  }

  return (
    <div className="check-in-container">
      <button
        className={`check-in-button ${status.today_checked_in ? 'checked' : ''}`}
        onClick={handleCheckIn}
        disabled={status.today_checked_in || loading}
      >
        {status.today_checked_in ? (
          <>
            <span className="check-icon">✓</span>
            <span>今日已签到</span>
          </>
        ) : (
          <>
            <span className="calendar-icon">📅</span>
            <span>{loading ? '签到中...' : '每日签到'}</span>
          </>
        )}
      </button>

      <div className="check-in-info">
        <div className="streak-display">
          <div className="fire-icon-wrapper">
            <span className="fire-icon">🔥</span>
            <span className="streak-count">{status.continuous_days}</span>
          </div>
          <div className="streak-text">
            <div className="streak-label">连续签到</div>
            <div className="streak-days">{status.continuous_days} 天</div>
          </div>
        </div>
        {status.next_milestone > 0 && (
          <div className="milestone-progress">
            <div className="milestone-text">
              距离 {status.next_milestone} 天里程碑还需 {status.next_milestone - status.continuous_days} 天
            </div>
            <div className="milestone-reward">
              🎁 奖励 {status.next_milestone_reward} 积分
            </div>
            <div className="milestone-bar">
              <div 
                className="milestone-fill" 
                style={{ width: `${(status.continuous_days / status.next_milestone) * 100}%` }}
              ></div>
            </div>
          </div>
        )}
      </div>

      {showResult && checkInResult && (
        <CheckInSuccessModal
          checkInResult={checkInResult}
          onClose={() => setShowResult(false)}
        />
      )}
    </div>
  );
}
