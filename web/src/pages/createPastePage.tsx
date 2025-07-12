import React from 'react';
import { useNavigate } from 'react-router-dom';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import Alert from '@mui/material/Alert';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import Paper from '@mui/material/Paper';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import { useForm } from '../hooks/useForm';
import {
  pasteValidationRules,
  formatExpiryTime,
  getExpiryOptions,
  getLanguageDisplayName,
  PASTE_CONSTRAINTS,
  type PasteFormData,
} from '../utils/validation';
import { pasteAPI, getErrorMessage } from '../services/api';
import { LoadingState } from '../components/LoadingState';
import { useResponsive } from '../theme/responsive';
import SecurityIcon from '@mui/icons-material/Security';
import TimerIcon from '@mui/icons-material/Timer';
import CodeIcon from '@mui/icons-material/Code';
import FireIcon from '@mui/icons-material/Whatshot';

const CreatePastePage: React.FC = () => {
  const navigate = useNavigate();
  const { isMobile } = useResponsive();

  // Form handling with validation
  const { values, errors, isValid, isSubmitting, getFieldProps, getSelectProps, handleSubmit } =
    useForm<PasteFormData>({
      initialValues: {
        content: '',
        language: 'plaintext',
        expires: '0',
        burn: false,
      },
      validationRules: pasteValidationRules,
      onSubmit: async formData => {
        try {
          const response = await pasteAPI.create({
            content: formData.content,
            language: formData.language,
            expiry_time: formatExpiryTime(formData.expires),
            burn: formData.burn,
          });

          navigate(`/paste/${response.uuid}`);
        } catch (error) {
          throw new Error(getErrorMessage(error));
        }
      },
    });

  // Language options with display names
  const languageOptions = PASTE_CONSTRAINTS.SUPPORTED_LANGUAGES.map(lang => ({
    value: lang,
    label: getLanguageDisplayName(lang),
  }));

  const expiryOptions = getExpiryOptions();

  return (
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
          minHeight: 0, // Important for proper flex sizing
        }}
      >
        {/* Header */}
        <Box sx={{ mb: 3 }}>
          <Typography variant="h4" component="h1" gutterBottom sx={{ fontWeight: 700 }}>
            Create a New Paste
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Share code, text, or any content securely. All pastes are automatically encrypted.
          </Typography>
        </Box>

        {/* Main Editor Card */}
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
            title="Content"
            subheader="Enter your paste content below"
            sx={{ pb: 1 }}
          />
          <CardContent
            sx={{
              flex: 1,
              display: 'flex',
              flexDirection: 'column',
              pt: 0,
              '&:last-child': { pb: 2 },
            }}
          >
            <form
              onSubmit={handleSubmit}
              style={{
                display: 'flex',
                flexDirection: 'column',
                height: '100%',
              }}
            >
              <TextField
                placeholder="Enter your content here..."
                multiline
                variant="outlined"
                fullWidth
                {...getFieldProps('content')}
                InputProps={{
                  sx: {
                    fontFamily: 'JetBrains Mono, Monaco, Consolas, monospace',
                    fontSize: '14px',
                    lineHeight: 1.6,
                    '& .MuiInputBase-input': {
                      height: '100% !important',
                      overflow: 'auto !important',
                    },
                  },
                }}
                sx={{
                  flex: 1,
                  '& .MuiOutlinedInput-root': {
                    height: '100%',
                    alignItems: 'stretch',
                    '& textarea': {
                      height: '100% !important',
                      overflow: 'auto !important',
                      resize: 'none',
                      minHeight: isMobile ? '300px' : '400px',
                    },
                  },
                }}
              />
            </form>
          </CardContent>
        </Card>
      </Box>

      {/* Sidebar */}
      <Paper
        sx={{
          width: { xs: '100%', lg: '350px' },
          p: 3,
          display: 'flex',
          flexDirection: 'column',
          gap: 3,
        }}
      >
        {/* Info Alert */}
        <Alert
          severity="info"
          icon={<SecurityIcon />}
          sx={{
            '& .MuiAlert-message': {
              display: 'flex',
              flexDirection: 'column',
              gap: 1,
            },
          }}
        >
          <Typography variant="subtitle2">Secure & Private</Typography>
          <Typography variant="body2">
            Maximum size: {Math.floor(PASTE_CONSTRAINTS.MAX_CONTENT_LENGTH / 1024 / 1024)}MB
          </Typography>
        </Alert>

        {/* Settings */}
        <Stack spacing={3}>
          <Box>
            <Typography
              variant="h6"
              gutterBottom
              sx={{ display: 'flex', alignItems: 'center', gap: 1 }}
            >
              <CodeIcon fontSize="small" />
              Language
            </Typography>
            <FormControl fullWidth size="small">
              <Select {...getSelectProps('language')}>
                {languageOptions.map(option => (
                  <MenuItem key={option.value} value={option.value}>
                    <Stack direction="row" alignItems="center" spacing={1}>
                      <CodeIcon fontSize="small" />
                      <span>{option.label}</span>
                    </Stack>
                  </MenuItem>
                ))}
              </Select>
              {errors.language && (
                <Typography variant="caption" color="error" sx={{ mt: 0.5 }}>
                  {errors.language}
                </Typography>
              )}
            </FormControl>
          </Box>

          <Box>
            <Typography
              variant="h6"
              gutterBottom
              sx={{ display: 'flex', alignItems: 'center', gap: 1 }}
            >
              <TimerIcon fontSize="small" />
              Expiration
            </Typography>
            <FormControl fullWidth size="small">
              <Select {...getSelectProps('expires')}>
                {expiryOptions.map(option => (
                  <MenuItem key={option.value} value={option.value}>
                    <Stack direction="row" alignItems="center" spacing={1}>
                      <TimerIcon fontSize="small" />
                      <span>{option.label}</span>
                    </Stack>
                  </MenuItem>
                ))}
              </Select>
              {errors.expires && (
                <Typography variant="caption" color="error" sx={{ mt: 0.5 }}>
                  {errors.expires}
                </Typography>
              )}
            </FormControl>
          </Box>

          <Box>
            <Typography
              variant="h6"
              gutterBottom
              sx={{ display: 'flex', alignItems: 'center', gap: 1 }}
            >
              <FireIcon fontSize="small" />
              Security Options
            </Typography>
            <FormControlLabel
              control={
                <Checkbox
                  checked={values.burn}
                  onChange={e =>
                    getFieldProps('burn').onChange({ target: { value: e.target.checked } } as any)
                  }
                />
              }
              label={
                <Box>
                  <Typography variant="body2">Burn after reading</Typography>
                  <Typography variant="caption" color="text.secondary">
                    Delete after first view
                  </Typography>
                </Box>
              }
            />
          </Box>
        </Stack>

        {/* Action Buttons */}
        <Box sx={{ mt: 'auto', pt: 2 }}>
          <Button
            variant="contained"
            color="primary"
            type="submit"
            onClick={handleSubmit}
            disabled={!isValid || isSubmitting}
            size="large"
            fullWidth
            sx={{
              py: 1.5,
              fontSize: '1rem',
              fontWeight: 600,
            }}
          >
            {isSubmitting ? <LoadingState variant="circular" size="small" /> : 'Create Paste'}
          </Button>

          {!isValid && Object.keys(errors).length > 0 && (
            <Alert severity="error" sx={{ mt: 2 }}>
              <Typography variant="body2">Please fix the errors above</Typography>
            </Alert>
          )}

          {/* Feature Tags */}
          <Stack direction="row" spacing={1} sx={{ mt: 2, flexWrap: 'wrap' }}>
            <Chip label="Encrypted" size="small" icon={<SecurityIcon />} variant="outlined" />
            <Chip label="Syntax Highlighting" size="small" icon={<CodeIcon />} variant="outlined" />
          </Stack>
        </Box>
      </Paper>
    </Box>
  );
};

export default CreatePastePage;
