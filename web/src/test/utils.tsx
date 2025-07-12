/**
 * Test utilities and custom render functions
 */

import React, { ReactElement, ReactNode } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import { ThemeContextProvider } from '@contexts/ThemeContext';
import { SecurityProvider } from '@contexts/SecurityContext';
import { ErrorBoundary } from '@components/ErrorBoundary';

// Custom render with providers
interface CustomRenderOptions extends Omit<RenderOptions, 'wrapper'> {
  initialEntries?: string[];
  withErrorBoundary?: boolean;
  theme?: 'light' | 'dark';
}

// All providers wrapper
const AllProviders: React.FC<{
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

// Custom render function
export const renderWithProviders = (ui: ReactElement, options: CustomRenderOptions = {}) => {
  const { initialEntries, withErrorBoundary, theme, ...renderOptions } = options;

  // Set theme in localStorage if specified
  if (theme) {
    localStorage.setItem('wastebin-theme-mode', theme);
  }

  return render(ui, {
    wrapper: ({ children }) => (
      <AllProviders initialEntries={initialEntries} withErrorBoundary={withErrorBoundary}>
        {children}
      </AllProviders>
    ),
    ...renderOptions,
  });
};

// Form testing utilities
export const fillForm = async (user: ReturnType<typeof userEvent.setup>, fields: Record<string, string | boolean>) => {
  for (const [name, value] of Object.entries(fields)) {
    const field = document.querySelector(`[name="${name}"]`) as HTMLElement;

    if (!field) {
      throw new Error(`Field with name "${name}" not found`);
    }

    if (field.getAttribute('type') === 'checkbox') {
      if (value) {
        await user.click(field);
      }
    } else {
      await user.clear(field);
      if (typeof value === 'string') {
        await user.type(field, value);
      }
    }
  }
};

// Wait for loading to complete
export const waitForLoadingToFinish = async () => {
  const { queryByText } = await import('@testing-library/react');

  // Wait for common loading indicators to disappear
  const loadingIndicators = ['Loading...', 'Please wait...', 'Submitting...'];

  for (const indicator of loadingIndicators) {
    const element = queryByText(document.body, indicator);
    if (element) {
      await new Promise(resolve => {
        const observer = new MutationObserver(() => {
          if (!document.body.contains(element)) {
            observer.disconnect();
            resolve(undefined);
          }
        });
        observer.observe(document.body, { childList: true, subtree: true });
      });
    }
  }
};

// Mock functions
export const createMockPaste = (overrides = {}) => ({
  uuid: '123e4567-e89b-12d3-a456-426614174000',
  content: 'Test content',
  language: 'javascript',
  burn: false,
  expiry_timestamp: new Date(Date.now() + 86400000).toISOString(),
  created_at: new Date().toISOString(),
  ...overrides,
});

// API mock helpers
export const mockApiSuccess = (data: unknown) => {
  return Promise.resolve({
    ok: true,
    status: 200,
    json: () => Promise.resolve(data),
    text: () => Promise.resolve(JSON.stringify(data)),
  });
};

export const mockApiError = (message = 'Server error') => {
  return Promise.reject(new Error(message));
};

// Test data generators
export const generateLargeContent = (size: number): string => {
  return 'A'.repeat(size);
};

export const generateUUID = (): string => {
  return '123e4567-e89b-12d3-a456-426614174000';
};

// Custom matchers for testing
export const customMatchers = {
  toBeValidUUID: (received: string) => {
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    const pass = uuidRegex.test(received);

    return {
      pass,
      message: () => `expected ${received} ${pass ? 'not ' : ''}to be a valid UUID`,
    };
  },

  toHaveValidTimestamp: (received: string) => {
    const date = new Date(received);
    const pass = !isNaN(date.getTime());

    return {
      pass,
      message: () => `expected ${received} ${pass ? 'not ' : ''}to be a valid timestamp`,
    };
  },
};

// Accessibility testing helpers (placeholder for now)
export const checkA11y = async (container: HTMLElement) => {
  // TODO: Add axe-core when needed
  // const { axe } = await import('@axe-core/react');
  // const results = await axe(container);
  // if (results.violations.length > 0) {
  //   throw new Error(`Accessibility violations: ${results.violations.map(v => v.description).join(', ')}`);
  // }
  console.log('A11y check placeholder for container:', container.tagName);
};

// Performance testing helpers
export const measureRenderTime = async (renderFn: () => void): Promise<number> => {
  const start = performance.now();
  renderFn();
  await waitForLoadingToFinish();
  const end = performance.now();
  return end - start;
};

// Error boundary testing
export const triggerErrorBoundary = () => {
  const ThrowError = () => {
    throw new Error('Test error for error boundary');
  };
  return <ThrowError />;
};

// Export everything for convenience
export * from '@testing-library/react';
export * from '@testing-library/user-event';
export { renderWithProviders as render };
