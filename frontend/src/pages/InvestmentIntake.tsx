import { useState, type ChangeEvent, type FormEvent } from 'react';
import {
  Alert,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
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
import { saveInvestment, updateInvestment, type Investment, type InvestmentInput } from '../lib/api';
import InvestmentChatDialog from '../components/investments/InvestmentChatDialog';
import { scriptForAssetClass } from '../investment-chat/scripts';

const initialForm: InvestmentInput = {
  name: '',
  ticker: '',
  assetClass: 'Stock',
  purchasePrice: 0,
  couponRate: 0,
  maturityDate: '',
  callableDateStart: '',
  callPrice: 0,
  callDate: '',
  thesis: '',
  targetAllocation: '',
  initialInvestment: 0,
  initialInvestmentDate: '',
  notes: '',
};

function formatDollarInput(value: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 20,
  }).format(value);
}

function formatPercentInput(value: number): string {
  return `${new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 3,
  }).format(value)}%`;
}

function parseNumericInput(value: string): number {
  const cleaned = value.replace(/[^0-9.-]/g, '');
  const parsed = Number(cleaned);
  return Number.isFinite(parsed) ? parsed : 0;
}

function isBondAssetClass(assetClass: string): boolean {
  return assetClass === 'Bond - US Treasury'
    || assetClass === 'Bond - State or Municipal'
    || assetClass === 'Bond - Corporate';
}

