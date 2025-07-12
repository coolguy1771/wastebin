import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import { Link } from 'react-router-dom';

const HomePage = () => {
  const samplePasteId = '12345'; // Replace with actual paste ID when available

  return (
    <Container maxWidth="md">
      <Box sx={{ my: 4 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Welcome to Wastebin
        </Typography>
        <Typography variant="body1" gutterBottom>
          Wastebin is a secure and easy-to-use pastebin service for sharing text and code snippets.
        </Typography>
        {/* Link to create a new paste */}
        <Button
          component={Link}
          to={`/paste/new`}
          variant="contained"
          color="primary"
          sx={{ mr: 2 }}
        >
          Create New Paste
        </Button>
        {/* Link to view a regular paste */}
        <Button
          component={Link}
          to={`/paste/${samplePasteId}`}
          variant="outlined"
          color="secondary"
          sx={{ mr: 2 }}
        >
          View Paste
        </Button>
        {/* Link to view the raw version of a paste */}
        <Button
          component={Link}
          to={`/paste/${samplePasteId}/raw`}
          variant="outlined"
          color="secondary"
        >
          View Raw Paste
        </Button>
      </Box>
    </Container>
  );
};

export default HomePage;
