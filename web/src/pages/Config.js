import { useEffect, useState } from 'react';

import {
  Route,
  Navigate,
  useParams,
} from 'react-router-dom';

import {
  AppBar,
  Box,
  Button,
  Toolbar,
  IconButton,
  Typography,
  TextField,
} from '@mui/material';
import {
  Menu as MenuIcon,
} from '@mui/icons-material'

import { createTheme, ThemeProvider } from '@mui/material/styles';
import { blue, lightBlue } from '@mui/material/colors';

import beautify from 'json-beautify';

const theme = createTheme();

function Config() {
  const [config, setConfig] = useState('');
  const fetchConfig = async () => {
    const res = await fetch('/api/v1/export');
    const data = await res.json();
    setConfig(beautify(data, null, 2, 80));
  };

  const saveConfig = () => {
        const url = window.URL.createObjectURL(
          new Blob([JSON.stringify(config)]),
        );
        const link = document.createElement('a');
        link.href = url;
        link.setAttribute(
          'download',
          `config.json`,
        );
        document.body.appendChild(link);
        link.click();
        link.parentNode.removeChild(link);
  }

  useEffect(() => {
    fetchConfig();
  }, []);
  return (
    <div style={{ display: 'flex', flexDirection: 'column' }}>
      <TextField
        id="outlined-multiline-static"
        label="Multiline"
        multiline
        rows={30}
        value={config}
        style={{ width: '100%', flex: 1 }}
        onChange={(e) => setConfig(e.target.value)}
      />
      <Button onClick={() => saveConfig()}>保存配置</Button>
    </div>
  );
}

export default Config;
