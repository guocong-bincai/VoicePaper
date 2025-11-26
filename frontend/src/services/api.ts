import axios from 'axios';
import type { Article } from '../types';

const API_BASE_URL = 'http://localhost:8080/api/v1';

export const api = axios.create({
    baseURL: API_BASE_URL,
});

export const getArticles = async () => {
    const response = await api.get<Article[]>('/articles');
    return response.data;
};

export const getArticle = async (id: number) => {
    const response = await api.get<Article>(`/articles/${id}`);
    return response.data;
};

export const createArticle = async (title: string, content: string) => {
    const response = await api.post<Article>('/articles', { title, content });
    return response.data;
};

export const getAudioUrl = (path: string) => {
    // path is like "data/audio/xxx.mp3"
    // We need to convert it to "http://localhost:8080/audio/xxx.mp3"
    // The backend serves /audio mapped to ./data/audio
    if (!path) return '';
    const filename = path.split('/').pop();
    return `http://localhost:8080/audio/${filename}`;
};
