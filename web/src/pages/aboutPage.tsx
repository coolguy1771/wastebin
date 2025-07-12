import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';

const AboutPage = () => {
  return (
    <Container maxWidth="md">
      <Box sx={{ my: 4 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          About Wastebin
        </Typography>
        <Typography variant="body1">
          Wastebin is designed to provide a simple and secure way to share code snippets and text online. Built with modern web technologies to ensure a smooth experience.
        </Typography>
      </Box>
    </Container>
  );
};

export default AboutPage;
