import React, { useState, useRef, useCallback } from 'react';
import { useStore } from '../store/useStore';
import { updateUserProfile, logout, uploadAvatar, getMyPoints } from '../services/api';
import { ThemeToggle } from './ThemeToggle';
import { PointsDisplay } from './PointsDisplay';
import { CheckInButton } from './CheckInButton';
import { TitleBadges } from './TitleBadges';
import { LevelBadge } from './LevelBadge';
import ReactCrop, { makeAspectCrop, centerCrop } from 'react-image-crop';
import type { Crop, PixelCrop } from 'react-image-crop';
import 'react-image-crop/dist/ReactCrop.css';
import './UserProfile.css';

interface UserProfileProps {
    onBack?: () => void;
}

export const UserProfile: React.FC<UserProfileProps> = ({ onBack }) => {
    const { user, setUser, setToken } = useStore();
    const [isEditing, setIsEditing] = useState(false);
    const [nickname, setNickname] = useState(user?.nickname || '');
    const [bio, setBio] = useState(user?.bio || '');
    const [avatar, setAvatar] = useState(user?.avatar || '');
    const [avatarModified, setAvatarModified] = useState(false); // 跟踪用户是否手动修改了头像URL
    const [loading, setLoading] = useState(false);
    const [uploading, setUploading] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [activeTab, setActiveTab] = useState<'profile' | 'points' | 'titles'>('profile');
    const [userLevel, setUserLevel] = useState<number>(1);
    const [titleDisplayName, setTitleDisplayName] = useState<string>('');    // 穿戴的称号名称
    const [titleDisplayRarity, setTitleDisplayRarity] = useState<string>(''); // 穿戴称号的稀有度
    const [titleId, setTitleId] = useState<number | undefined>(undefined);    // 穿戴称号的ID
    const fileInputRef = useRef<HTMLInputElement>(null);
    
    // 图片裁剪相关状态
    const [showCropModal, setShowCropModal] = useState(false);
    const [imgSrc, setImgSrc] = useState('');
    const [crop, setCrop] = useState<Crop>();
    const [completedCrop, setCompletedCrop] = useState<PixelCrop>();
    const imgRef = useRef<HTMLImageElement>(null);
    const canvasRef = useRef<HTMLCanvasElement>(null);

    // 当用户信息更新时，同步本地状态
    React.useEffect(() => {
        if (user) {
            setNickname(user.nickname || '');
            setBio(user.bio || '');
            setAvatar(user.avatar || '');
        }
    }, [user]);

    // 获取用户等级和穿戴称号信息
    const fetchUserTitleInfo = React.useCallback(async () => {
        try {
            const pointsData = await getMyPoints();
            setUserLevel(pointsData.level || 1);
            // 获取穿戴称号信息
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            setTitleDisplayName((pointsData as any).display_name || '');
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const equippedTitle = (pointsData as any).equipped_title;
            setTitleDisplayRarity(equippedTitle?.rarity || '');
            setTitleId(equippedTitle?.id || undefined);
        } catch {
            console.log('获取用户等级失败，使用默认等级1');
            setUserLevel(1);
        }
    }, []);

    React.useEffect(() => {
        if (user) {
            fetchUserTitleInfo();
        }
    }, [user, fetchUserTitleInfo]);

    const handleSave = async () => {
        if (!user) return;

        setLoading(true);
        setError('');
        setSuccess('');

        try {
            // 构建请求数据
            // BUG修复: 只发送用户实际修改的字段，避免签名URL被存储到数据库
            // 修复日期: 2024-12-12
            const updateData: { nickname?: string; avatar?: string; bio?: string } = {};
            
            // 昵称：总是发送（用户可能想清空或修改）
            updateData.nickname = nickname.trim() || user.nickname || '';
            
            // 头像：只有当用户手动修改了头像URL输入框时才发送
            // 避免把签名URL（带过期时间戳）覆盖原始OSS URL
            if (avatarModified && avatar.trim()) {
                updateData.avatar = avatar.trim();
            }
            // 如果用户没有修改头像，不发送avatar字段，让后端保持原值
            
            // 简介：总是发送（用户可能想清空或修改）
            updateData.bio = bio.trim() || user.bio || '';

            console.log('📤 发送更新请求:', updateData);
            const updatedUser = await updateUserProfile(updateData);
            console.log('✅ 更新成功:', updatedUser);
            setUser(updatedUser);
            setSuccess('资料更新成功！');
            setIsEditing(false);
            setAvatarModified(false); // 重置头像修改标记
        } catch (err: any) {
            console.error('❌ 更新失败:', err);
            const errorMsg = err.response?.data?.error || err.response?.data?.details || err.message || '更新失败';
            setError(errorMsg);
        } finally {
            setLoading(false);
        }
    };

    // 压缩图片到指定大小以内
    const compressImage = (file: File, maxSizeMB: number = 5): Promise<string> => {
        return new Promise((resolve, reject) => {
            const maxSize = maxSizeMB * 1024 * 1024;
            
            // 如果文件已经小于目标大小，直接返回
            if (file.size <= maxSize) {
                const reader = new FileReader();
                reader.onload = () => resolve(reader.result as string);
                reader.onerror = reject;
                reader.readAsDataURL(file);
                return;
            }

            // 需要压缩
            const img = new Image();
            const reader = new FileReader();
            
            reader.onload = (e) => {
                img.onload = () => {
                    const canvas = document.createElement('canvas');
                    const ctx = canvas.getContext('2d');
                    if (!ctx) {
                        reject(new Error('无法创建Canvas上下文'));
                        return;
                    }

                    // 计算压缩比例：根据文件大小动态调整
                    const ratio = Math.sqrt(maxSize / file.size);
                    let quality = Math.min(0.9, ratio);
                    
                    // 如果图片尺寸过大，也需要缩小尺寸
                    let width = img.width;
                    let height = img.height;
                    const maxDimension = 2048; // 最大尺寸限制
                    
                    if (width > maxDimension || height > maxDimension) {
                        if (width > height) {
                            height = Math.round(height * maxDimension / width);
                            width = maxDimension;
                        } else {
                            width = Math.round(width * maxDimension / height);
                            height = maxDimension;
                        }
                    }

                    canvas.width = width;
                    canvas.height = height;
                    ctx.drawImage(img, 0, 0, width, height);

                    // 尝试不同质量级别直到满足大小要求
                    const tryCompress = (q: number): void => {
                        canvas.toBlob((blob) => {
                            if (!blob) {
                                reject(new Error('压缩失败'));
                                return;
                            }
                            
                            console.log(`🗜️ 压缩尝试: 质量=${q.toFixed(2)}, 大小=${(blob.size / 1024 / 1024).toFixed(2)}MB`);
                            
                            if (blob.size <= maxSize || q <= 0.1) {
                                // 压缩成功或已到最低质量
                                const compressedReader = new FileReader();
                                compressedReader.onload = () => {
                                    console.log(`✅ 图片压缩完成: ${(file.size / 1024 / 1024).toFixed(2)}MB → ${(blob.size / 1024 / 1024).toFixed(2)}MB`);
                                    resolve(compressedReader.result as string);
                                };
                                compressedReader.onerror = reject;
                                compressedReader.readAsDataURL(blob);
                            } else {
                                // 继续降低质量
                                tryCompress(q - 0.1);
                            }
                        }, 'image/jpeg', q);
                    };

                    tryCompress(quality);
                };
                img.onerror = () => reject(new Error('图片加载失败'));
                img.src = e.target?.result as string;
            };
            reader.onerror = reject;
            reader.readAsDataURL(file);
        });
    };

    // 处理文件选择，显示裁剪界面
    const handleFileSelect = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (!file) return;

        // 验证文件类型
        const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
        if (!allowedTypes.includes(file.type)) {
            setError('不支持的文件类型，仅支持：jpg, jpeg, png, gif, webp');
            return;
        }

        setError('');
        
        try {
            // 如果图片超过5MB，自动压缩
        if (file.size > 5 * 1024 * 1024) {
                setSuccess(`正在压缩图片 (${(file.size / 1024 / 1024).toFixed(1)}MB)...`);
                const compressedDataUrl = await compressImage(file, 5);
                setSuccess('图片压缩完成！');
                setImgSrc(compressedDataUrl);
                setShowCropModal(true);
                // 清除提示信息
                setTimeout(() => setSuccess(''), 2000);
            } else {
                // 直接读取文件
        const reader = new FileReader();
        reader.addEventListener('load', () => {
            setImgSrc(reader.result?.toString() || '');
            setShowCropModal(true);
        });
        reader.readAsDataURL(file);
            }
        } catch (err) {
            console.error('图片处理失败:', err);
            setError('图片处理失败，请尝试其他图片');
        }
    };

    // 初始化裁剪区域（1:1比例，居中）
    const onImageLoad = useCallback((e: React.SyntheticEvent<HTMLImageElement>) => {
        const { width, height } = e.currentTarget;
        const crop = makeAspectCrop(
            {
                unit: '%',
                width: 90,
            },
            1, // 1:1 比例
            width,
            height
        );
        setCrop(centerCrop(crop, width, height));
    }, []);

    // 将裁剪后的图片转换为Blob并上传
    const handleCropComplete = async () => {
        if (!imgRef.current || !completedCrop || !canvasRef.current) {
            setError('请先选择并调整图片位置');
            return;
        }

        setUploading(true);
        setError('');
        setSuccess('');

        try {
            // 获取裁剪后的图片数据
            const image = imgRef.current;
            const crop = completedCrop;
            const canvas = canvasRef.current;
            const scaleX = image.naturalWidth / image.width;
            const scaleY = image.naturalHeight / image.height;

            canvas.width = crop.width;
            canvas.height = crop.height;

            const ctx = canvas.getContext('2d');
            if (!ctx) {
                throw new Error('无法获取canvas上下文');
            }

            ctx.drawImage(
                image,
                crop.x * scaleX,
                crop.y * scaleY,
                crop.width * scaleX,
                crop.height * scaleY,
                0,
                0,
                crop.width,
                crop.height
            );

            // 将canvas转换为Blob
            canvas.toBlob(async (blob) => {
                if (!blob) {
                    setError('图片处理失败');
                    setUploading(false);
                    return;
                }

                // 创建File对象
                const file = new File([blob], 'avatar.jpg', { type: 'image/jpeg' });

                console.log('📤 开始上传裁剪后的头像:', file.name, file.size);
                const result = await uploadAvatar(file);
                console.log('✅ 头像上传成功:', result);
                
                // 更新用户信息（保留所有字段）
                if (result.user) {
                    setUser({ ...user, ...result.user });
                }
                setAvatar(result.avatar_url);
                setSuccess('头像上传成功！');
                setShowCropModal(false);
                setImgSrc('');
                setCrop(undefined);
                setCompletedCrop(undefined);
            }, 'image/jpeg', 0.9);
        } catch (err: any) {
            console.error('❌ 头像上传失败:', err);
            const errorMsg = err.response?.data?.error || err.response?.data?.details || err.message || '上传失败';
            setError(errorMsg);
        } finally {
            setUploading(false);
            // 清空文件输入，允许重复选择同一文件
            if (fileInputRef.current) {
                fileInputRef.current.value = '';
            }
        }
    };

    const handleAvatarClick = () => {
        if (isEditing && fileInputRef.current) {
            fileInputRef.current.click();
        }
    };

    const handleLogout = async () => {
        try {
            await logout();
            setToken(null);
            setUser(null);
            window.location.href = '/';
        } catch (err) {
            console.error('登出失败:', err);
            // 即使API失败，也清除本地状态
            setToken(null);
            setUser(null);
            window.location.href = '/';
        }
    };

    if (!user) {
        return null;
    }

    const displayName = user.nickname || user.email?.split('@')[0] || '用户';
    const displayAvatar = user.avatar || `https://ui-avatars.com/api/?name=${encodeURIComponent(displayName)}&background=2563EB&color=fff&size=128&bold=true`;

    return (
        <>
            <div className="user-profile">
                {/* 返回按钮 - 精致图标按钮 */}
                {onBack && (
                    <button className="profile-back-btn" onClick={onBack} aria-label="返回">
                        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M12.5 15L7.5 10L12.5 5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        </svg>
                    </button>
                )}
                
                {/* 顶部信息卡片 */}
                <div className="profile-header-card">
                    <div className="profile-header">
                        <div className="profile-avatar-container">
                            <div 
                                className={`avatar-wrapper ${isEditing ? 'editable' : ''}`}
                                onClick={handleAvatarClick}
                                style={{ cursor: isEditing ? 'pointer' : 'default' }}
                            >
                            <img 
                                src={displayAvatar} 
                                alt={displayName}
                                className="profile-avatar"
                                onError={(e) => {
                                    // 如果头像加载失败，使用默认头像
                                    const target = e.target as HTMLImageElement;
                                    if (target.src !== displayAvatar) {
                                        // 避免无限循环
                                        return;
                                    }
                                    const defaultAvatar = `https://ui-avatars.com/api/?name=${encodeURIComponent(displayName)}&background=667eea&color=fff&size=140&bold=true`;
                                    target.src = defaultAvatar;
                                    console.warn('⚠️ 头像加载失败，使用默认头像:', displayAvatar);
                                }}
                            />
                                {isEditing && (
                                    <div className="avatar-overlay">
                                        {uploading ? (
                                            <div className="uploading-indicator">上传中...</div>
                                        ) : (
                                            <div className="upload-hint">点击上传</div>
                                        )}
                                    </div>
                                )}
                            </div>
                            {/* 等级徽章 - 显示在头像下方 */}
                            {!isEditing && (
                                <div className="avatar-badge-container">
                                    <LevelBadge 
                                        level={userLevel} 
                                        size="medium" 
                                        showLabel={true}
                                        displayName={titleDisplayName}
                                        displayRarity={titleDisplayRarity}
                                        titleId={titleId}
                                    />
                                </div>
                            )}
                            <input
                                ref={fileInputRef}
                                type="file"
                                accept="image/jpeg,image/jpg,image/png,image/gif,image/webp"
                                onChange={handleFileSelect}
                                style={{ display: 'none' }}
                                disabled={uploading}
                            />
                            {isEditing && (
                                <div className="avatar-edit-hint">
                                    <p>或输入头像URL</p>
                                    <input
                                        type="text"
                                        value={avatar}
                                        onChange={(e) => {
                                            setAvatar(e.target.value);
                                            setAvatarModified(true); // 标记用户手动修改了头像URL
                                        }}
                                        placeholder="输入头像URL"
                                        className="avatar-url-input"
                                    />
                                </div>
                            )}
                        </div>
                        <div className="profile-info">
                            {isEditing ? (
                                <>
                                    <input
                                        type="text"
                                        value={nickname}
                                        onChange={(e) => setNickname(e.target.value)}
                                        placeholder="昵称"
                                        className="nickname-input"
                                    />
                                    <textarea
                                        value={bio}
                                        onChange={(e) => setBio(e.target.value)}
                                        placeholder="个人简介"
                                        className="bio-input"
                                        rows={3}
                                    />
                                </>
                            ) : (
                                <>
                                    <h2 className="profile-name">{displayName}</h2>
                                    {user.bio && <p className="profile-bio">{user.bio}</p>}
                                    {/* 邀请码显示 - 简洁样式 */}
                                    {user.invite_code && (
                                        <div className="profile-invite-code">
                                            <div className="invite-code-label">
                                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                    <path d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
                                                </svg>
                                                <span>邀请码</span>
                                            </div>
                                            <div className="invite-code-compact">
                                                <code className="invite-code-text">{user.invite_code}</code>
                                                <button
                                                    className="invite-code-copy-compact"
                                                    onClick={async () => {
                                                        try {
                                                            await navigator.clipboard.writeText(user.invite_code || '');
                                                            setSuccess('邀请码已复制！');
                                                            setTimeout(() => setSuccess(''), 2000);
                                                        } catch (err) {
                                                            const textArea = document.createElement('textarea');
                                                            textArea.value = user.invite_code || '';
                                                            textArea.style.position = 'fixed';
                                                            textArea.style.opacity = '0';
                                                            document.body.appendChild(textArea);
                                                            textArea.select();
                                                            try {
                                                                document.execCommand('copy');
                                                                setSuccess('邀请码已复制！');
                                                                setTimeout(() => setSuccess(''), 2000);
                                                            } catch (e) {
                                                                setError('复制失败');
                                                            }
                                                            document.body.removeChild(textArea);
                                                        }
                                                    }}
                                                    title="复制邀请码"
                                                >
                                                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                                                    </svg>
                                                </button>
                                            </div>
                                        </div>
                                    )}
                                </>
                            )}
                        </div>
                    </div>
                </div>

                {/* Tab切换 - 放在顶部 */}
                <div className="profile-tabs">
                    <button
                        className={`profile-tab ${activeTab === 'profile' ? 'active' : ''}`}
                        onClick={() => setActiveTab('profile')}
                    >
                        <img src="/icon/geren.png" alt="" className="tab-icon" /> 个人资料
                    </button>
                    <button
                        className={`profile-tab ${activeTab === 'points' ? 'active' : ''}`}
                        onClick={() => setActiveTab('points')}
                    >
                        <img src="/icon/jifen.png" alt="" className="tab-icon" /> 积分签到
                    </button>
                    <button
                        className={`profile-tab ${activeTab === 'titles' ? 'active' : ''}`}
                        onClick={() => setActiveTab('titles')}
                    >
                        <img src="/icon/chengjiu.png" alt="" className="tab-icon" /> 称号成就
                    </button>
                </div>

                {/* Tab内容 */}
                <div className="profile-tab-content">
                    {activeTab === 'profile' && (
                        <div className="profile-content">
                            <div className="profile-details">
                                <div className="detail-item">
                                    <svg className="detail-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                                    </svg>
                                    <span className="detail-label">邮箱</span>
                                    <span className="detail-value">{user.email || '未绑定'}</span>
                                </div>
                                <div className="detail-item">
                                    <svg className="detail-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                    </svg>
                                    <span className="detail-label">注册时间</span>
                                    <span className="detail-value">
                                        {user.created_at ? new Date(user.created_at).toLocaleDateString('zh-CN', {
                                            year: 'numeric',
                                            month: '2-digit',
                                            day: '2-digit'
                                        }) : '未知'}
                                    </span>
                                </div>
                                <div className="detail-item">
                                    <svg className="detail-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
                                    </svg>
                                    <span className="detail-label">阅读主题</span>
                                    <div className="detail-value" style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end' }}>
                                        <ThemeToggle />
                                    </div>
                                </div>
                            </div>

                            {error && <div className="error-message">{error}</div>}
                            {success && <div className="success-message">{success}</div>}

                            <div className="profile-actions">
                            {isEditing ? (
                                <>
                                    <button
                                        className="btn btn-primary"
                                        onClick={handleSave}
                                        disabled={loading}
                                    >
                                        {loading ? '保存中...' : '保存'}
                                    </button>
                                    <button
                                        className="btn btn-secondary"
                                        onClick={() => {
                                            setIsEditing(false);
                                            setNickname(user.nickname || '');
                                            setBio(user.bio || '');
                                            setAvatar(user.avatar || '');
                                            setAvatarModified(false); // 重置头像修改标记
                                            setError('');
                                            setSuccess('');
                                        }}
                                        disabled={loading}
                                    >
                                        取消
                                    </button>
                                </>
                            ) : (
                                <button
                                    className="btn btn-primary"
                                    onClick={() => setIsEditing(true)}
                                >
                                    编辑资料
                                </button>
                            )}
                            <button
                                className="btn btn-danger"
                                onClick={handleLogout}
                            >
                                退出登录
                            </button>
                            </div>
                        </div>
                    )}
                    {activeTab === 'points' && (
                        <div className="points-section">
                            <PointsDisplay />
                            <div style={{ marginTop: '24px' }}>
                                <CheckInButton />
                            </div>
                        </div>
                    )}
                    {activeTab === 'titles' && (
                        <div className="titles-section">
                            <TitleBadges onTitleChange={fetchUserTitleInfo} />
                        </div>
                    )}
                </div>
            </div>

            {/* 图片裁剪模态框 */}
            {showCropModal && (
                <div className="crop-modal-overlay" onClick={() => !uploading && setShowCropModal(false)}>
                    <div className="crop-modal" onClick={(e) => e.stopPropagation()}>
                        <div className="crop-modal-header">
                            <h3>调整头像位置</h3>
                            <button 
                                className="crop-modal-close"
                                onClick={() => setShowCropModal(false)}
                                disabled={uploading}
                            >
                                ×
                            </button>
                        </div>
                        <div className="crop-modal-content">
                            {imgSrc && (
                                <ReactCrop
                                    crop={crop}
                                    onChange={(_, percentCrop) => setCrop(percentCrop)}
                                    onComplete={(c) => setCompletedCrop(c)}
                                    aspect={1}
                                    minWidth={100}
                                    minHeight={100}
                                >
                                    <img
                                        ref={imgRef}
                                        alt="Crop me"
                                        src={imgSrc}
                                        style={{ maxWidth: '100%', maxHeight: '70vh' }}
                                        onLoad={onImageLoad}
                                    />
                                </ReactCrop>
                            )}
                            <canvas
                                ref={canvasRef}
                                style={{ display: 'none' }}
                            />
                        </div>
                        <div className="crop-modal-actions">
                            <button
                                className="btn btn-secondary"
                                onClick={() => {
                                    setShowCropModal(false);
                                    setImgSrc('');
                                    setCrop(undefined);
                                    setCompletedCrop(undefined);
                                }}
                                disabled={uploading}
                            >
                                取消
                            </button>
                            <button
                                className="btn btn-primary"
                                onClick={handleCropComplete}
                                disabled={uploading || !completedCrop}
                            >
                                {uploading ? '上传中...' : '确认上传'}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </>
    );
};
