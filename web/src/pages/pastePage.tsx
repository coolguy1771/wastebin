import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import CardActions from '@mui/material/CardActions';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import { Light as SyntaxHighlighter } from 'react-syntax-highlighter';
// Use simpler style approach for build compatibility
const lightStyle = {
  hljs: {
    display: 'block' as const,
    overflowX: 'auto' as const,
    padding: '0.5em',
    background: '#f5f5f5',
    color: '#333',
  },
};

const darkStyle = {
  hljs: {
    display: 'block' as const,
    overflowX: 'auto' as const,
    padding: '0.5em',
    background: '#2b2b2b',
    color: '#f8f8f8',
  },
};
import { useApi } from '../hooks/useApi';
import { pasteAPI, type PasteDetails } from '../services/api';
import { LoadingState } from '../components/LoadingState';
import { ErrorDisplay } from '../components/ErrorDisplay';
import { useThemeMode } from '../contexts/ThemeContext';
import { useResponsive } from '../theme/responsive';
import { getLanguageDisplayName } from '../utils/validation';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import DownloadIcon from '@mui/icons-material/Download';
import ShareIcon from '@mui/icons-material/Share';
import CodeIcon from '@mui/icons-material/Code';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import SecurityIcon from '@mui/icons-material/Security';
const PastePage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { mode } = useThemeMode();
  const { isMobile } = useResponsive();
  const { data: paste, loading, error, retry, execute } = useApi<PasteDetails>();
  const [copySuccess, setCopySuccess] = useState(false);

  useEffect(() => {
    if (id) {
      execute(() => pasteAPI.get(id));
    }
  }, [id, execute]);

  if (loading) {
    return <LoadingState message="Loading paste..." />;
  }

  if (error) {
    return <ErrorDisplay error={error} onRetry={retry} title="Failed to load paste" />;
  }

  if (!paste) {
    return (
      <ErrorDisplay error="Paste not found" title="Not Found" severity="info" showRetry={false} />
    );
  }

  // Format timestamps
  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleString();
  };

  const isExpired = paste.expiry_timestamp && new Date(paste.expiry_timestamp) < new Date();

  // Theme options for syntax highlighter
  const syntaxTheme = mode === 'dark' ? darkStyle : lightStyle;

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(paste.content);
      setCopySuccess(true);
    } catch (err) {
      console.error('Failed to copy text: ', err);
    }
  };

  const handleDownload = () => {
    const blob = new Blob([paste.content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `paste-${id}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const handleShare = async () => {
    const url = window.location.href;
    if (navigator.share) {
      try {
        await navigator.share({
          title: 'Wastebin Paste',
          url: url,
        });
      } catch (err) {
        // Fallback to clipboard
        await navigator.clipboard.writeText(url);
        setCopySuccess(true);
      }
    } else {
      await navigator.clipboard.writeText(url);
      setCopySuccess(true);
    }
  };

  return (
    <>
      <Box
        sx={{
          height: '100%',
          display: 'flex',
          flexDirection: { xs: 'column', lg: 'row' },
          gap: 3,
          p: { xs: 2, md: 3, lg: 4 },
        }}
      >
        {/* Main Content Area */}
        <Box
          sx={{
            flex: 1,
            display: 'flex',
            flexDirection: 'column',
            minHeight: 0,
          }}
        >
          {/* Content Card */}
          <Card
            sx={{
              flex: 1,
              display: 'flex',
              flexDirection: 'column',
              minHeight: 0,
            }}
          >
            <CardHeader
              avatar={<CodeIcon />}
              title={
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap' }}>
                  <Typography variant="h6">Paste Content</Typography>
                  <Chip
                    label={getLanguageDisplayName(paste.language)}
                    size="small"
                    icon={<CodeIcon fontSize="small" />}
                  />
                  {paste.burn && (
                    <Chip
                      label="Burn after read"
                      color="warning"
                      size="small"
                      icon={<SecurityIcon fontSize="small" />}
                    />
                  )}
                  {isExpired && <Chip label="Expired" color="error" size="small" />}
                </Box>
              }
              subheader={
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap' }}>
                  <AccessTimeIcon fontSize="small" />
                  <Typography variant="body2" color="text.secondary">
                    Created: {formatTimestamp(paste.created_at)}
                    {paste.expiry_timestamp && (
                      <> • Expires: {formatTimestamp(paste.expiry_timestamp)}</>
                    )}
                  </Typography>
                </Box>
              }
              action={
                <Stack direction="row" spacing={1}>
                  <Tooltip title="Copy to clipboard">
                    <IconButton onClick={handleCopy} size="small">
                      <ContentCopyIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Download">
                    <IconButton onClick={handleDownload} size="small">
                      <DownloadIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Share">
                    <IconButton onClick={handleShare} size="small">
                      <ShareIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="View raw">
                    <IconButton
                      onClick={() => window.open(`/paste/${id}/raw`, '_blank')}
                      size="small"
                    >
                      <OpenInNewIcon />
                    </IconButton>
                  </Tooltip>
                </Stack>
              }
              sx={{ pb: 1 }}
            />
            <CardContent
              sx={{
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                p: 0,
                '&:last-child': { pb: 0 },
                overflow: 'hidden',
              }}
            >
              <Box
                sx={{
                  flex: 1,
                  overflow: 'auto',
                  '& pre': {
                    margin: 0,
                    fontSize: isMobile ? '12px' : '14px',
                    fontFamily: 'JetBrains Mono, Monaco, Consolas, monospace',
                    lineHeight: 1.6,
                  },
                  '& code': {
                    fontFamily: 'JetBrains Mono, Monaco, Consolas, monospace',
                  },
                }}
              >
                <SyntaxHighlighter
                  language={paste.language}
                  style={syntaxTheme}
                  showLineNumbers
                  customStyle={{
                    margin: 0,
                    borderRadius: 0,
                    minHeight: '100%',
                    background: 'transparent',
                  }}
                  lineNumberStyle={{
                    minWidth: '3em',
                    paddingRight: '1em',
                    textAlign: 'right',
                    userSelect: 'none',
                  }}
                >
                  {paste.content}
                </SyntaxHighlighter>
              </Box>
            </CardContent>
            <CardActions sx={{ p: 2, pt: 1, borderTop: 1, borderColor: 'divider' }}>
              <Button
                variant="outlined"
                startIcon={<OpenInNewIcon />}
                onClick={() => window.open(`/paste/${id}/raw`, '_blank')}
                size="small"
              >
                View Raw
              </Button>
              <Button
                variant="outlined"
                startIcon={<ContentCopyIcon />}
                onClick={handleCopy}
                size="small"
              >
                Copy Content
              </Button>
              <Button
                variant="outlined"
                startIcon={<DownloadIcon />}
                onClick={handleDownload}
                size="small"
              >
                Download
              </Button>
            </CardActions>
          </Card>
        </Box>

        {/* Sidebar with metadata */}
        <Paper
          sx={{
            width: { xs: '100%', lg: '300px' },
            p: 3,
            display: 'flex',
            flexDirection: 'column',
            gap: 3,
            height: 'fit-content',
          }}
        >
          <Box>
            <Typography
              variant="h6"
              gutterBottom
              sx={{ display: 'flex', alignItems: 'center', gap: 1 }}
            >
              <SecurityIcon fontSize="small" />
              Paste Information
            </Typography>

            <Stack spacing={2}>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  Language
                </Typography>
                <Typography variant="body1">{getLanguageDisplayName(paste.language)}</Typography>
              </Box>

              <Box>
                <Typography variant="body2" color="text.secondary">
                  Created
                </Typography>
                <Typography variant="body1">{formatTimestamp(paste.created_at)}</Typography>
              </Box>

              {paste.expiry_timestamp && (
                <Box>
                  <Typography variant="body2" color="text.secondary">
                    Expires
                  </Typography>
                  <Typography variant="body1" color={isExpired ? 'error.main' : 'text.primary'}>
                    {formatTimestamp(paste.expiry_timestamp)}
                  </Typography>
                </Box>
              )}

              <Box>
                <Typography variant="body2" color="text.secondary">
                  Size
                </Typography>
                <Typography variant="body1">
                  {paste.content.length.toLocaleString()} characters
                </Typography>
              </Box>

              <Box>
                <Typography variant="body2" color="text.secondary">
                  Lines
                </Typography>
                <Typography variant="body1">
                  {paste.content.split('\n').length.toLocaleString()}
                </Typography>
              </Box>
            </Stack>
          </Box>

          {/* Security Features */}
          <Box>
            <Typography
              variant="h6"
              gutterBottom
              sx={{ display: 'flex', alignItems: 'center', gap: 1 }}
            >
              <SecurityIcon fontSize="small" />
              Security
            </Typography>

            <Stack spacing={1}>
              <Chip
                label="End-to-end encrypted"
                size="small"
                icon={<SecurityIcon />}
                variant="outlined"
              />
              {paste.burn && (
                <Chip label="Burns after reading" size="small" color="warning" variant="outlined" />
              )}
            </Stack>
          </Box>
        </Paper>
      </Box>

      {/* Success Snackbar */}
      <Snackbar
        open={copySuccess}
        autoHideDuration={3000}
        onClose={() => setCopySuccess(false)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={() => setCopySuccess(false)} severity="success">
          Copied to clipboard!
        </Alert>
      </Snackbar>
    </>
  );
};

export default PastePage;
