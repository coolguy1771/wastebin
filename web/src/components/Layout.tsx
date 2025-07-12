import React, { ReactNode } from 'react';
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import { useTheme } from '@mui/material/styles';
import Header from './header';
import Footer from './footer';
import { useResponsive, componentStyles } from '../theme/responsive';

interface LayoutProps {
  children: ReactNode;
  maxWidth?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | false;
  disableGutters?: boolean;
}

/**
 * Responsive layout component with proper spacing and mobile optimization
 */
export const Layout: React.FC<LayoutProps> = ({
  children,
  maxWidth = 'lg',
  disableGutters = false,
}) => {
  const theme = useTheme();
  const { isMobile, isTablet } = useResponsive();

  // Responsive padding values
  const getContentPadding = () => {
    if (isMobile) return componentStyles.content.padding.mobile;
    if (isTablet) return componentStyles.content.padding.tablet;
    return componentStyles.content.padding.desktop;
  };

  const contentPadding = getContentPadding();

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        minHeight: '100vh',
        backgroundColor: theme.palette.background.default,
      }}
    >
      {/* Header */}
      <Header />

      {/* Main Content */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          display: 'flex',
          flexDirection: 'column',
          width: '100%',
          overflow: 'auto',
        }}
      >
        {maxWidth === false ? (
          // Full width layout without container
          <Box
            sx={{
              flexGrow: 1,
              display: 'flex',
              flexDirection: 'column',
              px: !disableGutters ? contentPadding.x : 0,
              py: contentPadding.y,
            }}
          >
            {children}
          </Box>
        ) : (
          // Container-based layout
          <Container
            maxWidth={maxWidth}
            disableGutters={disableGutters}
            sx={{
              flexGrow: 1,
              display: 'flex',
              flexDirection: 'column',
              px: !disableGutters ? contentPadding.x : 0,
              py: contentPadding.y,
            }}
          >
            {children}
          </Container>
        )}
      </Box>

      {/* Footer */}
      <Footer />
    </Box>
  );
};

export default Layout;