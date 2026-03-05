import React, { useState, useEffect, useRef } from 'react';
import { getMyPoints, getCheckInStatus, checkIn, getCheckInCalendar, getMakeupCardInfo, buyMakeupCard, useMakeupCard, type MakeupCardInfo } from '../services/api';
import { CheckInSuccessModal } from './CheckInSuccessModal';
import './QuickPointsCard.css';

interface QuickPointsCardProps {
    onOpenProfile?: () => void;
}

export const QuickPointsCard: React.FC<QuickPointsCardProps> = ({ onOpenProfile }) => {
    const [points, setPoints] = useState<number>(0);
    const [checkedIn, setCheckedIn] = useState<boolean>(false);
    const [continuousDays, setContinuousDays] = useState<number>(0);
    const [loading, setLoading] = useState<boolean>(false);
    const [showCheckInModal, setShowCheckInModal] = useState<boolean>(false);
    const [checkInResult, setCheckInResult] = useState<any>(null);
    const [showCalendar, setShowCalendar] = useState<boolean>(false);
    const [calendarDates, setCalendarDates] = useState<string[]>([]);
    const [calendarLoading, setCalendarLoading] = useState<boolean>(false);
    const calendarRef = useRef<HTMLDivElement>(null);
    const hoverTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    // 补签卡相关状态
    const [makeupCardInfo, setMakeupCardInfo] = useState<MakeupCardInfo | null>(null);
    const [showBuyModal, setShowBuyModal] = useState<boolean>(false);
    const [makingUp, setMakingUp] = useState<boolean>(false);
    const [buyLoading, setBuyLoading] = useState<boolean>(false);

    // 获取积分和签到状态
    useEffect(() => {
        fetchData();
    }, []);

    const fetchData = async () => {
        try {
            const [pointsData, checkInData] = await Promise.all([
                getMyPoints(),
                getCheckInStatus()
            ]);

            setPoints(pointsData.current_points || 0);
            setCheckedIn(checkInData.today_checked_in || false);
            setContinuousDays(checkInData.continuous_days || 0);
        } catch (error) {
            console.error('获取积分或签到状态失败:', error);
        }
    };

    // 获取签到日历
    const fetchCalendar = async () => {
        if (calendarLoading || calendarDates.length > 0) return;

        setCalendarLoading(true);
        try {
            const now = new Date();
            const [calendarData, makeupData] = await Promise.all([
                getCheckInCalendar(now.getFullYear(), now.getMonth() + 1),
                getMakeupCardInfo()
            ]);
            setCalendarDates(calendarData.check_in_dates || []);
            setMakeupCardInfo(makeupData);
        } catch (error) {
            console.error('获取签到日历失败:', error);
        } finally {
            setCalendarLoading(false);
        }
    };

    const handleMouseEnter = () => {
        if (hoverTimeoutRef.current) {
            clearTimeout(hoverTimeoutRef.current);
        }
        hoverTimeoutRef.current = setTimeout(() => {
            setShowCalendar(true);
            fetchCalendar();
        }, 300);
    };

    const handleMouseLeave = (e: React.MouseEvent) => {
        if (hoverTimeoutRef.current) {
            clearTimeout(hoverTimeoutRef.current);
        }
        // 检查鼠标是否移动到日历弹窗内
        const relatedTarget = e.relatedTarget as HTMLElement;
        if (relatedTarget && relatedTarget instanceof Node && calendarRef.current?.contains(relatedTarget)) {
            return;
        }
        // 延迟关闭，给用户移动鼠标到日历的时间
        hoverTimeoutRef.current = setTimeout(() => {
            setShowCalendar(false);
        }, 200);
    };

    const handleCalendarMouseEnter = () => {
        if (hoverTimeoutRef.current) {
            clearTimeout(hoverTimeoutRef.current);
        }
    };

    const handleCalendarMouseLeave = () => {
        hoverTimeoutRef.current = setTimeout(() => {
            setShowCalendar(false);
        }, 200);
    };

    const handleCheckIn = async (e: React.MouseEvent) => {
        e.stopPropagation();

        if (checkedIn || loading) return;

        setLoading(true);
        try {
            const result = await checkIn();
            console.log('✅ 签到成功，结果:', result);
            setCheckInResult(result);
            setCheckedIn(true);
            setContinuousDays(result.continuous_days);
            setPoints(prev => prev + result.total_points);
            console.log('🎉 准备显示弹框');
            setShowCheckInModal(true);

            // 更新日历（添加今天）
            const today = new Date().toISOString().split('T')[0];
            setCalendarDates(prev => [...prev, today]);
        } catch (error: any) {
            console.error('签到失败:', error);
            alert(error.response?.data?.error || '签到失败');
        } finally {
            setLoading(false);
        }
    };

    // 补签处理
    const handleMakeup = async (dateStr: string, e: React.MouseEvent) => {
        e.stopPropagation();

        if (makingUp) return;
        if (!makeupCardInfo || makeupCardInfo.makeup_cards <= 0) {
            setShowBuyModal(true);
            return;
        }

        if (!confirm(`确定要使用补签卡补签 ${dateStr} 吗？`)) return;

        setMakingUp(true);
        try {
            const result = await useMakeupCard(dateStr);
            alert(result.message);
            // 更新日历
            setCalendarDates(prev => [...prev, dateStr]);
            // 更新补签卡数量
            setMakeupCardInfo(prev => prev ? { ...prev, makeup_cards: result.makeup_cards, makeup_dates: prev.makeup_dates.filter(d => d !== dateStr) } : null);
            // 更新连续签到天数
            if (result.continuous_days !== undefined) {
                setContinuousDays(result.continuous_days);
            }
        } catch (error: any) {
            console.error('补签失败:', error);
            alert(error.response?.data?.error || '补签失败');
        } finally {
            setMakingUp(false);
        }
    };

    // 购买补签卡
    const handleBuyCard = async (e: React.MouseEvent) => {
        e.stopPropagation();

        if (buyLoading) return;
        if (!makeupCardInfo?.can_buy) {
            alert('积分不足，无法购买');
            return;
        }

        setBuyLoading(true);
        try {
            const result = await buyMakeupCard(1);
            alert(result.message);
            setPoints(result.current_points);
            setMakeupCardInfo(prev => prev ? { ...prev, makeup_cards: result.makeup_cards, current_points: result.current_points, can_buy: result.current_points >= prev.card_price } : null);
            setShowBuyModal(false);
        } catch (error: any) {
            console.error('购买失败:', error);
            alert(error.response?.data?.error || '购买失败');
        } finally {
            setBuyLoading(false);
        }
    };

    // 生成日历
    const renderCalendar = () => {
        const now = new Date();
        const year = now.getFullYear();
        const month = now.getMonth();
        const firstDay = new Date(year, month, 1).getDay();
        const daysInMonth = new Date(year, month + 1, 0).getDate();
        const today = now.getDate();

        const days = [];
        // 填充空白
        for (let i = 0; i < firstDay; i++) {
            days.push(<div key={`empty-${i}`} className="calendar-day empty"></div>);
        }
        // 填充日期
        for (let d = 1; d <= daysInMonth; d++) {
            const dateStr = `${year}-${String(month + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
            const isCheckedIn = calendarDates.includes(dateStr);
            const isToday = d === today;
            const isFuture = d > today;
            const canMakeup = !isCheckedIn && !isFuture && !isToday && makeupCardInfo?.makeup_dates?.includes(dateStr);

            days.push(
                <div
                    key={d}
                    className={`calendar-day ${isCheckedIn ? 'checked' : ''} ${isToday ? 'today' : ''} ${isFuture ? 'future' : ''} ${canMakeup ? 'can-makeup' : ''}`}
                    title={isCheckedIn ? '已签到' : (isToday ? '今天' : (canMakeup ? '点击补签' : ''))}
                    onClick={(e) => {
                        e.stopPropagation();
                        e.preventDefault();
                        if (canMakeup) {
                            handleMakeup(dateStr, e);
                        }
                    }}
                    style={canMakeup ? { cursor: 'pointer' } : undefined}
                >
                    {d}
                    {isCheckedIn && <span className="check-mark">✓</span>}
                    {canMakeup && <img src="/icon/buqianka.png" alt="补签" className="makeup-icon" />}
                </div>
            );
        }
        return days;
    };

    const monthNames = ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月'];
    const now = new Date();

    return (
        <div className="quick-points-card" onClick={onOpenProfile}>
            {/* 积分显示 */}
            <div className="quick-points-info">
                <div className="quick-points-value">
                    <img src="/icon/jifen.png" alt="积分" className="points-icon-img" />
                    <span className="points-number">{points.toLocaleString()}</span>
                </div>
            </div>

            {/* 签到按钮 */}
            <div
                className="checkin-wrapper"
                onMouseEnter={handleMouseEnter}
                onMouseLeave={handleMouseLeave}
                ref={calendarRef}
            >
            <button
                className={`quick-checkin-btn ${checkedIn ? 'checked' : ''} ${loading ? 'loading' : ''}`}
                onClick={handleCheckIn}
                disabled={checkedIn || loading}
                    title={checkedIn ? `连续签到 ${continuousDays} 天` : '每日签到领积分'}
            >
                {loading ? (
                    <span className="checkin-loading">⏳</span>
                ) : checkedIn ? (
                    <>
                            <img src="/icon/yiqiandao.png" alt="已签到" className="checkin-icon-img" />
                            <span className="checkin-days">{continuousDays}天</span>
                    </>
                ) : (
                    <>
                            <img src="/icon/lianxuqiandao.png" alt="签到" className="checkin-icon-img" />
                        <span className="checkin-text">签到</span>
                    </>
                )}
            </button>

                {/* 签到日历弹窗 */}
                {showCalendar && (
                    <div
                        className="checkin-calendar-popup"
                        onMouseEnter={handleCalendarMouseEnter}
                        onMouseLeave={handleCalendarMouseLeave}
                    >
                        <div className="calendar-header">
                            <span className="calendar-title">{now.getFullYear()}年{monthNames[now.getMonth()]}</span>
                            <span className="calendar-stats">本月签到 {calendarDates.filter(d => d.startsWith(`${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`)).length} 天</span>
                        </div>
                        <div className="calendar-weekdays">
                            <span>日</span><span>一</span><span>二</span><span>三</span><span>四</span><span>五</span><span>六</span>
                        </div>
                        <div className="calendar-grid">
                            {calendarLoading ? (
                                <div className="calendar-loading">加载中...</div>
                            ) : (
                                renderCalendar()
                            )}
                        </div>
                        <div className="calendar-footer">
                            <span className="streak-info">🔥 连续签到 {continuousDays} 天</span>
                            <span className="makeup-card-info" onClick={(e) => { e.stopPropagation(); setShowBuyModal(true); }} style={{ cursor: 'pointer' }}>
                                🎫 补签卡: {makeupCardInfo?.makeup_cards || 0}
                            </span>
                        </div>
                        {/* 购买补签卡弹窗 */}
                        {showBuyModal && (
                            <div className="buy-modal" onClick={(e) => e.stopPropagation()}>
                                <div className="buy-modal-content">
                                    <h3>购买补签卡</h3>
                                    <p>价格: {makeupCardInfo?.card_price || 500} 积分/张</p>
                                    <p>当前积分: {makeupCardInfo?.current_points || points}</p>
                                    <p>已有补签卡: {makeupCardInfo?.makeup_cards || 0} 张</p>
                                    <div className="buy-modal-actions">
                                        <button onClick={(e) => { e.stopPropagation(); setShowBuyModal(false); }} className="cancel-btn">取消</button>
                                        <button
                                            onClick={handleBuyCard}
                                            disabled={!makeupCardInfo?.can_buy || buyLoading}
                                            className="buy-btn"
                                        >
                                            {buyLoading ? '购买中...' : '购买1张'}
                                        </button>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                )}
            </div>

            {/* 签到成功弹框 */}
            {showCheckInModal && checkInResult && (
                <CheckInSuccessModal
                    checkInResult={checkInResult}
                    onClose={() => setShowCheckInModal(false)}
                />
            )}
        </div>
    );
};
