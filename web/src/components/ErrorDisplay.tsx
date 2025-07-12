import React from 'react';
import Box from '@mui/material/Box';
import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import RefreshIcon from '@mui/icons-material/Refresh';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';
import InfoIcon from '@mui/icons-material/Info';

interface ErrorDisplayProps {
  error: string;
  severity?: 'error' | 'warning' | 'info';
  title?: string;
  onRetry?: () => void;
  retryLabel?: string;
  showRetry?: boolean;
  variant?: 'standard' | 'filled' | 'outlined';
  sx?: object;
}

/**
 * Consistent error display component with retry functionality
 */
export const ErrorDisplay: React.FC<ErrorDisplayProps> = ({
  error,
  severity = 'error',
  title,
  onRetry,
  retryLabel = 'Try Again',
  showRetry = true,
  variant = 'standard',
  sx,
}) => {
  const getIcon = () => {
    switch (severity) {
      case 'warning':
        return <WarningAmberIcon />;
      case 'info':
        return <InfoIcon />;
      case 'error':
      default:
        return <ErrorOutlineIcon />;
    }
  };

  const getDefaultTitle = () => {
    switch (severity) {
      case 'warning':
        return 'Warning';
      case 'info':
        return 'Information';
      case 'error':
      default:
        return 'Error';
    }
  };

  return (
    <Box sx={{ width: '100%', ...sx }}>
      <Alert
        severity={severity}
        variant={variant}
        icon={getIcon()}
        sx={{
          alignItems: 'flex-start',
          '& .MuiAlert-message': {
            width: '100%',
          },
        }}
      >
        {(title || getDefaultTitle()) && <AlertTitle>{title || getDefaultTitle()}</AlertTitle>}

        <Typography variant="body2" sx={{ mb: showRetry && onRetry ? 2 : 0 }}>
          {error}
        </Typography>

        {showRetry && onRetry && (
          <Button
            variant="outlined"
            size="small"
            startIcon={<RefreshIcon />}
            onClick={onRetry}
            sx={{
              mt: 1,
              borderColor: 'currentColor',
              color: 'inherit',
              '&:hover': {
                borderColor: 'currentColor',
                backgroundColor: 'rgba(0, 0, 0, 0.04)',
              },
            }}
          >
            {retryLabel}
          </Button>
        )}
      </Alert>
    </Box>
  );
};

// Specialized error components for common scenarios
export const NetworkErrorDisplay: React.FC<{ onRetry?: () => void }> = ({ onRetry }) => (
  <ErrorDisplay
    error="Unable to connect to the server. Please check your internet connection and try again."
    title="Connection Error"
    onRetry={onRetry}
    severity="warning"
  />
);

export const NotFoundErrorDisplay: React.FC<{ resourceName?: string }> = ({
  resourceName = 'resource',
}) => (
  <ErrorDisplay
    error={`The requested ${resourceName} could not be found. It may have been deleted or the link may be incorrect.`}
    title="Not Found"
    severity="info"
    showRetry={false}
  />
);

export const ValidationErrorDisplay: React.FC<{
  errors: string[];
  onDismiss?: () => void;
}> = ({ errors, onDismiss: _onDismiss }) => (
  <ErrorDisplay
    error={errors.join(', ')}
    title="Please correct the following issues:"
    severity="warning"
    showRetry={false}
  />
);

export default ErrorDisplay;
