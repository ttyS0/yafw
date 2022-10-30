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
  TableContainer,
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
  Paper,
} from '@mui/material';
import {
  Menu as MenuIcon,
} from '@mui/icons-material'

import { createTheme, ThemeProvider } from '@mui/material/styles';
import { blue, lightBlue } from '@mui/material/colors';

import beautify from 'json-beautify';

const theme = createTheme();

const showAddress = (address) => {
  if (address === null) {
    return 'Any'
  } else {
    return typeof address === 'string' ? address : address.join(', ')
  }
}

function Connection() {
  const [data, setData] = useState({});
  const fetchData = async () => {
    const res = await fetch('/api/v1/connections');
    const data = await res.json();
    setData(data)
  };
  useEffect(() => {
    setInterval(() => fetchData(), 1000);
  }, []);
  return (
    <div style={{ display: 'flex', flexDirection: 'column' }}>
      <TextField
        id="outlined-multiline-static"
        multiline
        rows={30}
        value={data.raw && atob(data.raw).split('\n').filter(l => l.indexOf('9085') === -1).join('\n')}
        style={{ width: '100%', flex: 1 }}
        readonly
      />
    </div>
  );
}

export default Connection;
