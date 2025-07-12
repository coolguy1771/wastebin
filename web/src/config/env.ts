/**
 * Environment configuration with validation and type safety
 */

export interface AppConfig {
  apiBaseUrl: string;
  apiTimeout: number;
  enableDarkMode: boolean;
  enableAnalytics: boolean;
  debugMode: boolean;
  cspEnabled: boolean;
  isDevelopment: boolean;
  isProduction: boolean;
}

// Environment variable validation
const getEnvVar = (key: string, defaultValue?: string): string => {
  const value = import.meta.env[key];
  if (!value && !defaultValue) {
    throw new Error(`Missing required environment variable: ${key}`);
  }
  return value || defaultValue!;
};

const getBooleanEnvVar = (key: string, defaultValue: boolean = false): boolean => {
  const value = import.meta.env[key];
  if (!value) return defaultValue;
  return value.toLowerCase() === 'true';
};

const getNumberEnvVar = (key: string, defaultValue: number): number => {
  const value = import.meta.env[key];
  if (!value) return defaultValue;
  const parsed = parseInt(value, 10);
  if (isNaN(parsed)) {
    console.warn(`Invalid number for ${key}, using default: ${defaultValue}`);
    return defaultValue;
  }
  return parsed;
};

// Create configuration object
export const config: AppConfig = {
  apiBaseUrl: getEnvVar('VITE_API_BASE_URL', ''),
  apiTimeout: getNumberEnvVar('VITE_API_TIMEOUT', 10000),
  enableDarkMode: getBooleanEnvVar('VITE_ENABLE_DARK_MODE', true),
  enableAnalytics: getBooleanEnvVar('VITE_ENABLE_ANALYTICS', false),
  debugMode: getBooleanEnvVar('VITE_DEBUG_MODE', false),
  cspEnabled: getBooleanEnvVar('VITE_CSP_ENABLED', true),
  isDevelopment: import.meta.env.DEV,
  isProduction: import.meta.env.PROD,
};

// Validation
if (config.isDevelopment && config.debugMode) {
  console.log('App Configuration:', config);
}

// Export individual configs for convenience
export const {
  apiBaseUrl,
  apiTimeout,
  enableDarkMode,
  enableAnalytics,
  debugMode,
  isDevelopment,
  isProduction,
} = config;
