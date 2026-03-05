import React, { useState, useEffect } from 'react';
import { useStore } from '../store/useStore';
import { sendEmailCode, loginWithEmail, loginWithPassword, verifyEmailCode, resetPassword, /* loginWithGitHub, loginWithGoogle, */ getCurrentUser, register } from '../services/api';
import './LoginPage.css';

interface LoginPageProps {
    initialMode?: 'login' | 'register';
    onBack?: () => void;
}

export const LoginPage: React.FC<LoginPageProps> = ({ initialMode = 'login', onBack }) => {
    const { setUser, setToken, token } = useStore();
    const [mode, setMode] = useState<'login' | 'register' | 'forgot_password'>(initialMode);
    const isRegister = mode === 'register';
    const isForgotPassword = mode === 'forgot_password';
    const [loginMethod, setLoginMethod] = useState<'password' | 'email'>('password');
    const [email, setEmail] = useState('');
    const [code, setCode] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [inviteCode, setInviteCode] = useState('');
    const [countdown, setCountdown] = useState(0);
    const [emailVerified, setEmailVerified] = useState(false); // 邮箱是否已验证
    const [registerCountdown, setRegisterCountdown] = useState(0); // 注册验证码倒计时
    const [forgotPasswordCountdown, setForgotPasswordCountdown] = useState(0); // 找回密码验证码倒计时
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    // 禁用页面滚动
    useEffect(() => {
        document.body.classList.add('login-page-active');
        document.documentElement.classList.add('login-page-active');
        return () => {
            document.body.classList.remove('login-page-active');
            document.documentElement.classList.remove('login-page-active');
        };
    }, []);

    // 处理OAuth回调（从URL参数获取token）
    useEffect(() => {
        const urlParams = new URLSearchParams(window.location.search);
        const tokenParam = urlParams.get('token');
        const errorParam = urlParams.get('error');

        if (errorParam) {
            setError('OAuth登录失败，请重试');
            // 清除URL参数
            window.history.replaceState({}, '', window.location.pathname);
            return;
        }

        if (tokenParam) {
            // 保存token并获取用户信息
            setToken(tokenParam);
            handleOAuthLogin(tokenParam);
            // 清除URL参数
            window.history.replaceState({}, '', window.location.pathname);
        } else if (token) {
            // 如果已有token，检查有效性
            checkAuth();
        }
    }, []);

    const handleOAuthLogin = async (tokenValue: string) => {
        try {
            const user = await getCurrentUser();
            setUser(user);
            setToken(tokenValue);
        } catch (error: any) {
            setError('获取用户信息失败');
            setToken(null);
        }
    };

    const checkAuth = async () => {
        try {
            const user = await getCurrentUser();
            setUser(user);
        } catch (error) {
            // token无效，清除
            setToken(null);
            setUser(null);
        }
    };

    const handleSendCode = async () => {
        if (!email) {
            setError('请输入邮箱地址');
            return;
        }

        setLoading(true);
        setError('');
        setSuccess('');

        try {
            let purpose = 'login';
            let targetCountdownSetter = setCountdown;

            if (isRegister) {
                purpose = 'register';
                targetCountdownSetter = setRegisterCountdown;
            } else if (isForgotPassword) {
                purpose = 'reset_password';
                targetCountdownSetter = setForgotPasswordCountdown;
            }

            await sendEmailCode(email, purpose);
            setSuccess('验证码已发送');

            // 开始倒计时
            targetCountdownSetter(60);
            const timer = setInterval(() => {
                targetCountdownSetter((prev) => {
                    if (prev <= 1) {
                        clearInterval(timer);
                        return 0;
                    }
                    return prev - 1;
                });
            }, 1000);
        } catch (error: any) {
            setError(error.response?.data?.error || error.response?.data?.details || '发送验证码失败');
        } finally {
            setLoading(false);
        }
    };

    // 验证注册时的验证码
    // BUG修复: 不再调用登录接口，改为调用专门的验证接口
    // 修复策略: 使用 verifyEmailCode 接口只验证验证码，不创建用户
    // 影响范围: frontend/src/components/LoginPage.tsx:118-160
    // 修复日期: 2025-12-10
    const handleVerifyRegisterCode = async () => {
        if (!email || !code) {
            setError('请输入邮箱和验证码');
            return;
        }

        setLoading(true);
        setError('');

        try {
            // 调用专门的验证接口（只验证验证码，不创建用户）
            await verifyEmailCode(email, code);
            // 验证成功，允许继续注册
            setEmailVerified(true);
            setError('');
        } catch (error: any) {
            const errorMsg = error.response?.data?.error || error.response?.data?.details || '验证码验证失败';
            setError(errorMsg);
        } finally {
            setLoading(false);
        }
    };

    const handleEmailLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!email || !code) {
            setError('请输入邮箱和验证码');
            return;
        }

        setLoading(true);
        setError('');

        try {
            const result = await loginWithEmail(email, code);
            setToken(result.token);
            // 先设置登录返回的用户信息
            setUser(result.user);
            // 登录成功后立即刷新用户信息，确保获取完整数据（包括签名后的头像URL）
            setTimeout(async () => {
                try {
                    const freshUser = await getCurrentUser();
                    setUser(freshUser);
                } catch (e) {
                    console.log('刷新用户信息失败，使用登录返回的数据');
                }
            }, 100);
            setError('');
        } catch (error: any) {
            setError(error.response?.data?.error || error.response?.data?.details || '登录失败');
        } finally {
            setLoading(false);
        }
    };

    const handlePasswordLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!email || !password) {
            setError('请输入邮箱和密码');
            return;
        }

        setLoading(true);
        setError('');

        try {
            const result = await loginWithPassword(email, password);
            setToken(result.token);
            // 先设置登录返回的用户信息
            setUser(result.user);
            // 登录成功后立即刷新用户信息，确保获取完整数据（包括签名后的头像URL）
            setTimeout(async () => {
                try {
                    const freshUser = await getCurrentUser();
                    setUser(freshUser);
                } catch (e) {
                    console.log('刷新用户信息失败，使用登录返回的数据');
                }
            }, 100);
            setError('');
        } catch (error: any) {
            setError(error.response?.data?.error || error.response?.data?.details || '登录失败');
        } finally {
            setLoading(false);
        }
    };

    const handleRegister = async (e: React.FormEvent) => {
        e.preventDefault();

        // 必须先验证邮箱
        if (!emailVerified) {
            setError('请先验证邮箱');
            return;
        }

        if (!email || !password || !confirmPassword) {
            setError('请填写所有必填项');
            return;
        }

        if (password.length < 6) {
            setError('密码长度至少6位');
            return;
        }

        if (password !== confirmPassword) {
            setError('两次输入的密码不一致');
            return;
        }

        setLoading(true);
        setError('');

        try {
            // 注册时必须提供验证码
            if (!code) {
                setError('请先验证邮箱');
                setLoading(false);
                return;
            }
            const result = await register(email, password, code, inviteCode || undefined);
            setToken(result.token);
            setUser(result.user);
            setError('');
        } catch (error: any) {
            setError(error.response?.data?.error || error.response?.data?.details || '注册失败');
        } finally {
            setLoading(false);
        }
    };

    const handleResetPassword = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!email || !code || !password || !confirmPassword) {
            setError('请填写所有必填项');
            return;
        }

        if (password.length < 6) {
            setError('密码长度至少6位');
            return;
        }

        if (password !== confirmPassword) {
            setError('两次输入的密码不一致');
            return;
        }

        setLoading(true);
        setError('');

        try {
            await resetPassword(email, code, password);
            setSuccess('密码重置成功，请登录');
            setTimeout(() => {
                setMode('login');
                setLoginMethod('password');
                setSuccess('');
                setPassword('');
                setConfirmPassword('');
                setCode('');
            }, 2000);
        } catch (error: any) {
            setError(error.response?.data?.error || error.response?.data?.details || '重置密码失败');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="login-page">
            <div className="login-wrapper">
                <div className="login-container">
                    {/* 返回按钮 */}
                    {onBack && (
                        <button
                            className="login-back-btn"
                            onClick={onBack}
                            title="返回首页"
                        >
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                <path d="M19 12H5M12 19l-7-7 7-7"/>
                            </svg>
                        </button>
                    )}

                    <div className="login-header">
                        <h1 className="login-logo">VoicePaper<span className="dot">.</span></h1>
                        <p className="login-subtitle">
                            {isRegister ? '创建您的账户，开始学习之旅' :
                             isForgotPassword ? '重置您的密码' : '欢迎回来，继续您的学习'}
                        </p>
                    </div>

                    {/* 登录方式切换（仅登录时显示） */}
                    {mode === 'login' && (
                        <div className="login-method-toggle">
                            <button
                                type="button"
                                className={`method-btn ${loginMethod === 'password' ? 'active' : ''}`}
                                onClick={() => setLoginMethod('password')}
                            >
                                密码登录
                            </button>
                            <button
                                type="button"
                                className={`method-btn ${loginMethod === 'email' ? 'active' : ''}`}
                                onClick={() => setLoginMethod('email')}
                            >
                                验证码登录
                            </button>
                        </div>
                    )}

                    <div className="login-content">
                        {isRegister ? (
                            !emailVerified ? (
                                // 第一步：验证邮箱
                                <div className="auth-form">
                                    <div className="form-section">
                                        <div className="form-group">
                                            <label htmlFor="register-email">邮箱地址</label>
                                            <input
                                                id="register-email"
                                                type="email"
                                                value={email}
                                                onChange={(e) => {
                                                    setEmail(e.target.value);
                                                    setEmailVerified(false);
                                                }}
                                                placeholder="your.email@example.com"
                                                disabled={loading}
                                                required
                                            />
                                            <div className="supported-emails">
                                                <span className="supported-label">支持邮箱：</span>
                                                <div className="email-providers">
                                                    <img src="/youxiang_logo/QQyouxiang.png" alt="QQ邮箱" className="email-logo" />
                                                    <img src="/youxiang_logo/163mail.png" alt="163邮箱" className="email-logo" />
                                                    <img src="/youxiang_logo/126mail.png" alt="126邮箱" className="email-logo" />
                                                    <img src="/youxiang_logo/Gmailmail.png" alt="Gmail" className="email-logo" />
                                                    <img src="/youxiang_logo/Outlookmail.png" alt="Outlook" className="email-logo" />
                                                    <img src="/youxiang_logo/xinlangmail.png" alt="新浪邮箱" className="email-logo" />
                                                    <img src="/youxiang_logo/qiyemail.png" alt="企业邮箱" className="email-logo" />
                                                </div>
                                            </div>
                                        </div>

                                        <div className="form-group">
                                            <label htmlFor="register-code">验证码</label>
                                            <div className="code-input-group">
                                                <input
                                                    id="register-code"
                                                    type="text"
                                                    value={code}
                                                    onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                                                    placeholder="6位验证码"
                                                    maxLength={6}
                                                    disabled={loading}
                                                    required
                                                />
                                                <button
                                                    type="button"
                                                    className="send-code-btn"
                                                    onClick={handleSendCode}
                                                    disabled={loading || registerCountdown > 0 || !email}
                                                >
                                                    {registerCountdown > 0 ? `${registerCountdown}秒` : '发送验证码'}
                                                </button>
                                            </div>
                                            <span className="field-hint">请先验证邮箱，验证通过后才能继续注册</span>
                                        </div>
                                    </div>

                                    {error && <div className="error-message">{error}</div>}
                                    {success && <div className="success-message">{success}</div>}

                                    <button
                                        type="button"
                                        className="auth-btn primary"
                                        onClick={handleVerifyRegisterCode}
                                        disabled={loading || !email || !code}
                                    >
                                        {loading ? '验证中...' : '验证邮箱'}
                                    </button>
                                </div>
                            ) : (
                                // 第二步：填写密码等信息
                                <form onSubmit={handleRegister} className="auth-form">
                                    <div className="form-section">
                                        <div className="form-group">
                                            <label htmlFor="register-email-display">邮箱地址</label>
                                            <input
                                                id="register-email-display"
                                                type="email"
                                                value={email}
                                                disabled
                                                style={{ background: '#f3f4f6', color: '#6b7280' }}
                                            />
                                            <span className="field-hint" style={{ color: '#10b981' }}>
                                                ✓ 邮箱已验证
                                            </span>
                                        </div>

                                        <div className="form-group">
                                            <label htmlFor="register-password">密码</label>
                                            <input
                                                id="register-password"
                                                type="password"
                                                value={password}
                                                onChange={(e) => setPassword(e.target.value)}
                                                placeholder="至少6位字符"
                                                disabled={loading}
                                                required
                                                minLength={6}
                                            />
                                            <span className="field-hint">密码长度至少6位</span>
                                        </div>

                                        <div className="form-group">
                                            <label htmlFor="confirm-password">确认密码</label>
                                            <input
                                                id="confirm-password"
                                                type="password"
                                                value={confirmPassword}
                                                onChange={(e) => setConfirmPassword(e.target.value)}
                                                placeholder="请再次输入密码"
                                                disabled={loading}
                                                required
                                                minLength={6}
                                            />
                                        </div>

                                        <div className="form-group">
                                            <label htmlFor="invite-code">
                                                邀请码
                                                <span className="optional-badge">可选</span>
                                            </label>
                                            <div className="invite-code-group">
                                                <input
                                                    id="invite-code"
                                                    type="text"
                                                    value={inviteCode}
                                                    onChange={(e) => setInviteCode(e.target.value.toUpperCase())}
                                                    placeholder="输入邀请码（可选）"
                                                    disabled={loading}
                                                    autoComplete="off"
                                                />
                                            </div>
                                        </div>
                                    </div>

                                {error && <div className="error-message">{error}</div>}
                                {success && <div className="success-message">{success}</div>}

                                <button type="submit" className="auth-btn primary" disabled={loading}>
                                    {loading ? '注册中...' : '创建账户'}
                                </button>
                            </form>
                            )
                        ) : isForgotPassword ? (
                            <form onSubmit={handleResetPassword} className="auth-form">
                                <div className="form-section">
                                    <div className="form-group">
                                        <label htmlFor="forgot-email">邮箱地址</label>
                                        <input
                                            id="forgot-email"
                                            type="email"
                                            value={email}
                                            onChange={(e) => setEmail(e.target.value)}
                                            placeholder="your.email@example.com"
                                            disabled={loading}
                                            required
                                        />
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="forgot-code">验证码</label>
                                        <div className="code-input-group">
                                            <input
                                                id="forgot-code"
                                                type="text"
                                                value={code}
                                                onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                                                placeholder="6位验证码"
                                                maxLength={6}
                                                disabled={loading}
                                                required
                                            />
                                            <button
                                                type="button"
                                                className="send-code-btn"
                                                onClick={handleSendCode}
                                                disabled={loading || forgotPasswordCountdown > 0 || !email}
                                            >
                                                {forgotPasswordCountdown > 0 ? `${forgotPasswordCountdown}秒` : '发送验证码'}
                                            </button>
                                        </div>
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="new-password">新密码</label>
                                        <input
                                            id="new-password"
                                            type="password"
                                            value={password}
                                            onChange={(e) => setPassword(e.target.value)}
                                            placeholder="至少6位字符"
                                            disabled={loading}
                                            required
                                            minLength={6}
                                        />
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="confirm-new-password">确认新密码</label>
                                        <input
                                            id="confirm-new-password"
                                            type="password"
                                            value={confirmPassword}
                                            onChange={(e) => setConfirmPassword(e.target.value)}
                                            placeholder="请再次输入新密码"
                                            disabled={loading}
                                            required
                                            minLength={6}
                                        />
                                    </div>
                                </div>

                                {error && <div className="error-message">{error}</div>}
                                {success && <div className="success-message">{success}</div>}

                                <button type="submit" className="auth-btn primary" disabled={loading}>
                                    {loading ? '重置中...' : '重置密码'}
                                </button>

                                <div className="form-footer">
                                    <button
                                        type="button"
                                        className="link-btn"
                                        onClick={() => setMode('login')}
                                    >
                                        返回登录
                                    </button>
                                </div>
                            </form>
                        ) : (
                            loginMethod === 'email' ? (
                            <form onSubmit={handleEmailLogin} className="auth-form">
                                <div className="form-section">
                                    <div className="form-group">
                                        <label htmlFor="email">邮箱地址</label>
                                        <input
                                            id="email"
                                            type="email"
                                            value={email}
                                            onChange={(e) => setEmail(e.target.value)}
                                            placeholder="your.email@example.com"
                                            disabled={loading}
                                            required
                                        />
                                        <div className="supported-emails">
                                            <span className="supported-label">支持邮箱：</span>
                                            <div className="email-providers">
                                                <img src="/youxiang_logo/QQyouxiang.png" alt="QQ邮箱" className="email-logo" />
                                                <img src="/youxiang_logo/163mail.png" alt="163邮箱" className="email-logo" />
                                                <img src="/youxiang_logo/126mail.png" alt="126邮箱" className="email-logo" />
                                                <img src="/youxiang_logo/Gmailmail.png" alt="Gmail" className="email-logo" />
                                                <img src="/youxiang_logo/Outlookmail.png" alt="Outlook" className="email-logo" />
                                                <img src="/youxiang_logo/xinlangmail.png" alt="新浪邮箱" className="email-logo" />
                                                <img src="/youxiang_logo/qiyemail.png" alt="企业邮箱" className="email-logo" />
                                            </div>
                                        </div>
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="code">验证码</label>
                                        <div className="code-input-group">
                                            <input
                                                id="code"
                                                type="text"
                                                value={code}
                                                onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                                                placeholder="6位验证码"
                                                maxLength={6}
                                                disabled={loading}
                                                required
                                            />
                                            <button
                                                type="button"
                                                className="send-code-btn"
                                                onClick={handleSendCode}
                                                disabled={loading || countdown > 0 || !email}
                                            >
                                                {countdown > 0 ? `${countdown}秒` : '发送验证码'}
                                            </button>
                                        </div>
                                    </div>
                                </div>

                                {error && <div className="error-message">{error}</div>}
                                {success && <div className="success-message">{success}</div>}

                                <button type="submit" className="auth-btn primary" disabled={loading}>
                                    {loading ? '登录中...' : '登录'}
                                </button>
                            </form>
                        ) : (
                            <form onSubmit={handlePasswordLogin} className="auth-form">
                                <div className="form-section">
                                    <div className="form-group">
                                        <label htmlFor="password-email">邮箱地址</label>
                                        <input
                                            id="password-email"
                                            type="email"
                                            value={email}
                                            onChange={(e) => setEmail(e.target.value)}
                                            placeholder="your.email@example.com"
                                            disabled={loading}
                                            required
                                        />
                                        <div className="supported-emails">
                                            <span className="supported-label">支持邮箱：</span>
                                            <div className="email-providers">
                                                <img src="/youxiang_logo/QQyouxiang.png" alt="QQ邮箱" className="email-logo" />
                                                <img src="/youxiang_logo/163mail.png" alt="163邮箱" className="email-logo" />
                                                <img src="/youxiang_logo/126mail.png" alt="126邮箱" className="email-logo" />
                                                <img src="/youxiang_logo/Gmailmail.png" alt="Gmail" className="email-logo" />
                                                <img src="/youxiang_logo/Outlookmail.png" alt="Outlook" className="email-logo" />
                                                <img src="/youxiang_logo/xinlangmail.png" alt="新浪邮箱" className="email-logo" />
                                                <img src="/youxiang_logo/qiyemail.png" alt="企业邮箱" className="email-logo" />
                                            </div>
                                        </div>
                                    </div>

                                    <div className="form-group">
                                        <div className="label-row">
                                            <label htmlFor="password">密码</label>
                                            <button
                                                type="button"
                                                className="forgot-password-link"
                                                onClick={() => setMode('forgot_password')}
                                            >
                                                忘记密码？
                                            </button>
                                        </div>
                                        <input
                                            id="password"
                                            type="password"
                                            value={password}
                                            onChange={(e) => setPassword(e.target.value)}
                                            placeholder="请输入您的密码"
                                            disabled={loading}
                                            required
                                        />
                                    </div>
                                </div>

                                {error && <div className="error-message">{error}</div>}
                                {success && <div className="success-message">{success}</div>}

                                <button type="submit" className="auth-btn primary" disabled={loading}>
                                    {loading ? '登录中...' : '登录'}
                                </button>
                            </form>
                            )
                        )}
                    </div>

                    <div className="login-footer">
                        <p>登录即表示您同意我们的<button className="link-btn">服务条款</button>和<button className="link-btn">隐私政策</button></p>
                    </div>
                </div>
            </div>
        </div>
    );
};

