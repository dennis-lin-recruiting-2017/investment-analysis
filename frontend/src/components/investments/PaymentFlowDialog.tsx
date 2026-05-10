import {
  Alert,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  MenuItem,
  Stack,
  Tab,
  Tabs,
  TextField,
} from '@mui/material';
import type { ChangeEvent, FormEvent, SyntheticEvent } from 'react';
import type { InvestmentExpenseInput } from '../../lib/api';
import type { FlowTab } from './investmentDetailShared';

type Props = {
  open: boolean;
  saving: boolean;
  error: string | null;
  editingExpenseId: string | null;
  flowTab: FlowTab;
  expenseForm: InvestmentExpenseInput;
  amountInput: string;
  categories: string[];
  onClose: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onTabChange: (_event: SyntheticEvent, value: FlowTab) => void;
  onExpenseChange: <K extends keyof InvestmentExpenseInput>(key: K, value: InvestmentExpenseInput[K]) => void;
  setAmountInput: (value: string) => void;
  clearError: () => void;
};

export default function PaymentFlowDialog({
  open,
  saving,
  error,
  editingExpenseId,
  flowTab,
  expenseForm,
  amountInput,
  categories,
  onClose,
  onSubmit,
  onTabChange,
  onExpenseChange,
  setAmountInput,
  clearError,
}: Props) {
  return (
    <Dialog open={open} onClose={saving ? undefined : onClose} fullWidth maxWidth="md">
      <DialogTitle>{editingExpenseId === null ? 'Add Payment Flow' : 'Edit Payment Flow'}</DialogTitle>
      <DialogContent dividers>
        <Stack component="form" spacing={2} onSubmit={onSubmit} id="payment-flow-form">
          {error ? <Alert severity="error">{error}</Alert> : null}

          <Tabs value={flowTab} onChange={onTabChange} variant="fullWidth">
            <Tab value="one-time" label="One-Time Payment / Income" />
            <Tab value="recurring" label="Recurrent Payment / Income" />
          </Tabs>

          <TextField
            label="Name"
            value={expenseForm.label}
            onChange={(event: ChangeEvent<HTMLInputElement>) => onExpenseChange('label', event.target.value)}
            required
            fullWidth
          />

          <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
            <TextField
              select
              label="Event type"
              value={expenseForm.eventType}
              onChange={(event: ChangeEvent<HTMLInputElement>) => onExpenseChange('eventType', event.target.value as 'cash-flow' | 'deferred-tax')}
              fullWidth
            >
              <MenuItem value="cash-flow">Cash flow event</MenuItem>
              <MenuItem value="deferred-tax">Deferred tax event</MenuItem>
            </TextField>
            <TextField
              select
              label="Category"
              value={expenseForm.category}
              onChange={(event: ChangeEvent<HTMLInputElement>) => onExpenseChange('category', event.target.value)}
              fullWidth
            >
              <MenuItem value="">Uncategorized</MenuItem>
              {categories.map((category: string) => (
                <MenuItem key={category} value={category}>{category}</MenuItem>
              ))}
            </TextField>
            <TextField
              label="Amount"
              value={amountInput}
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                clearError();
                setAmountInput(event.target.value);
              }}
              type="number"
              inputProps={{ step: 0.01 }}
              required
              fullWidth
            />
          </Stack>

          <TextField
            label="Description"
            value={expenseForm.description}
            onChange={(event: ChangeEvent<HTMLInputElement>) => onExpenseChange('description', event.target.value)}
            multiline
            minRows={2}
            fullWidth
          />

          {flowTab === 'one-time' ? (
            <TextField
              label="Payment date"
              value={expenseForm.endDate}
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                onExpenseChange('endDate', event.target.value);
                onExpenseChange('startDate', event.target.value);
              }}
              type="date"
              InputLabelProps={{ shrink: true }}
              required
              fullWidth
            />
          ) : (
            <>
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                <TextField
                  label="Start date"
                  value={expenseForm.startDate}
                  onChange={(event: ChangeEvent<HTMLInputElement>) => onExpenseChange('startDate', event.target.value)}
                  type="date"
                  InputLabelProps={{ shrink: true }}
                  required
                  fullWidth
                />
                <TextField
                  label="End date"
                  value={expenseForm.endDate}
                  onChange={(event: ChangeEvent<HTMLInputElement>) => onExpenseChange('endDate', event.target.value)}
                  type="date"
                  InputLabelProps={{ shrink: true }}
                  required
                  fullWidth
                />
              </Stack>

              <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                <TextField
                  select
                  label="Recurs"
                  value={expenseForm.recurrenceInterval}
                  onChange={(event: ChangeEvent<HTMLInputElement>) =>
                    onExpenseChange('recurrenceInterval', event.target.value as '' | 'daily' | 'weekly' | 'monthly' | 'annually')
                  }
                  required
                  fullWidth
                >
                  <MenuItem value="daily">Daily</MenuItem>
                  <MenuItem value="weekly">Weekly</MenuItem>
                  <MenuItem value="monthly">Monthly</MenuItem>
                  <MenuItem value="annually">Annually</MenuItem>
                </TextField>
                <TextField
                  select
                  label="Adjustment frequency"
                  value={expenseForm.adjustmentFrequency}
                  onChange={(event: ChangeEvent<HTMLInputElement>) => onExpenseChange('adjustmentFrequency', event.target.value as '' | 'month' | 'year')}
                  fullWidth
                >
                  <MenuItem value="">None</MenuItem>
                  <MenuItem value="month">Month</MenuItem>
                  <MenuItem value="year">Year</MenuItem>
                </TextField>
                <TextField
                  select
                  label="Adjustment type"
                  value={expenseForm.adjustmentMode}
                  onChange={(event: ChangeEvent<HTMLInputElement>) => onExpenseChange('adjustmentMode', event.target.value as '' | 'amount' | 'percentage')}
                  fullWidth
                >
                  <MenuItem value="">None</MenuItem>
                  <MenuItem value="amount">Amount</MenuItem>
                  <MenuItem value="percentage">Percentage</MenuItem>
                </TextField>
                <TextField
                  label={expenseForm.adjustmentMode === 'percentage' ? 'Adjustment percentage' : 'Adjustment amount'}
                  value={expenseForm.adjustmentValue}
                  onChange={(event: ChangeEvent<HTMLInputElement>) => onExpenseChange('adjustmentValue', Number(event.target.value))}
                  type="number"
                  inputProps={{ step: expenseForm.adjustmentMode === 'percentage' ? 0.1 : 0.01 }}
                  fullWidth
                />
              </Stack>
            </>
          )}

          <TextField
            label="Notes"
            value={expenseForm.notes}
            onChange={(event: ChangeEvent<HTMLInputElement>) => onExpenseChange('notes', event.target.value)}
            multiline
            minRows={2}
            fullWidth
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={saving}>Cancel</Button>
        <Button type="submit" form="payment-flow-form" variant="contained" disabled={saving}>
          {saving ? <CircularProgress size={20} color="inherit" /> : (editingExpenseId === null ? 'Save Cash Flow' : 'Save Changes')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
