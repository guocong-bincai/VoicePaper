import { useEffect, useState, useCallback, useRef } from 'react';
import { AudioPlayer } from './components/AudioPlayer';
import { ArticleViewer } from './components/ArticleViewer';
import { ArticleSidebar } from './components/ArticleSidebar';
import { LoginPage } from './components/LoginPage';
import { LandingPage } from './components/LandingPage';
import { LoadingPage } from './components/LoadingPage';
import { UserProfile } from './components/UserProfile';
import { DictationPage } from './components/DictationPage';
import { CategorySelection } from './components/CategorySelection';
import { CategoryArticleList } from './components/CategoryArticleList';
import { FeedbackWidget } from './components/FeedbackWidget';
import { FloatingActionBadge } from './components/FloatingActionBadge';
import { QuickPointsCard } from './components/QuickPointsCard';
import { RankingButton } from './components/RankingButton';
import { VocabularyPage } from './components/VocabularyPage';
import { WordbookPage } from './components/WordbookPage';
import { FloatingBackButton } from './components/FloatingBackButton';
import { YearEndReportModal } from './components/YearEndReportModal';
import { useStore } from './store/useStore';
import { getArticles, getArticle, getCurrentUser, getMyPoints, getReadingStats, getReadingProgress, syncReadingDuration, exportArticlePDF, getVocabularyStats } from './services/api';
import type { Article } from './types';

