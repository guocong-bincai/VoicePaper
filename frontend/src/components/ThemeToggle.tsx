import React from 'react';
import { useStore } from '../store/useStore';
import './ThemeToggle.css';

export const ThemeToggle: React.FC = () => {
    const { theme, setTheme } = useStore();

    const toggleTheme = () => {
        const newTheme = theme === 'light' ? 'dark' : 'light';
        setTheme(newTheme);
    };

    return (
        <button
            className="theme-toggle"
            onClick={toggleTheme}
            aria-label={theme === 'light' ? '切换到深色模式' : '切换到浅色模式'}
            title={theme === 'light' ? '切换到深色模式' : '切换到浅色模式'}
        >
            {theme === 'light' ? (
                // 月亮图标 - 深色模式
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path
                        d="M17.293 13.293C16.378 14.207 15.19 14.736 13.95 14.88C12.71 15.024 11.46 14.777 10.38 14.18C9.3 13.583 8.44 12.663 7.92 11.55C7.4 10.437 7.25 9.19 7.48 8C7.71 6.81 8.31 5.72 9.2 4.9C10.09 4.08 11.22 3.58 12.41 3.48C13.6 3.38 14.78 3.68 15.78 4.33C16.78 4.98 17.55 5.95 18 7.1"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        fill="currentColor"
                    />
                </svg>
            ) : (
                // 太阳图标 - 浅色模式
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <circle cx="10" cy="10" r="3.5" stroke="currentColor" strokeWidth="1.5" fill="none"/>
                    <path d="M10 2V4M10 16V18M18 10H16M4 10H2M15.657 4.343L14.243 5.757M5.757 14.243L4.343 15.657M15.657 15.657L14.243 14.243M5.757 5.757L4.343 4.343" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                </svg>
            )}
        </button>
    );
};

