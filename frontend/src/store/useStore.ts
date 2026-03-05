import { create } from 'zustand';
import type { Article, User } from '../types';

type Theme = 'light' | 'dark';

interface AppState {
    currentArticle: Article | null;
    isPlaying: boolean;
    currentTime: number;
    duration: number;
    user: User | null;
    token: string | null;
    theme: Theme;
    setCurrentArticle: (article: Article | null) => void;
    setIsPlaying: (isPlaying: boolean) => void;
    setCurrentTime: (time: number) => void;
    setDuration: (duration: number) => void;
    setUser: (user: User | null) => void;
    setToken: (token: string | null) => void;
    setTheme: (theme: Theme) => void;
    logout: () => void;
    showUserProfile: boolean;
    setShowUserProfile: (show: boolean) => void;
}

// 从localStorage读取主题，默认为light
const getInitialTheme = (): Theme => {
    const savedTheme = localStorage.getItem('theme') as Theme;
    return savedTheme === 'dark' ? 'dark' : 'light';
};

export const useStore = create<AppState>((set) => {
    const initialTheme = getInitialTheme();
    
    // 初始化时应用主题
    if (typeof document !== 'undefined') {
        document.documentElement.setAttribute('data-theme', initialTheme);
    }
    
    return {
        currentArticle: null,
        isPlaying: false,
        currentTime: 0,
        duration: 0,
        user: null,
        token: localStorage.getItem('token'),
        theme: initialTheme,
        setCurrentArticle: (article) => set({ currentArticle: article }),
        setIsPlaying: (isPlaying) => set({ isPlaying }),
        setCurrentTime: (time) => set({ currentTime: time }),
        setDuration: (duration) => set({ duration }),
        setUser: (user) => set({ user }),
        setToken: (token) => {
            if (token) {
                localStorage.setItem('token', token);
            } else {
                localStorage.removeItem('token');
            }
            set({ token });
        },
        setTheme: (theme) => {
            localStorage.setItem('theme', theme);
            if (typeof document !== 'undefined') {
                document.documentElement.setAttribute('data-theme', theme);
            }
            set({ theme });
        },
        logout: () => {
            localStorage.removeItem('token');
            set({ user: null, token: null });
        },
        showUserProfile: false,
        setShowUserProfile: (show) => set({ showUserProfile: show }),
    };
});
