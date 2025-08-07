import { ReactElement } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { AllProviders } from './providers';

// Custom render with providers
interface CustomRenderOptions extends Omit<RenderOptions, 'wrapper'> {
  initialEntries?: string[];
  withErrorBoundary?: boolean;
  theme?: 'light' | 'dark';
}

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

// Export testing utilities
export { screen, waitFor, fireEvent, act } from '@testing-library/react';

// Error boundary testing
export const triggerErrorBoundary = () => {
  const ThrowError = () => {
    throw new Error('Test error for error boundary');
  };
  return <ThrowError />;
}; 