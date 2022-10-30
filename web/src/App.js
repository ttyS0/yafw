import logo from './logo.svg';
import './App.css';

import { useState } from 'react';

import {
  Route,
  Navigate,
  useParams,
  Routes,
} from 'react-router-dom';

import {
  AppBar,
  Box,
  Button,
  Toolbar,
  IconButton,
  Typography,
} from '@mui/material';
import {
  Menu as MenuIcon,
} from '@mui/icons-material'

import { createTheme, ThemeProvider } from '@mui/material/styles';
import { blue, lightBlue } from '@mui/material/colors';
import routes from './routes';

import { BrowserRouter, Link, Switch } from 'react-router-dom';

const theme = createTheme();

function App() {
  const [route, setRoute] = useState({})
  return (
    <ThemeProvider theme={theme}>
      <div>
        <AppBar position="static">
          <Toolbar>
            <IconButton
              size="large"
              edge="start"
              color="inherit"
              aria-label="open drawer"
              sx={{ mr: 2 }}
            >
              <MenuIcon />
            </IconButton>
            <Typography
              variant="h6"
              noWrap
              component="div"
              sx={{ flexGrow: 1, display: { xs: 'none', sm: 'block' } }}
            >
              YAFW - Yet Another Firewall
            </Typography>
            {
              routes.map(r => (
                <Button as={Link} to={r.path} key={r.path} color="inherit" variant="secondary">
                  {r.title}
                </Button>
              ))
            }
          </Toolbar>
        </AppBar>
        <Routes>
          {
            routes.map(r => (
              <Route path={r.path} element={r.component} key={r.path} />
            ))
          }
        </Routes>
      </div>
    </ThemeProvider>
  );
}

export default App;
