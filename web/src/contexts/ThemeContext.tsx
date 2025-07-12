import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { PaletteMode } from '@mui/material';
import { createAppTheme } from '../theme/theme';
import { config } from '../config/env';

interface ThemeContextType {
  mode: PaletteMode;
  toggleColorMode: () => void;
  setColorMode: (mode: PaletteMode) => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

interface ThemeContextProviderProps {
  children: ReactNode;
}

// Local storage key for theme preference
const THEME_STORAGE_KEY = 'wastebin-theme-mode';

// Get initial theme mode
const getInitialThemeMode = (): PaletteMode => {
  // Check if dark mode is enabled in config
  if (!config.enableDarkMode) {
    return 'light';
  }

  // Check local storage first
  const storedMode = localStorage.getItem(THEME_STORAGE_KEY) as PaletteMode | null;
  if (storedMode && (storedMode === 'light' || storedMode === 'dark')) {
    return storedMode;
  }

  // Fall back to system preference
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    return 'dark';
  }

  return 'light';
};

/**
 * Theme context provider with dark mode support and system preference detection
 */
export const ThemeContextProvider: React.FC<ThemeContextProviderProps> = ({ children }) => {
  const [mode, setMode] = useState<PaletteMode>(getInitialThemeMode);

  // Listen for system theme changes
  useEffect(() => {
    if (!config.enableDarkMode) return;

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    
    const handleChange = (e: MediaQueryListEvent) => {
      // Only auto-switch if user hasn't set a preference
      const storedMode = localStorage.getItem(THEME_STORAGE_KEY);
      if (!storedMode) {
        setMode(e.matches ? 'dark' : 'light');
      }
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, []);

  // Save theme preference to localStorage
  useEffect(() => {
    localStorage.setItem(THEME_STORAGE_KEY, mode);
  }, [mode]);

  const toggleColorMode = () => {
    if (!config.enableDarkMode) return;
    setMode(prevMode => prevMode === 'light' ? 'dark' : 'light');
  };

  const setColorMode = (newMode: PaletteMode) => {
    if (!config.enableDarkMode && newMode === 'dark') return;
    setMode(newMode);
  };

  const theme = createAppTheme(mode);

  const contextValue: ThemeContextType = {
    mode,
    toggleColorMode,
    setColorMode,
  };

  return (
    <ThemeContext.Provider value={contextValue}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </ThemeProvider>
    </ThemeContext.Provider>
  );
};

/**
 * Hook to use theme context
 */
export const useThemeMode = (): ThemeContextType => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useThemeMode must be used within a ThemeContextProvider');
  }
  return context;
};

export default ThemeContextProvider;