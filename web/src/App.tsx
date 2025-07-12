import { BrowserRouter as Router, Routes, Route, useLocation } from 'react-router-dom';
import HomePage from './pages/homePage';
import AboutPage from './pages/aboutPage';
import PastePage from './pages/pastePage';
import RawPastePage from './pages/rawPastePage';
import CreatePastePage from './pages/createPastePage';
import { ThemeContextProvider } from './contexts/ThemeContext';
import { SecurityProvider } from './contexts/SecurityContext';
import { ErrorBoundary } from './components/ErrorBoundary';
import Layout from './components/Layout';

const AppRoutes = () => {
  const location = useLocation();

  // Define routes that should use full-width layout
  const fullWidthRoutes = ['/paste/new', '/paste/'];
  const isFullWidth = fullWidthRoutes.some(
    route =>
      location.pathname === route ||
      (route === '/paste/' &&
        location.pathname.startsWith('/paste/') &&
        !location.pathname.endsWith('/raw'))
  );

  return (
    <Layout maxWidth={isFullWidth ? false : 'lg'} disableGutters={isFullWidth}>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/about" element={<AboutPage />} />
        <Route path="/paste/new" element={<CreatePastePage />} />
        <Route path="/paste/:id" element={<PastePage />} />
        <Route path="/paste/:id/raw" element={<RawPastePage />} />
      </Routes>
    </Layout>
  );
};

const App = () => {
  return (
    <SecurityProvider>
      <ThemeContextProvider>
        <Router>
          <ErrorBoundary>
            <AppRoutes />
          </ErrorBoundary>
        </Router>
      </ThemeContextProvider>
    </SecurityProvider>
  );
};

export default App;
