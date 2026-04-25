import {
  Alert,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
} from '@mui/material';
import type { ChangeEvent, FormEvent } from 'react';
import type { InvestmentInput } from '../../lib/api';

type Props = {
  open: boolean;
  saving: boolean;
  error: string | null;
  form: InvestmentInput;
  initialInvestmentInput: string;
  onClose: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onChange: <K extends keyof InvestmentInput>(key: K, value: InvestmentInput[K]) => void;
  setInitialInvestmentInput: (value: string) => void;
  clearError: () => void;
};

export default function InvestmentDetailsDialog({
  open,
  saving,
  error,
  form,
  initialInvestmentInput,
  onClose,
  onSubmit,
  onChange,
  setInitialInvestmentInput,
  clearError,
}: Props) {
  return (
    <Dialog open={open} onClose={saving ? undefined : onClose} fullWidth maxWidth="md">
      <DialogTitle>Investment Details</DialogTitle>
      <DialogContent dividers>
        <Stack component="form" spacing={2} id="investment-details-form" onSubmit={onSubmit}>
          {error ? <Alert severity="error">{error}</Alert> : null}

          <TextField
            label="Investment name"
            value={form.name}
            onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('name', event.target.value)}
            required
            fullWidth
          />

          <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
            <TextField
              label="Ticker or identifier"
              value={form.ticker}
              onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('ticker', event.target.value)}
              fullWidth
            />
            <TextField
              label="Asset class"
              value={form.assetClass}
              onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('assetClass', event.target.value)}
              required
              fullWidth
            />
          </Stack>

          <TextField
            label="Target allocation"
            value={form.targetAllocation}
            onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('targetAllocation', event.target.value)}
            fullWidth
          />

          <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
            <TextField
              label="Initial investment"
              value={initialInvestmentInput}
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                clearError();
                setInitialInvestmentInput(event.target.value);
              }}
              type="number"
              inputProps={{ step: 0.01 }}
              fullWidth
            />
            <TextField
              label="Initial investment date"
              value={form.initialInvestmentDate}
              onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('initialInvestmentDate', event.target.value)}
              type="date"
              InputLabelProps={{ shrink: true }}
              fullWidth
            />
          </Stack>

          <TextField
            label="Investment thesis"
            value={form.thesis}
            onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('thesis', event.target.value)}
            multiline
            minRows={4}
            required
            fullWidth
          />

          <TextField
            label="Notes"
            value={form.notes}
            onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('notes', event.target.value)}
            multiline
            minRows={3}
            fullWidth
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={saving}>Cancel</Button>
        <Button type="submit" form="investment-details-form" variant="contained" disabled={saving}>
          {saving ? <CircularProgress size={20} color="inherit" /> : 'Save Investment Details'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
