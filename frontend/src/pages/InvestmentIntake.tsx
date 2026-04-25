import { useState, type ChangeEvent, type FormEvent } from 'react';
import {
  Alert,
  Button,
  CircularProgress,
  FormControl,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  TextField,
  Typography,
  type SelectChangeEvent,
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { saveInvestment, type InvestmentInput } from '../lib/api';

const initialForm: InvestmentInput = {
  name: '',
  ticker: '',
  assetClass: 'Stock',
  thesis: '',
  targetAllocation: '',
  initialInvestment: 0,
  initialInvestmentDate: '',
  notes: '',
};

export default function InvestmentIntake() {
  const navigate = useNavigate();
  const [form, setForm] = useState<InvestmentInput>(initialForm);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const update = <K extends keyof InvestmentInput>(key: K, value: InvestmentInput[K]) => {
    setError(null);
    setForm((current: InvestmentInput) => ({ ...current, [key]: value }));
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSaving(true);
    setError(null);

    try {
      const investment = await saveInvestment(form);
      navigate(`/investments/${investment.uuid}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save investment');
    } finally {
      setSaving(false);
    }
  };

  const handleClear = () => {
    setError(null);
    setForm(initialForm);
  };

  return (
    <Stack spacing={2}>
      <div>
        <Typography variant="h4" gutterBottom>
          Investment Intake
        </Typography>
        <Typography color="text.secondary">
          Capture a candidate investment, its role in the portfolio, and the starting thesis.
        </Typography>
      </div>

      {error ? <Alert severity="error">{error}</Alert> : null}

      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack component="form" spacing={2} onSubmit={handleSubmit}>
          <Typography variant="subtitle1">Investment details</Typography>

          <TextField
            label="Investment name"
            value={form.name}
            onChange={(event: ChangeEvent<HTMLInputElement>) => update('name', event.target.value)}
            required
            fullWidth
          />

          <TextField
            label="Ticker or identifier"
            value={form.ticker}
            onChange={(event: ChangeEvent<HTMLInputElement>) => update('ticker', event.target.value)}
            fullWidth
          />

          <FormControl fullWidth>
            <InputLabel id="asset-class-label">Asset class</InputLabel>
            <Select
              labelId="asset-class-label"
              label="Asset class"
              value={form.assetClass}
              onChange={(event: SelectChangeEvent) => update('assetClass', event.target.value)}
            >
              <MenuItem value="Stock">Stock</MenuItem>
              <MenuItem value="ETF">ETF</MenuItem>
              <MenuItem value="Bond">Bond</MenuItem>
              <MenuItem value="Fund">Fund</MenuItem>
              <MenuItem value="Residential Real Estate">Residential Real Estate</MenuItem>
              <MenuItem value="Commercial Real Estate">Commercial Real Estate</MenuItem>
              <MenuItem value="Private investment">Private investment</MenuItem>
              <MenuItem value="Other">Other</MenuItem>
            </Select>
          </FormControl>

          <TextField
            label="Target allocation"
            value={form.targetAllocation}
            onChange={(event: ChangeEvent<HTMLInputElement>) => update('targetAllocation', event.target.value)}
            placeholder="Example: 5%"
            fullWidth
          />

          <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
            <TextField
              label="Initial investment"
              value={form.initialInvestment}
              onChange={(event: ChangeEvent<HTMLInputElement>) => update('initialInvestment', Number(event.target.value))}
              type="number"
              inputProps={{ step: 0.01 }}
              fullWidth
            />

            <TextField
              label="Initial investment date"
              value={form.initialInvestmentDate}
              onChange={(event: ChangeEvent<HTMLInputElement>) => update('initialInvestmentDate', event.target.value)}
              type="date"
              InputLabelProps={{ shrink: true }}
              fullWidth
            />
          </Stack>

          <TextField
            label="Investment thesis"
            value={form.thesis}
            onChange={(event: ChangeEvent<HTMLInputElement>) => update('thesis', event.target.value)}
            multiline
            minRows={4}
            required
            fullWidth
          />

          <TextField
            label="Notes"
            value={form.notes}
            onChange={(event: ChangeEvent<HTMLInputElement>) => update('notes', event.target.value)}
            multiline
            minRows={3}
            fullWidth
          />

          <Stack direction="row" spacing={2}>
            <Button type="submit" variant="contained" disabled={saving}>
              {saving ? <CircularProgress size={20} color="inherit" /> : 'Save Investment'}
            </Button>
            <Button variant="outlined" onClick={handleClear} disabled={saving}>
              Clear
            </Button>
          </Stack>
        </Stack>
      </Paper>
    </Stack>
  );
}
