import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import CircularProgress from '@mui/material/CircularProgress';

const RawPastePage = () => {
  const { id } = useParams(); // Get the paste ID from the URL
  const [pasteContent, setPasteContent] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    // Fetch the raw paste content from the backend
    const fetchRawPaste = async () => {
      try {
        const response = await fetch(`/api/v1/paste/${id}/raw`);
        if (!response.ok) {
          throw new Error('Failed to fetch the raw paste content');
        }
        const text = await response.text();
        setPasteContent(text);
      } catch (err) {
        if (err instanceof Error) {
          setError(err.message);
        } else {
          setError('An unknown error occurred');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchRawPaste();
  }, [id]);

  if (loading) {
    return (
      <Box sx={{ textAlign: 'center', mt: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box sx={{ textAlign: 'center', mt: 4 }}>
        <Typography variant="h6" color="error">
          {error}
        </Typography>
      </Box>
    );
  }

  return (
    <Box
      sx={{
        my: 4,
        p: 2,
        border: '1px solid #ddd',
        borderRadius: '4px',
        backgroundColor: '#f4f4f4',
      }}
    >
      <Typography
        variant="body1"
        component="pre"
        sx={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}
      >
        {pasteContent}
      </Typography>
    </Box>
  );
};

export default RawPastePage;
