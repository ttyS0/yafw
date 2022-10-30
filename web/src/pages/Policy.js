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
  Chip,
  DialogActions,
  DialogContentText,
  DialogContent,
  DialogTitle,
  Dialog,
  SpeedDial,
  SpeedDialIcon,
  FormLabel,
  MenuItem,
  Select,
  InputLabel,
  FormControl,
  Switch,
} from '@mui/material';
import {
  Menu as MenuIcon,
} from '@mui/icons-material'

import {
  Delete as DeleteIcon,
  Edit as EditIcon,
} from '@mui/icons-material';

import { createTheme, ThemeProvider } from '@mui/material/styles';
import { blue, lightBlue } from '@mui/material/colors';

import beautify from 'json-beautify';
import api from '../api';
import Address from '../components/Address';

const theme = createTheme();

const showAddress = (address) => {
  if (address === null) {
    return 'Any'
  } else {
    return typeof address === 'string' ? address : address.join(', ')
  }
}

const showAction = (action) => {
  switch (action) {
    case "accept":
      return <Chip label="Accept" color="success" />
    case "drop":
      return <Chip label="Drop" color="error" />
    default:
      return <Chip label="Unknown" />
  }
}

function Policy() {
  const [data, setData] = useState([]);
  const [activePolicy, setActivePolicy] = useState({});
  const [dialog, setDialog] = useState(false);

  const fetchData = async () => {
    setData(await api.policies())
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
              <TableCell align="right">ID</TableCell>
              <TableCell>Source IP</TableCell>
              <TableCell>Destination IP</TableCell>
              <TableCell>Action</TableCell>
              <TableCell></TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {
              data.map((d, i) => (
                <TableRow key={d.id}>
                  <TableCell align="right">{ d.id }</TableCell>
                  <TableCell>{ showAddress(d.source) }</TableCell>
                  <TableCell>{ showAddress(d.destination) }</TableCell>
                  <TableCell>{ showAction(d.action) }</TableCell>
                  <TableCell>
                    <IconButton
                      onClick={async () => {
                        setActivePolicy(d)
                        setDialog(true)
                      }}
                      aria-label="edit">
                      <EditIcon />
                    </IconButton>
                    <IconButton
                      onClick={async () => {
                        await api.removePolicy(d.id);
                        await fetchData();
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
          { activePolicy.id === undefined ? `Create a new policy` : `Edit policy #${activePolicy.id}` }
        </DialogTitle>
        <DialogContent>
          <Box>
            <FormControl fullWidth>
              <FormLabel>Source IP</FormLabel>
              <Address value={activePolicy.source} onChange={v => {
                setActivePolicy({
                  ...activePolicy,
                  source: v,
                })
              }}/>
            </FormControl>
            <FormControl fullWidth>
              <FormLabel>Destination IP</FormLabel>
              <Address value={activePolicy.destination} onChange={v => {
                console.log(v)
                setActivePolicy({
                  ...activePolicy,
                  destination: v,
                })
              }}/>
            </FormControl>
            <FormControl fullWidth>
              <FormLabel>Action</FormLabel>
              {/* <InputLabel id="action-select-label">Action</InputLabel> */}
              <Select
                labelId="action-select-label"
                id="action-select"
                value={activePolicy.action}
                onChange={e => {
                  setActivePolicy({
                    ...activePolicy,
                    action: e.target.value,
                  })
                }}
              >
                <MenuItem value="accept">Accept</MenuItem>
                <MenuItem value="drop">Drop</MenuItem>
              </Select>
            </FormControl>
            <FormControl fullWidth>
              <FormLabel>Log</FormLabel>
              <Switch
                labelId="log-switch-label"
                id="log-switch"
                checked={activePolicy.log}
                onChange={e => {
                  console.log(e)
                  setActivePolicy({
                    ...activePolicy,
                    log: e.target.checked,
                  })
                }}
              />
            </FormControl>
          </Box>
        </DialogContent>
        <DialogActions>
          {
            activePolicy.id === undefined ?
            <Button
              color="success"
              onClick={async () => {
                await api.addPolicy(activePolicy)
                await fetchData()
                setDialog(false)
              }}
            >
            Create
            </Button>
            :
            <Button
              onClick={async () => {
                await api.modifyPolicy(activePolicy.id, activePolicy)
                await fetchData()
                setDialog(false)
              }}
            >
              Apply
            </Button>
          }
          <Button
            onClick={() => {
              setDialog(false)
            }}
            color="error"
            autoFocus
          >
            Cancel
          </Button>
        </DialogActions>
      </Dialog>
      <SpeedDial
        ariaLabel="add new policy"
        sx={{ position: 'absolute', bottom: 16, right: 16 }}
        onClick={() => {
          setActivePolicy({
            source: null,
            destination: null,
            action: 'accept',
            log: false,
          })
          setDialog(true)
        }}
        open={false}
        icon={<SpeedDialIcon />}
      >
      </SpeedDial>
    </div>
  );
}

export default Policy;
