import React, { ReactNode } from 'react';
import { BrowserRouter, MemoryRouter } from 'react-router-dom';
import { ThemeContextProvider } from '@contexts/ThemeContext';
import { SecurityProvider } from '@contexts/SecurityContext';
import { ErrorBoundary } from '@components/ErrorBoundary';

// All providers wrapper
export const AllProviders: React.FC<{
  children: ReactNode;
  initialEntries?: string[];
  withErrorBoundary?: boolean;
}> = ({ children, initialEntries, withErrorBoundary = true }) => {
  const content = withErrorBoundary ? <ErrorBoundary>{children}</ErrorBoundary> : children;

  const Router = initialEntries ? MemoryRouter : BrowserRouter;
  const routerProps = initialEntries ? { initialEntries } : {};

  return (
    <Router {...routerProps}>
      <SecurityProvider>
        <ThemeContextProvider>{content}</ThemeContextProvider>
      </SecurityProvider>
    </Router>
  );
};
