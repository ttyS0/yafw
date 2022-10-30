import { useEffect, useRef, useState } from 'react';

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
  ToggleButton,
  ToggleButtonGroup,
  Chip,
} from '@mui/material';
import {
  Add as AddIcon,
} from '@mui/icons-material'

import { createTheme, ThemeProvider } from '@mui/material/styles';
import { blue, lightBlue } from '@mui/material/colors';

import beautify from 'json-beautify';

const theme = createTheme();

function Address({ value = [], onChange = () => {} }) {
  const [type, setType] = useState(value === null ? 'any' : (typeof value === 'object' ? 'immediate' : 'ipset'));
  const [immediate, setImmediate] = useState((value !== null && typeof value === 'object') ? value : []);
  const [ipSet, setIpSet] = useState(typeof value === 'object' ? '' : value);
  const onChangeRef = useRef(onChange);

  useEffect(() => {
    switch (type) {
      case 'any':
        onChangeRef.current(null)
        break
      case 'immediate':
        onChangeRef.current(immediate)
        break
      case 'ipset':
        onChangeRef.current(ipSet)
        break
      default:
    }
  }, [type, immediate, ipSet])

  return (
    <div>
      <ToggleButtonGroup
        value={type}
        exclusive
        onChange={(_, type) => setType(type)}
        aria-label="text alignment"
      >
        <ToggleButton value="any" aria-label="address any">
          Any
        </ToggleButton>
        <ToggleButton value="immediate" aria-label="address immediate">
          Entries
        </ToggleButton>
        <ToggleButton value="ipset" aria-label="address ipset">
          IPSet
        </ToggleButton>
      </ToggleButtonGroup>
      {
        type === 'ipset' ?
        <TextField
          variant="standard"
          value={ipSet}
          label="IPSet Name"
          onChange={e => setIpSet(e.target.value)}
          InputLabelProps={{
            shrink: true,
          }}
        />
        :
        (
          type === 'immediate' && 
          <>
            {
              immediate.map((address, i) => (
                <Chip key={i} label={address} onDelete={() => {
                  setImmediate(immediate.filter((_, key) => key !== i))
                }} />
              ))
            }
            <Chip color="success" label="Add" icon={<AddIcon />} onClick={() => {
              setImmediate([
                ...immediate,
                prompt('New Address'),
              ])
            }} />
          </>
        )
      }
    </div>
  );
}

export default Address;
