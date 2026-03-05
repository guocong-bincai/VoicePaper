import './FloatingBackButton.css';

interface FloatingBackButtonProps {
  onBack: () => void;
  show: boolean;
}

export function FloatingBackButton({ onBack, show }: FloatingBackButtonProps) {
  if (!show) return null;

  return (
    <button
      className="floating-back-button"
      onClick={onBack}
      title="返回上一级"
      aria-label="返回上一级"
    >
      <img src="/fanhui.png" alt="返回" className="back-icon" />
    </button>
  );
}
