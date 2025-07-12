import { useTheme, useMediaQuery } from '@mui/material';
import { Breakpoint, Theme } from '@mui/material/styles';

/**
 * Responsive design utilities and breakpoints
 */

// Custom breakpoints (in addition to Material-UI defaults)
export const customBreakpoints = {
  xs: 0,
  sm: 600,
  md: 900,
  lg: 1200,
  xl: 1536,
  mobile: 480,
  tablet: 768,
  desktop: 1024,
  wide: 1440,
} as const;

/**
 * Provides boolean flags for common responsive breakpoints and exposes the current Material-UI theme.
 *
 * Returns an object with flags indicating if the viewport matches mobile, tablet, desktop, large, extra-large, small mobile (custom), or wide screen (custom) breakpoints, along with the theme object for further queries.
 *
 * @returns An object containing responsive breakpoint flags and the current theme.
 */
export function useResponsive() {
  const theme = useTheme();

  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const isTablet = useMediaQuery(theme.breakpoints.between('sm', 'md'));
  const isDesktop = useMediaQuery(theme.breakpoints.up('md'));
  const isLarge = useMediaQuery(theme.breakpoints.up('lg'));
  const isXLarge = useMediaQuery(theme.breakpoints.up('xl'));

  // Custom breakpoints
  const isSmallMobile = useMediaQuery(`(max-width: ${customBreakpoints.mobile}px)`);
  const isWideScreen = useMediaQuery(`(min-width: ${customBreakpoints.wide}px)`);

  return {
    isMobile,
    isTablet,
    isDesktop,
    isLarge,
    isXLarge,
    isSmallMobile,
    isWideScreen,
    theme, // Expose theme for external breakpoint queries
  };
}

// Container sizes for different breakpoints
export const containerSizes = {
  xs: '100%',
  sm: '540px',
  md: '720px',
  lg: '960px',
  xl: '1140px',
} as const;

// Spacing utilities for responsive design
export const spacing = {
  mobile: {
    xs: 1,
    sm: 2,
    md: 3,
    lg: 4,
    xl: 5,
  },
  tablet: {
    xs: 2,
    sm: 3,
    md: 4,
    lg: 5,
    xl: 6,
  },
  desktop: {
    xs: 3,
    sm: 4,
    md: 5,
    lg: 6,
    xl: 8,
  },
} as const;

// Typography scales for different devices
export const typography = {
  mobile: {
    h1: { fontSize: '1.75rem', lineHeight: 1.3 },
    h2: { fontSize: '1.5rem', lineHeight: 1.3 },
    h3: { fontSize: '1.25rem', lineHeight: 1.4 },
    h4: { fontSize: '1.125rem', lineHeight: 1.4 },
    h5: { fontSize: '1rem', lineHeight: 1.5 },
    h6: { fontSize: '0.875rem', lineHeight: 1.5 },
    body1: { fontSize: '0.875rem', lineHeight: 1.6 },
    body2: { fontSize: '0.75rem', lineHeight: 1.6 },
  },
  tablet: {
    h1: { fontSize: '2rem', lineHeight: 1.3 },
    h2: { fontSize: '1.75rem', lineHeight: 1.3 },
    h3: { fontSize: '1.5rem', lineHeight: 1.4 },
    h4: { fontSize: '1.25rem', lineHeight: 1.4 },
    h5: { fontSize: '1.125rem', lineHeight: 1.5 },
    h6: { fontSize: '1rem', lineHeight: 1.5 },
    body1: { fontSize: '1rem', lineHeight: 1.6 },
    body2: { fontSize: '0.875rem', lineHeight: 1.6 },
  },
  desktop: {
    h1: { fontSize: '2.25rem', lineHeight: 1.2 },
    h2: { fontSize: '2rem', lineHeight: 1.3 },
    h3: { fontSize: '1.75rem', lineHeight: 1.3 },
    h4: { fontSize: '1.5rem', lineHeight: 1.4 },
    h5: { fontSize: '1.25rem', lineHeight: 1.4 },
    h6: { fontSize: '1.125rem', lineHeight: 1.5 },
    body1: { fontSize: '1rem', lineHeight: 1.6 },
    body2: { fontSize: '0.875rem', lineHeight: 1.6 },
  },
} as const;

// Component-specific responsive styles
export const componentStyles = {
  header: {
    height: {
      mobile: '56px',
      desktop: '64px',
    },
    padding: {
      mobile: { x: 2, y: 1 },
      desktop: { x: 3, y: 1.5 },
    },
  },

  navigation: {
    width: {
      mobile: '100%',
      tablet: '240px',
      desktop: '280px',
    },
  },

  content: {
    padding: {
      mobile: { x: 2, y: 2 },
      tablet: { x: 3, y: 3 },
      desktop: { x: 4, y: 4 },
    },
    maxWidth: {
      mobile: '100%',
      tablet: '100%',
      desktop: '1200px',
    },
  },

  form: {
    spacing: {
      mobile: 2,
      tablet: 3,
      desktop: 4,
    },
    buttonSize: {
      mobile: 'medium',
      desktop: 'large',
    } as const,
  },

  card: {
    padding: {
      mobile: 2,
      tablet: 3,
      desktop: 4,
    },
    borderRadius: {
      mobile: 1,
      desktop: 2,
    },
  },
} as const;

// Grid system utilities
export const grid = {
  columns: 12,
  gutter: {
    mobile: 2,
    tablet: 3,
    desktop: 4,
  },
} as const;

// Utility functions for breakpoint queries (use with theme from useResponsive hook)
export const createBreakpointHelpers = (theme: Theme) => ({
  only: (breakpoint: Breakpoint) => theme.breakpoints.only(breakpoint),
  up: (breakpoint: Breakpoint) => theme.breakpoints.up(breakpoint),
  down: (breakpoint: Breakpoint) => theme.breakpoints.down(breakpoint),
  between: (start: Breakpoint, end: Breakpoint) => theme.breakpoints.between(start, end),
});

/**
 * Returns an object mapping responsive breakpoint keys to the provided values for mobile, tablet, and desktop.
 *
 * @param mobileValue - Value to use for the mobile (`xs`) breakpoint
 * @param tabletValue - Optional value for the tablet (`sm`) breakpoint
 * @param desktopValue - Optional value for the desktop (`md`) breakpoint
 * @returns An object with keys `xs`, and optionally `sm` and `md`, assigned to the corresponding values
 */
export function getResponsiveValue<T>(
  mobileValue: T,
  tabletValue?: T,
  desktopValue?: T
): { xs: T; sm?: T; md?: T } {
  return {
    xs: mobileValue,
    ...(tabletValue && { sm: tabletValue }),
    ...(desktopValue && { md: desktopValue }),
  };
}
