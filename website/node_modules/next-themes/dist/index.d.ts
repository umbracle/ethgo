import React from 'react';
interface UseThemeProps {
    themes: string[];
    setTheme: (theme: string) => void;
    theme?: string;
    forcedTheme?: string;
    resolvedTheme?: string;
    systemTheme?: 'dark' | 'light';
}
export declare const useTheme: () => UseThemeProps;
interface ValueObject {
    [themeName: string]: string;
}
export interface ThemeProviderProps {
    forcedTheme?: string;
    disableTransitionOnChange?: boolean;
    enableSystem?: boolean;
    storageKey?: string;
    themes?: string[];
    defaultTheme?: string;
    attribute?: string;
    value?: ValueObject;
}
export declare const ThemeProvider: React.FC<ThemeProviderProps>;
export {};