function App() {
  const { currentArticle, setCurrentArticle, user, token, setUser, setToken, showUserProfile, setShowUserProfile, theme, setTheme } = useStore();
  const [articles, setArticles] = useState<Article[]>([]);
  const [loading, setLoading] = useState(true);
  const [checkingAuth, setCheckingAuth] = useState(true);
  const [showLoginPage, setShowLoginPage] = useState(false);
  const [isRegisterMode, setIsRegisterMode] = useState(false);
  const [showDictationPage, setShowDictationPage] = useState(false);
  const [showCategorySelection, setShowCategorySelection] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<{id: number, name: string} | null>(null);
  const [showFeedback, setShowFeedback] = useState(false);
  const [showYearReport, setShowYearReport] = useState(false);
  const [forceScrollFn, setForceScrollFn] = useState<(() => void) | null>(null);
  const [autoScrollEnabled, setAutoScrollEnabled] = useState(true);
  const [userLevel, setUserLevel] = useState<number>(1);
  const [userDisplayName, setUserDisplayName] = useState<string>('');
  const [userDisplayRarity, setUserDisplayRarity] = useState<string>('');
  const [userTitleId, setUserTitleId] = useState<number | undefined>(undefined);
  const [isImmersiveMode, setIsImmersiveMode] = useState(false);
  const [totalReadingTime, setTotalReadingTime] = useState<number>(0);
  const [currentArticleReadTime, setCurrentArticleReadTime] = useState<number>(0);
  const [showVocabularyPage, setShowVocabularyPage] = useState(false);
  const [showWordbookPage, setShowWordbookPage] = useState(false);
  const [todayReviewCount, setTodayReviewCount] = useState<number>(0);

  // 获取生词复习统计
  const fetchReviewStats = useCallback(async () => {
    if (user) {
      try {
        const stats = await getVocabularyStats();
        setTodayReviewCount(stats.today_review);
      } catch (err) {
        console.warn('获取复习统计失败:', err);
      }
    }
  }, [user]);

  // 初始加载和用户变化时获取统计
  useEffect(() => {
    if (user) {
      fetchReviewStats();
    }

    // 窗口获得焦点时刷新（防止多端操作或长时间挂起后数据过期）
    const handleFocus = () => {
      if (user) {
        fetchReviewStats();
      }
    };
    window.addEventListener('focus', handleFocus);
    return () => window.removeEventListener('focus', handleFocus);
  }, [user]); // 只依赖 user，避免 fetchReviewStats 导致的循环

  // 自动弹出年终总结（每次刷新/登录后显示）
  const hasAutoPopped = useRef(false);
  useEffect(() => {
    if (user && !hasAutoPopped.current) {
      // 延迟一点弹出，避免和加载冲突
      const timer = setTimeout(() => {
        setShowYearReport(true);
        hasAutoPopped.current = true;
      }, 1000);
      return () => clearTimeout(timer);
    }
  }, [user]);

  // 格式化阅读时间
  const formatReadingTime = (seconds: number): string => {
    if (seconds < 60) return `${seconds}秒`;
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}分钟`;
    const hours = Math.floor(minutes / 60);
    const remainingMinutes = minutes % 60;
    if (remainingMinutes === 0) return `${hours}小时`;
    return `${hours}小时${remainingMinutes}分`;
  };

  // 回调函数：从ArticleViewer接收定位函数
  const handleForceScrollToHighlight = useCallback((fn: () => void) => {
    setForceScrollFn(() => fn);
  }, []);

  // 进入沉浸式全屏模式
  const enterImmersiveMode = useCallback(() => {
    setIsImmersiveMode(true);
    // 请求浏览器全屏
    const elem = document.documentElement;
    if (elem.requestFullscreen) {
      elem.requestFullscreen();
    } else if ((elem as any).webkitRequestFullscreen) {
      (elem as any).webkitRequestFullscreen();
    } else if ((elem as any).msRequestFullscreen) {
      (elem as any).msRequestFullscreen();
    }
    // 自动定位到当前高亮位置
    setTimeout(() => {
      if (forceScrollFn) {
        forceScrollFn();
      }
    }, 300); // 等待全屏动画完成
  }, [forceScrollFn]);

  // 退出沉浸式全屏模式
  const exitImmersiveMode = useCallback(() => {
    setIsImmersiveMode(false);
    // 退出浏览器全屏
    if (document.fullscreenElement) {
      document.exitFullscreen();
    } else if ((document as any).webkitFullscreenElement) {
      (document as any).webkitExitFullscreen();
    } else if ((document as any).msFullscreenElement) {
      (document as any).msExitFullscreen();
    }
  }, []);

  // 监听全屏状态变化（用户按ESC或点击浏览器退出全屏时同步状态）
  useEffect(() => {
    const handleFullscreenChange = () => {
      const isFullscreen = !!(document.fullscreenElement || (document as any).webkitFullscreenElement || (document as any).msFullscreenElement);
      if (!isFullscreen && isImmersiveMode) {
        setIsImmersiveMode(false);
      }
    };

    document.addEventListener('fullscreenchange', handleFullscreenChange);
    document.addEventListener('webkitfullscreenchange', handleFullscreenChange);
    document.addEventListener('MSFullscreenChange', handleFullscreenChange);

    return () => {
      document.removeEventListener('fullscreenchange', handleFullscreenChange);
      document.removeEventListener('webkitfullscreenchange', handleFullscreenChange);
      document.removeEventListener('MSFullscreenChange', handleFullscreenChange);
    };
  }, [isImmersiveMode]);

  // 沉浸式模式下：用户滑动后3秒无操作自动回滚到高亮位置
  useEffect(() => {
    if (!isImmersiveMode) return;

    let scrollTimeout: ReturnType<typeof setTimeout> | null = null;

    const handleScroll = () => {
      // 清除之前的定时器
      if (scrollTimeout) {
        clearTimeout(scrollTimeout);
      }
      // 3秒后自动回滚到高亮位置
      scrollTimeout = setTimeout(() => {
        if (forceScrollFn && isImmersiveMode) {
          forceScrollFn();
        }
      }, 3000);
    };

    window.addEventListener('scroll', handleScroll, { passive: true });

    return () => {
      window.removeEventListener('scroll', handleScroll);
      if (scrollTimeout) {
        clearTimeout(scrollTimeout);
      }
    };
  }, [isImmersiveMode, forceScrollFn]);

  // 回调函数：从ArticleViewer接收自动滚动状态
  const handleAutoScrollChange = useCallback((enabled: boolean) => {
    setAutoScrollEnabled(enabled);
  }, []);

  // 处理返回逻辑：支持多层导航
  const handleBackButton = () => {
    // 从最深层开始返回
    if (selectedCategory) {
      // 如果在分类详情，返回到分类选择
      setSelectedCategory(null);
    } else if (showDictationPage) {
      // 如果在默写页面，返回
      setShowDictationPage(false);
    } else if (showCategorySelection) {
      // 如果在分类选择，返回到阅读页面
      setShowCategorySelection(false);
    } else if (showWordbookPage) {
      // 如果在单词书，返回到阅读页面
      setShowWordbookPage(false);
    } else if (showVocabularyPage) {
      // 如果在生词本，返回到阅读页面
      setShowVocabularyPage(false);
    } else if (showUserProfile) {
      // 如果在个人中心，返回到阅读页面
      setShowUserProfile(false);
    }
  };

  // 浏览器返回键处理：子页面打开时推送历史，返回时关闭页面
  useEffect(() => {
    const handlePopState = () => {
      handleBackButton();
    };

    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, [showVocabularyPage, showWordbookPage, showUserProfile, showDictationPage, showCategorySelection, selectedCategory]);

  // 子页面打开时推送历史记录
  useEffect(() => {
    if (showVocabularyPage || showWordbookPage || showUserProfile || showDictationPage || showCategorySelection || selectedCategory) {
      window.history.pushState({ isSubPage: true }, '');
    }
  }, [showVocabularyPage, showWordbookPage, showUserProfile, showDictationPage, showCategorySelection, selectedCategory]);

  // 判断是否显示浮动返回按钮
  const isAnySubPageOpen = showVocabularyPage || showWordbookPage || showUserProfile || showDictationPage || showCategorySelection || !!selectedCategory;
  const shouldShowBackButton = isAnySubPageOpen;

  // 获取用户等级和阅读统计
  useEffect(() => {
    const fetchUserData = async () => {
      if (user) {
        try {
          const [pointsData, readingStats] = await Promise.all([
            getMyPoints(),
            getReadingStats().catch(() => ({ total_reading_time: 0 }))
          ]);
          setUserLevel(pointsData.level || 1);
          setUserDisplayName((pointsData as any).display_name || '');
          // 获取穿戴称号的稀有度
          const equippedTitle = (pointsData as any).equipped_title;
          setUserDisplayRarity(equippedTitle?.rarity || '');
          setUserTitleId(equippedTitle?.id || undefined);
          const totalSeconds = readingStats.total_reading_time || 0;
          setTotalReadingTime(totalSeconds);

          // 同步阅读时长到排名系统（转换为秒）
          if (totalSeconds > 0) {
            syncReadingDuration(totalSeconds).catch(err => {
              console.log('同步阅读时长失败:', err);
            });
          }
        } catch (err) {
          console.log('获取用户数据失败');
        }
      }
    };

    fetchUserData();
  }, [user]);

  // 检查登录状态和处理OAuth回调
  useEffect(() => {
    // 处理OAuth回调（从URL参数获取token）
    const urlParams = new URLSearchParams(window.location.search);
    const tokenParam = urlParams.get('token');
    const errorParam = urlParams.get('error');

    if (errorParam) {
      console.error('OAuth登录失败:', errorParam);
      // 清除URL参数
      window.history.replaceState({}, '', window.location.pathname);
      setShowLoginPage(true);
      return;
    }

    if (tokenParam) {
      // 处理OAuth登录（函数内部会保存token）
      handleOAuthLogin(tokenParam);
      // 清除URL参数
      window.history.replaceState({}, '', window.location.pathname);
      return;
    }

    // 如果没有OAuth回调，正常检查登录状态
    checkAuth();
  }, []);

  const handleOAuthLogin = async (tokenValue: string) => {
    try {
      // 先保存 token，这样 getCurrentUser 才能正确发送请求
      setToken(tokenValue);
      // 等待 token 保存到 localStorage
      await new Promise(resolve => setTimeout(resolve, 100));
      // 然后获取用户信息
      const userData = await getCurrentUser();
      setUser(userData);
      setCheckingAuth(false); // 重要：设置 checkingAuth 为 false，这样才能触发文章加载
      console.log('✅ OAuth登录成功:', userData);
    } catch (error) {
      console.error('❌ OAuth登录失败:', error);
      setToken(null);
      setUser(null);
      setCheckingAuth(false);
      setShowLoginPage(true);
    }
  };

  // 如果已登录，加载文章
  useEffect(() => {
    if (user && !checkingAuth && articles.length === 0) {
      loadArticles();
    }
  }, [user, checkingAuth]);

  const checkAuth = async () => {
    if (token) {
      try {
        const userData = await getCurrentUser();
        setUser(userData);
      } catch (error) {
        // token无效，清除
        setToken(null);
        setUser(null);
      }
    }
    setCheckingAuth(false);
  };

  const loadArticles = async () => {
    try {
      setLoading(true);
      const articlesList = await getArticles();
      setArticles(articlesList);

      console.log('✅ 文章列表加载成功，共', articlesList.length, '篇');

      // 默认加载第一篇文章
      if (articlesList.length > 0) {
        await loadArticle(articlesList[0].id);
      } else {
        console.warn('⚠️  数据库中没有文章，请先添加文章数据');
        setLoading(false);
      }
    } catch (error) {
      console.error("❌ 加载文章列表失败:", error);
      setLoading(false);
    }
  };

  const loadArticle = async (id: number) => {
    try {
      const article = await getArticle(id);
      setCurrentArticle(article);
      console.log('✅ 文章详情加载成功:', article.title);
      setLoading(false);

      // 获取当前文章的阅读进度
      if (user) {
        try {
          const progress = await getReadingProgress(id);
          setCurrentArticleReadTime(progress.total_time || 0);
        } catch {
          setCurrentArticleReadTime(0);
        }
      }
    } catch (error) {
      console.error("❌ 加载文章详情失败:", error);
      setLoading(false);
    }
  };

  // 检查登录状态
  if (checkingAuth) {
    return <LoadingPage message="正在验证身份..." />;
  }

  // 未登录，显示首页或登录页面
  if (!user) {
    if (showLoginPage) {
      return (
        <LoginPage
          initialMode={isRegisterMode ? 'register' : 'login'}
          onBack={() => setShowLoginPage(false)}
        />
      );
    }
    return (
      <LandingPage
        onShowLogin={() => {
          setIsRegisterMode(false);
          setShowLoginPage(true);
        }}
        onShowRegister={() => {
          setIsRegisterMode(true);
          setShowLoginPage(true);
        }}
      />
    );
  }

  if (loading) {
    return <LoadingPage message="正在为您准备精彩内容..." />;
  }

  if (articles.length === 0) {
    return (
      <div style={{ padding: '2rem', textAlign: 'center' }}>
        <p>暂无文章</p>
        <p style={{ fontSize: '0.9rem', color: '#666', marginTop: '1rem' }}>
          请先在数据库中添加文章数据
        </p>
      </div>
    );
  }

  // 如果显示单词书或生词本，渲染对应页面
  if (showWordbookPage) {
    return <WordbookPage onClose={() => setShowWordbookPage(false)} />;
  }

  if (showVocabularyPage) {
    return <VocabularyPage onClose={() => setShowVocabularyPage(false)} />;
  }

  // 如果显示个人中心，渲染个人中心页面
  if (showUserProfile) {
    return (
      <div style={{ backgroundColor: 'var(--bg-color)', minHeight: '100vh', padding: '2rem 0', transition: 'background-color 0.3s ease' }}>
        <div style={{ maxWidth: '1200px', margin: '0 auto', padding: '0 1rem' }}>
          <UserProfile onBack={() => setShowUserProfile(false)} />
        </div>
      </div>
    );
  }

  // 如果显示默写练习页面，直接返回DictationPage，不要包裹额外的div
  if (showDictationPage && currentArticle) {
    return (
      <DictationPage
        articleId={currentArticle.id}
        onBack={() => {
          setShowDictationPage(false);
          // 如果是从分类列表进来的，不需要重新加载所有文章，直接返回列表即可
          if (!selectedCategory) {
            loadArticles();
          }
        }}
      />
    );
  }

  // 如果显示合集选择页面
  if (showCategorySelection) {
    // 如果已选择分类，显示该分类下的文章列表
    if (selectedCategory) {
      return (
        <CategoryArticleList
          categoryId={selectedCategory.id}
          categoryName={selectedCategory.name}
          onSelectArticle={async (articleId) => {
            try {
              setLoading(true);
              await loadArticle(articleId);
              // 加载完成后显示默写页面
              setShowDictationPage(true);
              // 保持 selectedCategory 状态，以便从默写页面返回时能回到列表
            } catch (error) {
              console.error('❌ 加载文章失败:', error);
              setLoading(false);
            }
          }}
          onBack={() => setSelectedCategory(null)}
        />
      );
    }

    // 显示分类选择列表
    return (
      <CategorySelection
        onSelectCategory={(categoryId, categoryName) => {
          setSelectedCategory({ id: categoryId, name: categoryName || '文章列表' });
        }}
        onBack={() => setShowCategorySelection(false)}
      />
    );
  }

  return (
    <div className={isImmersiveMode ? 'immersive-mode' : ''} style={{ backgroundColor: 'var(--bg-color)', minHeight: '100vh', transition: 'background-color 0.3s ease' }}>
      {/* 浮动返回按钮 */}
      <FloatingBackButton show={shouldShowBackButton} onBack={handleBackButton} />

      {/* 顶部导航 (极简) */}
      <nav className="top-nav">
        <button
          className="logo logo-clickable"
          onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}
          title={theme === 'light' ? '切换到深色模式' : '切换到浅色模式'}
        >
          <img src="/logo.png" alt="VoicePaper" className="logo-image" />
          VoicePaper<span className="dot">.</span>
        </button>
        <div className="nav-right">
          <div className="nav-info">
            {articles.length > 0 ? `双语精读 · 第 ${articles.length} 期` : '暂无文章'}
            {totalReadingTime > 0 && <span className="total-read-time">· 累计 {formatReadingTime(totalReadingTime)}</span>}
          </div>
          {user && (
            <div className="user-menu">
              {/* 排名入口 */}
              <RankingButton />

              {/* 单词书入口 */}
              <button
                className={`wordbook-btn ${showWordbookPage ? 'active' : ''}`}
                onClick={() => {
                  setShowWordbookPage(true);
                  setShowVocabularyPage(false);
                  setShowCategorySelection(false);
                }}
                title="高效单词书"
              >
                <img src="/icon/dancishu/book.png" alt="" className="nav-icon" />
                <span>高效单词书</span>
              </button>

              {/* 句乐部入口 */}
              <button
                className={`wordbook-btn ${showCategorySelection ? 'active' : ''}`}
                onClick={() => {
                  setShowCategorySelection(true);
                  setShowWordbookPage(false);
                  setShowVocabularyPage(false);
                }}
                title="句乐部"
              >
                <img src="/icon/julebu/juzi.png" alt="" className="nav-icon" />
                <span>句乐部</span>
              </button>

              {/* 生词本入口 */}
              <button
                className={`vocabulary-btn ${showVocabularyPage ? 'active' : ''}`}
                onClick={() => {
                  setShowVocabularyPage(true);
                  setShowWordbookPage(false);
                  setShowCategorySelection(false);
                }}
                title="我的生词本"
                style={{ position: 'relative' }}
              >
                <img src="/icon/shengciben.png" alt="" className="nav-icon" />
                <span>生词本</span>
                {todayReviewCount > 0 && (
                  <span className="nav-badge">{todayReviewCount > 99 ? '99+' : todayReviewCount}</span>
                )}
              </button>

              {/* 积分和签到快捷入口 */}
              <QuickPointsCard onOpenProfile={() => setShowUserProfile(true)} />



              {/* 用户头像 */}
              <button
                className="user-avatar-btn"
                onClick={() => setShowUserProfile(true)}
                title="个人中心"
              >
                <img
                  key={user.avatar || 'default-avatar'}
                  src={user.avatar || `https://ui-avatars.com/api/?name=${encodeURIComponent(user.nickname || user.email?.split('@')[0] || 'U')}&background=2563EB&color=fff&size=40&bold=true`}
                  alt={user.nickname || user.email}
                  className="user-avatar"
                  onError={(e) => {
                      // 如果头像加载失败，等待300ms后重试一次
                      const target = e.target as HTMLImageElement;
                      const originalSrc = user.avatar;
                      const defaultAvatar = `https://ui-avatars.com/api/?name=${encodeURIComponent(user.nickname || user.email?.split('@')[0] || 'U')}&background=2563EB&color=fff&size=40&bold=true`;

                      // 如果还没重试过，等待后重试
                      if (!target.dataset.retried && originalSrc) {
                          target.dataset.retried = 'true';
                          setTimeout(() => {
                              target.src = originalSrc + (originalSrc.includes('?') ? '&' : '?') + '_t=' + Date.now();
                          }, 500);
                      } else if (target.src !== defaultAvatar) {
                          target.src = defaultAvatar;
                          console.warn('⚠️ 顶部导航头像加载失败，使用默认头像:', user.avatar);
                      }
                  }}
                />
              </button>
            </div>
          )}
        </div>
      </nav>

      <div className="app-container">
        {/* 文章阅读区 */}
        <main className="reading-area">
          <article className="paper-content">
            {/* 封面图/装饰区域 */}
            <div className="article-header-decoration">
              <div className="category-tag">
                {currentArticle?.category?.name || 'ARTICLE'}
              </div>
              <div className="header-right-actions">
                <div className="read-time">⏱ 本文 {formatReadingTime(currentArticleReadTime)}</div>
                {/* 导出PDF按钮 */}
                <button
                  className="theme-btn" // 复用样式
                  onClick={() => {
                    if (currentArticle) {
                      exportArticlePDF(currentArticle.id, currentArticle.title);
                    }
                  }}
                  title="导出精读PDF"
                >
                  <img
                    src="/icon/PDF.png"
                    alt="Export PDF"
                    style={{ width: '20px', height: '20px', objectFit: 'contain' }}
                  />
                </button>
                {/* 主题切换按钮 */}
                <button
                  className="theme-btn"
                  onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}
                  title={theme === 'light' ? '切换到深色模式' : '切换到浅色模式'}
                >
                  {theme === 'light' ? (
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
                    </svg>
                  ) : (
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <circle cx="12" cy="12" r="5"/>
                      <line x1="12" y1="1" x2="12" y2="3"/>
                      <line x1="12" y1="21" x2="12" y2="23"/>
                      <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
                      <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
                      <line x1="1" y1="12" x2="3" y2="12"/>
                      <line x1="21" y1="12" x2="23" y2="12"/>
                      <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
                      <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
                    </svg>
                  )}
                </button>
                {/* 沉浸式阅读按钮 */}
                <button
                  className="immersive-btn"
                  onClick={enterImmersiveMode}
                  title="沉浸式阅读（全屏）"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/>
                  </svg>
                </button>
              </div>
            </div>

            <ArticleViewer
              onForceScrollToHighlight={handleForceScrollToHighlight}
              onAutoScrollChange={handleAutoScrollChange}
            />
          </article>
        </main>

        {/* 浮动操作徽章 - 融合等级+定位+反馈 */}
        {user && !showDictationPage && (
          <FloatingActionBadge
            level={userLevel}
            displayName={userDisplayName}
            displayRarity={userDisplayRarity}
            titleId={userTitleId}
            onLocateReading={forceScrollFn || undefined}
            onOpenFeedback={() => setShowFeedback(true)}
            onOpenProfile={() => setShowUserProfile(true)}
            showLocateButton={!autoScrollEnabled}
          />
        )}
      </div>

      {/* 底部悬浮播放器 */}
      <AudioPlayer />

      {/* 文章列表侧边栏 */}
      <ArticleSidebar
        articles={articles}
        currentArticleId={currentArticle?.id}
        onSelectArticle={loadArticle}
        onStartDictation={() => setShowCategorySelection(true)}
      />

      {/* 问题反馈弹窗 */}
      <FeedbackWidget
        isOpen={showFeedback}
        onClose={() => setShowFeedback(false)}
      />

      {/* 年终总结弹窗 */}
      <YearEndReportModal
        isOpen={showYearReport}
        onClose={() => setShowYearReport(false)}
      />

      {/* 生词本页面 */}
      {showVocabularyPage && (
        <VocabularyPage onClose={() => {
          setShowVocabularyPage(false);
          fetchReviewStats(); // 关闭时刷新待复习数量
        }} />
      )}

      {/* 沉浸式模式控制按钮 */}
      {isImmersiveMode && (
        <div className="immersive-controls">
          {/* 主题切换按钮 */}
          <button
            className="immersive-theme-btn"
            onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}
            title={theme === 'light' ? '切换到深色模式' : '切换到浅色模式'}
          >
            {theme === 'light' ? (
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
              </svg>
            ) : (
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <circle cx="12" cy="12" r="5"/>
                <line x1="12" y1="1" x2="12" y2="3"/>
                <line x1="12" y1="21" x2="12" y2="23"/>
                <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
                <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
                <line x1="1" y1="12" x2="3" y2="12"/>
                <line x1="21" y1="12" x2="23" y2="12"/>
                <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
                <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
              </svg>
            )}
          </button>
          {/* 退出全屏按钮 */}
          <button
            className="exit-immersive-btn"
            onClick={exitImmersiveMode}
            title="退出全屏 (ESC)"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3"/>
            </svg>
          </button>
        </div>
      )}
    </div>
  );
}

export default App;
