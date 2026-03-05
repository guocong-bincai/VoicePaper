import React, { useEffect, useState, useRef } from 'react';
import html2canvas from 'html2canvas';
import { getYearEndReport } from '../services/api';
import type { YearEndReport } from '../services/api';
import './YearEndReportModal.css';

interface YearEndReportModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const YearEndReportModal: React.FC<YearEndReportModalProps> = ({ isOpen, onClose }) => {
  const [report, setReport] = useState<YearEndReport | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [downloading, setDownloading] = useState(false);
  const reportRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (isOpen && !report && !error) {
      setLoading(true);
      setError(null);
      getYearEndReport()
        .then(data => {
          setReport(data);
          setError(null);
        })
        .catch(err => {
          console.error("Failed to load year end report:", err);
          setError('加载年终报告失败，请稍后重试');
          // 不自动关闭，给用户一个重试的机会
        })
        .finally(() => {
          setLoading(false);
        });
    }
  }, [isOpen, report, error]);

  const handleDownload = async () => {
    if (!reportRef.current) return;

    try {
      setDownloading(true);

      // 等待图片完全加载
      const img = reportRef.current.querySelector('.year-report-bg') as HTMLImageElement;
      if (img && !img.complete) {
        await new Promise((resolve) => {
          img.onload = resolve;
        });
      }

      // 使用 html2canvas 截图
      const canvas = await html2canvas(reportRef.current, {
        useCORS: true, // 允许跨域图片
        scale: 2, // 提高清晰度
        backgroundColor: '#000000', // 强制黑色背景
        logging: false,
        onclone: (clonedDoc) => {
          // html2canvas 不支持 backdrop-filter，卡片背景设为完全透明
          const items = clonedDoc.querySelectorAll('.report-item') as NodeListOf<HTMLElement>;
          items.forEach(item => {
            // 完全透明背景，只保留边框
            item.style.background = 'transparent';
            item.style.border = '1px solid rgba(0, 255, 255, 0.2)';
            item.style.boxShadow = 'none';
            item.style.backdropFilter = 'none';
          });

          // 用户信息条 - 更轻的背景
          const userInfo = clonedDoc.querySelector('.user-info-tag') as HTMLElement;
          if (userInfo) {
            userInfo.style.background = 'rgba(0, 0, 0, 0.3)';
            userInfo.style.backdropFilter = 'none';
          }

          // 确保全站排名没有边框
          const rankTitle = clonedDoc.querySelector('.rank-title') as HTMLElement;
          if (rankTitle) {
            rankTitle.style.border = 'none';
          }

          // html2canvas 不支持 background-clip: text，需要将渐变文字改为纯色
          const rankValue = clonedDoc.querySelector('.rank-value') as HTMLElement;
          if (rankValue) {
            rankValue.style.background = 'none';
            rankValue.style.webkitBackgroundClip = 'initial';
            rankValue.style.webkitTextFillColor = '#ffd700'; // 金色
            rankValue.style.color = '#ffd700';
            rankValue.style.textShadow = '0 0 20px rgba(255, 215, 0, 0.6), 0 2px 4px rgba(0,0,0,0.8)';
          }

          // 同样处理 rank-hash
          const rankHash = clonedDoc.querySelector('.rank-hash') as HTMLElement;
          if (rankHash) {
            rankHash.style.webkitTextFillColor = '#ffd700';
            rankHash.style.color = '#ffd700';
          }
        }
      });

      // 转换为图片并下载
      const link = document.createElement('a');
      link.download = `VoicePaper_2025年度总结_${report?.nickname || 'user'}.png`;
      link.href = canvas.toDataURL('image/png');
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (error) {
      console.error('Download failed:', error);
      alert('保存图片失败，请重试或截屏保存');
    } finally {
      setDownloading(false);
    }
  };

  if (!isOpen) return null;

  const formatDuration = (minutes: number) => {
    if (minutes < 60) return `${minutes}分钟`;
    const hours = (minutes / 60).toFixed(1);
    return `${hours}小时`;
  };

  const handleRetry = () => {
    setError(null);
    setReport(null);
    setLoading(true);
    getYearEndReport()
      .then(data => {
        setReport(data);
        setError(null);
      })
      .catch(err => {
        console.error("Failed to load year end report:", err);
        setError('加载年终报告失败，请稍后重试');
      })
      .finally(() => {
        setLoading(false);
      });
  };

  return (
    <div className="year-report-overlay" onClick={onClose}>
      <div className="year-report-wrapper" onClick={e => e.stopPropagation()}>
        <div className="year-report-container" ref={reportRef}>
          {loading && <div className="year-report-loading">正在生成年度报告...</div>}

          {error && (
            <div className="year-report-error" style={{
              padding: '40px',
              textAlign: 'center',
              color: '#ff6b6b',
              fontSize: '16px'
            }}>
              <div style={{ marginBottom: '20px' }}>{error}</div>
              <button onClick={handleRetry} style={{
                padding: '10px 20px',
                marginRight: '10px',
                background: '#4CAF50',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer'
              }}>重试</button>
              <button onClick={onClose} style={{
                padding: '10px 20px',
                background: '#999',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer'
              }}>关闭</button>
            </div>
          )}

          {!loading && !error && report && (
            <div className="year-report-content">
              <img
                src="/icon/nianzhong/nianzhong.png"
                alt="2025 Annual Report"
                className="year-report-bg"
                crossOrigin="anonymous"
              />

              {/* 顶部数据区域： Grid布局 */}
              <div className="report-grid-area">

                {/* 1. 已读外刊 */}
                <div className="report-item item-articles">
                  <div className="report-label">已读外刊</div>
                  <div className="report-value text-cyan">
                    {report.articles_read}<span className="unit">篇</span>
                  </div>
                </div>

                {/* 2. 连续打卡 */}
                <div className="report-item item-checkin">
                  <div className="report-label">连续打卡</div>
                  <div className="report-value text-purple">
                    {report.consecutive_days}<span className="unit">天</span>
                  </div>
                </div>

                {/* 3. 学习生词 */}
                <div className="report-item item-words">
                  <div className="report-label">学习生词</div>
                  <div className="report-value text-green">
                    {report.words_learned}<span className="unit">个</span>
                  </div>
                </div>

                {/* 4. 累计时长 */}
                <div className="report-item item-duration">
                  <div className="report-label">累计时长</div>
                  <div className="report-value text-orange">
                    {formatDuration(report.total_duration)}
                  </div>
                </div>

                {/* 5. 累计积分 */}
                <div className="report-item item-points">
                  <div className="report-label">累计积分</div>
                  <div className="report-value text-yellow">
                    {report.total_points}<span className="unit">分</span>
                  </div>
                </div>

              </div>

              {/* 底部排名区域：位于奖杯处 */}
              <div className="report-rank-area">
                <div className="rank-title">全站排名</div>
                <div className="rank-value">
                  <span className="rank-hash">NO.</span>
                  {report.global_rank}
                </div>
              </div>

              {/* 用户信息 */}
              <div className="user-info-tag">
                <span className="user-nickname">@{report.nickname}</span> · 2025年度总结
              </div>
            </div>
          )}
        </div>

        {/* Action Buttons (Outside the screenshot area) */}
        {!loading && report && (
          <div className="report-actions">
            <button className="action-btn download-btn" onClick={handleDownload} disabled={downloading}>
              {downloading ? '保存中...' : '保存海报'}
            </button>
            <button className="action-btn report-close-btn" onClick={onClose}>
              关闭
            </button>
          </div>
        )}
      </div>
    </div>
  );
};