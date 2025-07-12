import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';

const Footer = () => {
  return (
    <Box component="footer" sx={{ py: 3, textAlign: 'center', backgroundColor: '#1e1e1e', color: '#fff' }}>
      <Typography variant="body2">
        &copy; {new Date().getFullYear()} Wastebin. All rights reserved.
      </Typography>
    </Box>
  );
};

export default Footer;
