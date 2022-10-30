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
  DialogTitle,
  Dialog,
  DialogContentText,
  DialogContent,
  DialogActions,
  SpeedDial,
  SpeedDialIcon,
} from '@mui/material';

import {
  Delete as DeleteIcon,
  Edit as EditIcon,
} from '@mui/icons-material';

import { createTheme, ThemeProvider } from '@mui/material/styles';
import { blue, lightBlue } from '@mui/material/colors';

import beautify from 'json-beautify';
import api from '../api';

const theme = createTheme();

const showAddress = (address) => {
  if (address === null) {
    return 'Any'
  } else {
    return typeof address === 'string' ? address : address.join(', ')
  }
}

function Snat() {
  const [data, setData] = useState([]);
  const [activeSnat, setActiveSnat] = useState({});
  const [dialog, setDialog] = useState(false);

  const fetchData = async () => {
    const res = await fetch('/api/v1/nat');
    const data = await res.json();
    setData(data)
  };
  useEffect(() => {
    fetchData();
  }, []);

  return (
    <div style={{ display: 'flex', flexDirection: 'column' }}>
      <TableContainer component={Paper}>
        <Table sx={{ minWidth: 650 }} aria-label="simple table">
          <TableHead>
            <TableRow>
              <TableCell>ID</TableCell>
              <TableCell>Source IP</TableCell>
              <TableCell>Destination IP</TableCell>
              <TableCell>Egress</TableCell>
              <TableCell>SNAT Target</TableCell>
              <TableCell></TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {
              data.map(d => (
                <TableRow key={d.id}>
                  <TableCell>{ d.id }</TableCell>
                  <TableCell>{ showAddress(d.source) }</TableCell>
                  <TableCell>{ showAddress(d.destination) }</TableCell>
                  <TableCell>{ d.egress }</TableCell>
                  <TableCell>{ d.target === 0 ? 'Dynamic (Egress Masquerade)' : `静态 ${d.target_ip.IP}` }</TableCell>
                  <TableCell>
                    <IconButton aria-label="edit">
                      <EditIcon />
                    </IconButton>
                    <IconButton
                      onClick={() => {
                        api.removePolicy(d.id)
                      }}
                      aria-label="delete" color="error">
                      <DeleteIcon />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))
            }
          </TableBody>
        </Table>
      </TableContainer>
      <Dialog
        open={dialog}
        onClose={() => setDialog(false)}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <DialogTitle id="alert-dialog-title">
          
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">
          </DialogContentText>
        </DialogContent>
        <DialogActions>
        </DialogActions>
      </Dialog>
      <SpeedDial
        ariaLabel="add new nat rule"
        sx={{ position: 'absolute', bottom: 16, right: 16 }}
        onClick={() => {
          
        }}
        open={false}
        icon={<SpeedDialIcon />}
      >
      </SpeedDial>
    </div>
  );
}

export default Snat;
