/**
 * Test setup file
 * This file runs before all tests
 */

import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach, beforeAll, afterAll, vi } from 'vitest';
import { setupServer } from 'msw/node';
import { handlers } from './mocks/handlers';

// Set up test environment variables
vi.stubEnv('VITE_API_BASE_URL', 'http://localhost:3000');
vi.stubEnv('VITE_APP_TITLE', 'Wastebin Test');
vi.stubEnv('VITE_DEFAULT_LANGUAGE', 'plaintext');
vi.stubEnv('VITE_DEFAULT_EXPIRY', '60');
vi.stubEnv('VITE_ENABLE_ANALYTICS', 'false');
vi.stubEnv('VITE_ENABLE_TELEMETRY', 'false');
vi.stubEnv('VITE_MAX_PASTE_SIZE', '10485760');
vi.stubEnv('VITE_SUPPORTED_LANGUAGES', 'plaintext,javascript,python,go,rust');
vi.stubEnv('VITE_THEME_MODE', 'light');
vi.stubEnv('VITE_ENABLE_BURN_AFTER_READ', 'true');
vi.stubEnv('VITE_ENABLE_CUSTOM_EXPIRY', 'true');
vi.stubEnv('VITE_ENABLE_SYNTAX_HIGHLIGHTING', 'true');
vi.stubEnv('VITE_ENABLE_LINE_NUMBERS', 'true');
vi.stubEnv('VITE_ENABLE_WORD_WRAP', 'true');
vi.stubEnv('VITE_ENABLE_RAW_VIEW', 'true');
vi.stubEnv('VITE_ENABLE_DOWNLOAD', 'true');
vi.stubEnv('VITE_ENABLE_SHARE', 'true');
vi.stubEnv('VITE_ENABLE_COPY', 'true');
vi.stubEnv('VITE_ENABLE_QR_CODE', 'false');
vi.stubEnv('VITE_ENABLE_PRINT', 'true');
vi.stubEnv('VITE_ENABLE_FULLSCREEN', 'true');
vi.stubEnv('VITE_ENABLE_THEME_TOGGLE', 'true');
vi.stubEnv('VITE_ENABLE_SHORTCUTS', 'true');
vi.stubEnv('VITE_ENABLE_SEARCH', 'true');
vi.stubEnv('VITE_ENABLE_HISTORY', 'true');
vi.stubEnv('VITE_ENABLE_STATS', 'false');
vi.stubEnv('VITE_ENABLE_RATE_LIMITING', 'true');
vi.stubEnv('VITE_RATE_LIMIT_MAX_REQUESTS', '100');
vi.stubEnv('VITE_RATE_LIMIT_WINDOW_MS', '60000');

// Mock server setup
export const server = setupServer(...handlers);

// Start server before all tests
beforeAll(() => {
  server.listen({ onUnhandledRequest: 'error' });
});

// Reset handlers after each test
afterEach(() => {
  server.resetHandlers();
  cleanup();
});

// Close server after all tests
afterAll(() => {
  server.close();
});

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock ResizeObserver
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

// Mock IntersectionObserver
global.IntersectionObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

// Mock scrollTo
Object.defineProperty(window, 'scrollTo', {
  value: vi.fn(),
  writable: true,
});

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

// Mock sessionStorage
const sessionStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
Object.defineProperty(window, 'sessionStorage', {
  value: sessionStorageMock,
});

// Mock crypto for security utilities
Object.defineProperty(global, 'crypto', {
  value: {
    getRandomValues: vi.fn((arr: Uint8Array | Uint16Array | Uint32Array) => {
      for (let i = 0; i < arr.length; i++) {
        arr[i] = Math.floor(Math.random() * 256);
      }
      return arr;
    }),
  },
});

// Mock performance.now
Object.defineProperty(global.performance, 'now', {
  value: vi.fn(() => Date.now()),
});

// Mock URL constructor
global.URL = URL;

// Reset all mocks after each test
afterEach(() => {
  vi.clearAllMocks();
  localStorageMock.clear();
  sessionStorageMock.clear();
});