export default function InvestmentIntake() {
  const navigate = useNavigate();
  const [form, setForm] = useState<InvestmentInput>(initialForm);
  const [purchasePriceInput, setPurchasePriceInput] = useState(formatDollarInput(initialForm.purchasePrice));
  const [couponRateInput, setCouponRateInput] = useState(formatPercentInput(initialForm.couponRate));
  const [callPriceInput, setCallPriceInput] = useState(formatDollarInput(initialForm.callPrice));
  const [initialInvestmentInput, setInitialInvestmentInput] = useState(formatDollarInput(initialForm.initialInvestment));
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [savedInvestment, setSavedInvestment] = useState<Investment | null>(null);
  const [showChatPrompt, setShowChatPrompt] = useState(false);
  const [showChatDialog, setShowChatDialog] = useState(false);
  const [savingChatNotes, setSavingChatNotes] = useState(false);

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
      setSavedInvestment(investment);
      setShowChatPrompt(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save investment');
    } finally {
      setSaving(false);
    }
  };

  const handleClear = () => {
    setError(null);
    setForm(initialForm);
    setPurchasePriceInput(formatDollarInput(initialForm.purchasePrice));
    setCouponRateInput(formatPercentInput(initialForm.couponRate));
    setCallPriceInput(formatDollarInput(initialForm.callPrice));
    setInitialInvestmentInput(formatDollarInput(initialForm.initialInvestment));
  };

  const continueToInvestment = () => {
    if (!savedInvestment) {
      return;
    }
    navigate(`/investments/${savedInvestment.uuid}`);
  };

  const handleSkipChat = () => {
    setShowChatPrompt(false);
    continueToInvestment();
  };

  const handleOpenChat = () => {
    setShowChatPrompt(false);
    setShowChatDialog(true);
  };

  const handleSaveChatSummary = async (summary: string) => {
    if (!savedInvestment) {
      return;
    }
    if (!summary.trim()) {
      setShowChatDialog(false);
      continueToInvestment();
      return;
    }

    setSavingChatNotes(true);
    setError(null);
    try {
      const nextNotes = savedInvestment.notes
        ? `${savedInvestment.notes}\n\n${summary.trim()}`
        : summary.trim();
      const updated = await updateInvestment(savedInvestment.uuid, {
        name: savedInvestment.name,
        ticker: savedInvestment.ticker,
        assetClass: savedInvestment.assetClass,
        purchasePrice: savedInvestment.purchasePrice,
        couponRate: savedInvestment.couponRate,
        maturityDate: savedInvestment.maturityDate,
        callableDateStart: savedInvestment.callableDateStart,
        callPrice: savedInvestment.callPrice,
        callDate: savedInvestment.callDate,
        thesis: savedInvestment.thesis,
        targetAllocation: savedInvestment.targetAllocation,
        initialInvestment: savedInvestment.initialInvestment,
        initialInvestmentDate: savedInvestment.initialInvestmentDate,
        notes: nextNotes,
        saleAssumptions: savedInvestment.saleAssumptions,
        categories: savedInvestment.categories,
      });
      setSavedInvestment(updated);
      setShowChatDialog(false);
      navigate(`/investments/${updated.uuid}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save chat notes');
    } finally {
      setSavingChatNotes(false);
    }
  };
  const chatScript = scriptForAssetClass(savedInvestment?.assetClass || form.assetClass);

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

      <Dialog open={showChatPrompt} onClose={savingChatNotes ? undefined : handleSkipChat} maxWidth="sm" fullWidth>
        <DialogTitle>Continue With Chat?</DialogTitle>
        <DialogContent dividers>
          <Typography>
            Would you like to use the chat interface to provide more information about this investment?
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleSkipChat}>No</Button>
          <Button variant="contained" onClick={handleOpenChat}>Yes</Button>
        </DialogActions>
      </Dialog>

      <InvestmentChatDialog
        open={showChatDialog}
        script={chatScript}
        saving={savingChatNotes}
        error={error}
        onClose={() => {
          setShowChatDialog(false);
          continueToInvestment();
        }}
        onComplete={(summary: string) => void handleSaveChatSummary(summary)}
      />

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
              <MenuItem value="Bond - US Treasury">Bond - US Treasury</MenuItem>
              <MenuItem value="Bond - State or Municipal">Bond - State or Municipal</MenuItem>
              <MenuItem value="Bond - Corporate">Bond - Corporate</MenuItem>
              <MenuItem value="Fund">Fund</MenuItem>
              <MenuItem value="Residential Real Estate">Residential Real Estate</MenuItem>
              <MenuItem value="Commercial Real Estate">Commercial Real Estate</MenuItem>
              <MenuItem value="Private investment">Private investment</MenuItem>
              <MenuItem value="Other">Other</MenuItem>
            </Select>
          </FormControl>

          {isBondAssetClass(form.assetClass) ? (
            <>
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                <TextField
                  label="Purchase price"
                  value={purchasePriceInput}
                  onChange={(event: ChangeEvent<HTMLInputElement>) => {
                    const value = event.target.value;
                    setPurchasePriceInput(value);
                    update('purchasePrice', parseNumericInput(value));
                  }}
                  onBlur={() => {
                    setPurchasePriceInput(formatDollarInput(form.purchasePrice));
                  }}
                  inputProps={{ inputMode: 'decimal' }}
                  fullWidth
                />
                <TextField
                  label="Coupon rate"
                  value={couponRateInput}
                  onChange={(event: ChangeEvent<HTMLInputElement>) => {
                    const value = event.target.value;
                    setCouponRateInput(value);
                    update('couponRate', parseNumericInput(value));
                  }}
                  onBlur={() => {
                    setCouponRateInput(formatPercentInput(form.couponRate));
                  }}
                  inputProps={{ inputMode: 'decimal' }}
                  fullWidth
                />
              </Stack>

              <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                <TextField
                  label="Maturity date"
                  value={form.maturityDate}
                  onChange={(event: ChangeEvent<HTMLInputElement>) => update('maturityDate', event.target.value)}
                  type="date"
                  InputLabelProps={{ shrink: true }}
                  fullWidth
                />
                <TextField
                  label="Callable date start"
                  value={form.callableDateStart}
                  onChange={(event: ChangeEvent<HTMLInputElement>) => update('callableDateStart', event.target.value)}
                  type="date"
                  InputLabelProps={{ shrink: true }}
                  fullWidth
                />
              </Stack>

              <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                <TextField
                  label="Call price"
                  value={callPriceInput}
                  onChange={(event: ChangeEvent<HTMLInputElement>) => {
                    const value = event.target.value;
                    setCallPriceInput(value);
                    update('callPrice', parseNumericInput(value));
                  }}
                  onBlur={() => {
                    setCallPriceInput(formatDollarInput(form.callPrice));
                  }}
                  inputProps={{ inputMode: 'decimal' }}
                  fullWidth
                />
                <TextField
                  label="Call date"
                  value={form.callDate}
                  onChange={(event: ChangeEvent<HTMLInputElement>) => update('callDate', event.target.value)}
                  type="date"
                  InputLabelProps={{ shrink: true }}
                  fullWidth
                />
              </Stack>
            </>
          ) : null}

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
              value={initialInvestmentInput}
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                const value = event.target.value;
                setInitialInvestmentInput(value);
                update('initialInvestment', parseNumericInput(value));
              }}
              onBlur={() => {
                setInitialInvestmentInput(formatDollarInput(form.initialInvestment));
              }}
              inputProps={{ inputMode: 'decimal' }}
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
