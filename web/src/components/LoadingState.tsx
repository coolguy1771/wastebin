import React from 'react';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import LinearProgress from '@mui/material/LinearProgress';
import Typography from '@mui/material/Typography';
import Skeleton from '@mui/material/Skeleton';

interface LoadingStateProps {
  variant?: 'circular' | 'linear' | 'skeleton';
  size?: 'small' | 'medium' | 'large';
  message?: string;
  overlay?: boolean;
  rows?: number; // For skeleton variant
  height?: number; // For skeleton variant
}

/**
 * Flexible loading state component with multiple variants
 */
export const LoadingState: React.FC<LoadingStateProps> = ({
  variant = 'circular',
  size = 'medium',
  message,
  overlay = false,
  rows = 3,
  height = 40,
}) => {
  const getSizeProps = () => {
    const sizes = {
      small: { size: 24, fontSize: '0.875rem' },
      medium: { size: 40, fontSize: '1rem' },
      large: { size: 60, fontSize: '1.125rem' },
    };
    return sizes[size];
  };

  const renderContent = () => {
    const { size: circularSize, fontSize } = getSizeProps();

    switch (variant) {
      case 'linear':
        return (
          <Box sx={{ width: '100%' }}>
            <LinearProgress />
            {message && (
              <Typography
                variant="body2"
                color="text.secondary"
                sx={{ mt: 1, textAlign: 'center', fontSize }}
              >
                {message}
              </Typography>
            )}
          </Box>
        );

      case 'skeleton':
        return (
          <Box sx={{ width: '100%' }}>
            {Array.from({ length: rows }).map((_, index) => (
              <Skeleton
                key={index}
                variant="rectangular"
                height={height}
                sx={{ mb: 1, borderRadius: 1 }}
              />
            ))}
          </Box>
        );

      case 'circular':
      default:
        return (
          <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
            <CircularProgress size={circularSize} />
            {message && (
              <Typography
                variant="body2"
                color="text.secondary"
                sx={{ mt: 2, textAlign: 'center', fontSize }}
              >
                {message}
              </Typography>
            )}
          </Box>
        );
    }
  };

  if (overlay) {
    return (
      <Box
        sx={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: 'rgba(255, 255, 255, 0.8)',
          zIndex: 1000,
        }}
      >
        {renderContent()}
      </Box>
    );
  }

  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        p: 3,
      }}
    >
      {renderContent()}
    </Box>
  );
};

export default LoadingState;
