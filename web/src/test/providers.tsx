import React, { ReactNode } from 'react';
import { BrowserRouter } from 'react-router-dom';
import { ThemeContextProvider } from '@contexts/ThemeContext';
import { SecurityProvider } from '@contexts/SecurityContext';
import { ErrorBoundary } from '@components/ErrorBoundary';

// All providers wrapper
export const AllProviders: React.FC<{
  children: ReactNode;
  initialEntries?: string[];
  withErrorBoundary?: boolean;
}> = ({ children, withErrorBoundary = true }) => {
  const content = withErrorBoundary ? <ErrorBoundary>{children}</ErrorBoundary> : children;

  return (
    <BrowserRouter>
      <SecurityProvider>
        <ThemeContextProvider>{content}</ThemeContextProvider>
      </SecurityProvider>
    </BrowserRouter>
  );
}; 