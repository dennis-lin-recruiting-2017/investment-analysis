import React from 'react';
import ReactDOM from 'react-dom/client';
import {
  Box,
  CssBaseline,
  ThemeProvider,
  Toolbar,
  createTheme,
} from '@mui/material';
import { HashRouter, Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom';
import Header from './components/Header';
import AppDrawer from './components/AppDrawer';
import Home from './pages/Home';
import InvestmentDetail from './pages/InvestmentDetail';
import InvestmentIntake from './pages/InvestmentIntake';
import Author from './pages/Author';
import Demo from './pages/Demo';
import LMStudioSettings from './pages/LMStudioSettings';
import OllamaSettings from './pages/OllamaSettings';
import ExternalEndpointSettings from './pages/ExternalEndpointSettings';

const drawerWidth = 300;

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#007AFF',
      light: '#5AC8FA',
      dark: '#0051D5',
      contrastText: '#FFFFFF',
    },
    secondary: {
      main: '#5856D6',
      contrastText: '#FFFFFF',
    },
    error: {
      main: '#FF3B30',
    },
    warning: {
      main: '#FF9500',
    },
    info: {
      main: '#5AC8FA',
    },
    success: {
      main: '#34C759',
    },
    background: {
      default: '#F2F2F7',
      paper: '#FFFFFF',
    },
    text: {
      primary: '#000000',
      secondary: '#6C6C70',
    },
    divider: 'rgba(60, 60, 67, 0.12)',
  },
  typography: {
    fontFamily: '-apple-system, BlinkMacSystemFont, "SF Pro Display", "SF Pro Text", "Helvetica Neue", Arial, sans-serif',
    h1: { fontWeight: 700, letterSpacing: '-0.5px' },
    h2: { fontWeight: 700, letterSpacing: '-0.5px' },
    h3: { fontWeight: 600, letterSpacing: '-0.3px' },
    h4: { fontWeight: 600, letterSpacing: '-0.2px' },
    h5: { fontWeight: 600 },
    h6: { fontWeight: 600 },
    button: { textTransform: 'none', fontWeight: 500 },
  },
  shape: {
    borderRadius: 10,
  },
  components: {
    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundColor: 'rgba(242, 242, 247, 0.85)',
          backdropFilter: 'blur(20px)',
          WebkitBackdropFilter: 'blur(20px)',
          boxShadow: '0 0.5px 0 rgba(60, 60, 67, 0.2)',
          color: '#000000',
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          boxShadow: 'none',
          '&:hover': { boxShadow: 'none' },
          '&:active': { boxShadow: 'none' },
        },
        containedPrimary: {
          '&:hover': { backgroundColor: '#0051D5' },
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
        },
        elevation1: {
          boxShadow: '0 1px 4px rgba(0, 0, 0, 0.08), 0 0.5px 1px rgba(0, 0, 0, 0.06)',
        },
      },
    },
    MuiDialog: {
      styleOverrides: {
        paper: {
          borderRadius: 14,
          boxShadow: '0 20px 60px rgba(0, 0, 0, 0.18)',
        },
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            borderRadius: 8,
            backgroundColor: '#FFFFFF',
          },
        },
      },
    },
    MuiOutlinedInput: {
      styleOverrides: {
        notchedOutline: {
          borderColor: 'rgba(60, 60, 67, 0.2)',
        },
      },
    },
    MuiTableCell: {
      styleOverrides: {
        root: {
          borderBottom: '0.5px solid rgba(60, 60, 67, 0.12)',
        },
        head: {
          fontWeight: 600,
          color: '#6C6C70',
          fontSize: '0.75rem',
          textTransform: 'uppercase',
          letterSpacing: '0.04em',
        },
      },
    },
    MuiTableRow: {
      styleOverrides: {
        root: {
          '&:last-child td': { border: 0 },
        },
      },
    },
    MuiDivider: {
      styleOverrides: {
        root: {
          borderColor: 'rgba(60, 60, 67, 0.12)',
        },
      },
    },
    MuiToggleButtonGroup: {
      styleOverrides: {
        root: {
          backgroundColor: 'rgba(118, 118, 128, 0.12)',
          borderRadius: 9,
          padding: 2,
          gap: 2,
          border: 'none',
        },
        grouped: {
          border: 'none !important',
          '&:not(:first-of-type)': { borderRadius: '7px !important', marginLeft: 0 },
          '&:first-of-type': { borderRadius: '7px !important' },
        },
      },
    },
    MuiToggleButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 500,
          border: 'none',
          borderRadius: '7px !important',
          padding: '4px 12px',
          color: '#000000',
          '&.Mui-selected': {
            backgroundColor: '#FFFFFF',
            color: '#000000',
            boxShadow: '0 1px 3px rgba(0, 0, 0, 0.12), 0 0.5px 1px rgba(0, 0, 0, 0.08)',
            '&:hover': { backgroundColor: '#FFFFFF' },
          },
          '&:hover': {
            backgroundColor: 'transparent',
          },
        },
      },
    },
    MuiTab: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 500,
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: { borderRadius: 8 },
      },
    },
    MuiAlert: {
      styleOverrides: {
        root: { borderRadius: 10 },
      },
    },
  },
});

function AppShell() {
  const navigate = useNavigate();
  const location = useLocation();
  const [mobileOpen, setMobileOpen] = React.useState(false);

  const handleDrawerToggle = () => {
    setMobileOpen((prev: boolean) => !prev);
  };

  const handleNavigate = (path: string) => {
    navigate(path);
    setMobileOpen(false);
  };

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <CssBaseline />

      <Header drawerWidth={drawerWidth} onMenuClick={handleDrawerToggle} />

      <AppDrawer
        drawerWidth={drawerWidth}
        mobileOpen={mobileOpen}
        onClose={handleDrawerToggle}
        pathname={location.pathname}
        onNavigate={handleNavigate}
      />

      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 3,
          width: { sm: `calc(100% - ${drawerWidth}px)` },
          backgroundColor: 'background.default',
        }}
      >
        <Toolbar />
        <Routes>
          <Route path="/" element={<Navigate to="/home" replace />} />
          <Route path="/home" element={<Home />} />
          <Route path="/investments/new" element={<InvestmentIntake />} />
          <Route path="/investments/intake" element={<Navigate to="/investments/new" replace />} />
          <Route path="/investments/:uuid" element={<InvestmentDetail />} />
          <Route path="/about/author" element={<Author />} />
          <Route path="/about/demo" element={<Demo />} />
          <Route path="/settings/lm-studio" element={<LMStudioSettings />} />
          <Route path="/settings/ollama" element={<OllamaSettings />} />
          <Route path="/settings/external-endpoint" element={<ExternalEndpointSettings />} />
          <Route path="*" element={<Navigate to="/home" replace />} />
        </Routes>
      </Box>
    </Box>
  );
}

function RootApp() {
  return (
    <ThemeProvider theme={theme}>
      <HashRouter>
        <AppShell />
      </HashRouter>
    </ThemeProvider>
  );
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <RootApp />
  </React.StrictMode>
);
