import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { config } from '../config/env';
import { generateCSPHeader, isSecureContext, rateLimiter } from '../utils/security';

interface SecurityContextType {
  isSecure: boolean;
  cspEnabled: boolean;
  checkRateLimit: (key: string) => boolean;
  resetRateLimit: (key: string) => void;
}

const SecurityContext = createContext<SecurityContextType | undefined>(undefined);

interface SecurityProviderProps {
  children: ReactNode;
}

/**
 * Security context provider for managing security features
 */
export const SecurityProvider: React.FC<SecurityProviderProps> = ({ children }) => {
  const [isSecure, setIsSecure] = useState<boolean>(false);

  useEffect(() => {
    // Check if we're in a secure context
    setIsSecure(isSecureContext());

    // Set up CSP if enabled
    if (config.cspEnabled) {
      const cspHeader = generateCSPHeader();
      
      // Add CSP meta tag if not already present
      const existingCSP = document.querySelector('meta[http-equiv="Content-Security-Policy"]');
      if (!existingCSP) {
        const meta = document.createElement('meta');
        meta.httpEquiv = 'Content-Security-Policy';
        meta.content = cspHeader;
        document.head.appendChild(meta);
      }
    }

    // Log security warnings in development
    if (config.isDevelopment) {
      if (!isSecureContext()) {
        console.warn('⚠️ Not running in secure context (HTTPS). Some features may be limited.');
      }
      
      if (config.debugMode) {
        console.log('🔒 Security Context initialized:', {
          isSecure: isSecureContext(),
          cspEnabled: config.cspEnabled,
        });
      }
    }
  }, []);

  const checkRateLimit = (key: string): boolean => {
    return rateLimiter.isAllowed(key);
  };

  const resetRateLimit = (key: string): void => {
    rateLimiter.reset(key);
  };

  const contextValue: SecurityContextType = {
    isSecure,
    cspEnabled: config.cspEnabled,
    checkRateLimit,
    resetRateLimit,
  };

  return (
    <SecurityContext.Provider value={contextValue}>
      {children}
    </SecurityContext.Provider>
  );
};

/**
 * Hook to use security context
 */
export const useSecurity = (): SecurityContextType => {
  const context = useContext(SecurityContext);
  if (!context) {
    throw new Error('useSecurity must be used within a SecurityProvider');
  }
  return context;
};

export default SecurityProvider;