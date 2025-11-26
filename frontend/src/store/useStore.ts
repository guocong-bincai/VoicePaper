import { create } from 'zustand';
import type { Article } from '../types';

interface AppState {
    currentArticle: Article | null;
    isPlaying: boolean;
    currentTime: number;
    duration: number;
    setCurrentArticle: (article: Article | null) => void;
    setIsPlaying: (isPlaying: boolean) => void;
    setCurrentTime: (time: number) => void;
    setDuration: (duration: number) => void;
}

export const useStore = create<AppState>((set) => ({
    currentArticle: null,
    isPlaying: false,
    currentTime: 0,
    duration: 0,
    setCurrentArticle: (article) => set({ currentArticle: article }),
    setIsPlaying: (isPlaying) => set({ isPlaying }),
    setCurrentTime: (time) => set({ currentTime: time }),
    setDuration: (duration) => set({ duration }),
}));
