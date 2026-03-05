import { useState } from 'react';
import { createPortal } from 'react-dom';
import { X, Send, Check, MessageCircle } from 'lucide-react';
import './FeedbackWidget.css';
import { submitFeedback } from '../services/api';

type FeedbackType = 'feature' | 'bug' | 'ui' | 'other';

interface FeedbackFormData {
  type: FeedbackType;
  description: string;
  contact: string;
}

interface FeedbackWidgetProps {
  isOpen: boolean;
  onClose: () => void;
}

export const FeedbackWidget: React.FC<FeedbackWidgetProps> = ({ isOpen, onClose }) => {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [formData, setFormData] = useState<FeedbackFormData>({
    type: 'feature',
    description: '',
    contact: '',
  });

  const feedbackTypes = [
    {
      value: 'feature' as const,
      label: '功能建议',
      desc: '建议新功能或改进',
      icon: '/icon/gongneng.png'
    },
    {
      value: 'bug' as const,
      label: 'Bug反馈',
      desc: '报告错误或问题',
      icon: '/icon/bug.png'
    },
    {
      value: 'ui' as const,
      label: '界面问题',
      desc: '界面显示或交互问题',
      icon: '/icon/jiemian.png'
    },
    {
      value: 'other' as const,
      label: '其他',
      desc: '其他意见或建议',
      icon: '/icon/qita.png'
    },
  ];

  const handleClose = () => {
    onClose();
    // 延迟重置表单，等动画结束
    setTimeout(() => {
      setFormData({ type: 'feature', description: '', contact: '' });
      setIsSuccess(false);
    }, 300);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.description.trim()) {
      alert('请填写问题描述');
      return;
    }

    setIsSubmitting(true);

    try {
      await submitFeedback(formData);
      setIsSuccess(true);

      // 2秒后自动关闭
      setTimeout(() => {
        handleClose();
      }, 2000);
    } catch (error) {
      console.error('提交反馈失败:', error);
      alert('提交失败，请稍后重试');
    } finally {
      setIsSubmitting(false);
    }
  };

  // ESC键关闭
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      handleClose();
    }
  };

  if (!isOpen) return null;

  return createPortal(
    <div className="feedback-overlay" onClick={handleClose} onKeyDown={handleKeyDown}>
      <div className="feedback-modal" onClick={(e) => e.stopPropagation()}>
        {/* 头部 */}
        <div className="feedback-header">
          <div className="feedback-header-content">
            <MessageCircle size={24} />
            <h2>问题反馈</h2>
          </div>
          <button
            className="feedback-close"
            onClick={handleClose}
            title="关闭 (ESC)"
          >
            <X size={20} />
          </button>
        </div>

        {/* 成功状态 */}
        {isSuccess ? (
          <div className="feedback-success">
            <div className="feedback-success-icon">
              <div className="success-icon-ring"></div>
              <Check size={48} />
            </div>
            <h3>感谢您的反馈！</h3>
            <p>我们会认真处理您的建议</p>
          </div>
        ) : (
          /* 表单 */
          <form onSubmit={handleSubmit} className="feedback-form">
            {/* 问题类型 */}
            <div className="feedback-field">
              <label className="feedback-label">问题类型</label>
              <div className="feedback-type-grid">
                {feedbackTypes.map((type) => (
                  <button
                    key={type.value}
                    type="button"
                    className={`feedback-type-item ${formData.type === type.value ? 'active' : ''}`}
                    onClick={() => setFormData({ ...formData, type: type.value })}
                  >
                    <div className="feedback-type-label">
                      <img src={type.icon} alt={type.label} className="feedback-type-icon" />
                      {type.label}
                    </div>
                    <div className="feedback-type-desc">{type.desc}</div>
                  </button>
                ))}
              </div>
            </div>

            {/* 问题描述 */}
            <div className="feedback-field">
              <label className="feedback-label">
                问题描述 <span className="feedback-required">*</span>
              </label>
              <textarea
                className="feedback-textarea"
                placeholder="请详细描述您遇到的问题或建议..."
                rows={6}
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                required
              />
            </div>

            {/* 联系方式 */}
            <div className="feedback-field">
              <label className="feedback-label">
                联系方式 <span className="feedback-optional">(可选)</span>
              </label>
              <input
                type="text"
                className="feedback-input"
                placeholder="邮箱或微信，方便我们回复您"
                value={formData.contact}
                onChange={(e) => setFormData({ ...formData, contact: e.target.value })}
              />
            </div>

            {/* 提交按钮 */}
            <div className="feedback-actions">
              <button
                type="button"
                className="feedback-btn feedback-btn-cancel"
                onClick={handleClose}
                disabled={isSubmitting}
              >
                取消
              </button>
              <button
                type="submit"
                className="feedback-btn feedback-btn-submit"
                disabled={isSubmitting}
              >
                {isSubmitting ? (
                  <>
                    <span className="feedback-spinner"></span>
                    提交中...
                  </>
                ) : (
                  <>
                    <Send size={16} />
                    提交反馈
                  </>
                )}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>,
    document.body
  );
};
