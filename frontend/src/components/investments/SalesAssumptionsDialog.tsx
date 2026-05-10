import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  MenuItem,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from '@mui/material';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import EditOutlinedIcon from '@mui/icons-material/EditOutlined';
import type { ChangeEvent, FormEvent } from 'react';
import type { InvestmentSaleAssumption, InvestmentSaleAssumptionInput } from '../../lib/api';

type Props = {
  open: boolean;
  saving: boolean;
  deletingId: string | null;
  editingId: string | null;
  error: string | null;
  form: InvestmentSaleAssumptionInput;
  amountInput: string;
  assumptions: InvestmentSaleAssumption[];
  onClose: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onChange: <K extends keyof InvestmentSaleAssumptionInput>(key: K, value: InvestmentSaleAssumptionInput[K]) => void;
  onAmountInputChange: (value: string) => void;
  onEdit: (assumption: InvestmentSaleAssumption) => void;
  onDelete: (assumption: InvestmentSaleAssumption) => void;
  clearError: () => void;
};

export default function SalesAssumptionsDialog({
  open,
  saving,
  deletingId,
  editingId,
  error,
  form,
  amountInput,
  assumptions,
  onClose,
  onSubmit,
  onChange,
  onAmountInputChange,
  onEdit,
  onDelete,
  clearError,
}: Props) {
  return (
    <Dialog open={open} onClose={saving ? undefined : onClose} fullWidth maxWidth="lg">
      <DialogTitle>Sales Assumptions</DialogTitle>
      <DialogContent dividers>
        <Stack spacing={2}>
          {error ? <Alert severity="error">{error}</Alert> : null}

          <Stack component="form" spacing={2} id="sale-assumption-form" onSubmit={onSubmit}>
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                label="Name"
                value={form.label}
                onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('label', event.target.value)}
                required
                fullWidth
              />
              <TextField
                label="Amount"
                value={amountInput}
                onChange={(event: ChangeEvent<HTMLInputElement>) => {
                  clearError();
                  onAmountInputChange(event.target.value);
                }}
                type="number"
                inputProps={{ step: 0.01 }}
                fullWidth
              />
            </Stack>

            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                select
                label="Amount behavior"
                value={form.growthType}
                onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('growthType', event.target.value as 'fixed' | 'growing')}
                fullWidth
              >
                <MenuItem value="fixed">Fixed number</MenuItem>
                <MenuItem value="growing">Grows per period</MenuItem>
              </TextField>
              {form.growthType === 'growing' ? (
                <>
                  <TextField
                    select
                    label="Growth period"
                    value={form.growthPeriod}
                    onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('growthPeriod', event.target.value as '' | 'month' | 'year')}
                    fullWidth
                  >
                    <MenuItem value="month">Month</MenuItem>
                    <MenuItem value="year">Year</MenuItem>
                  </TextField>
                  <TextField
                    select
                    label="Growth type"
                    value={form.growthMode}
                    onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('growthMode', event.target.value as '' | 'amount' | 'percentage')}
                    fullWidth
                  >
                    <MenuItem value="amount">Amount</MenuItem>
                    <MenuItem value="percentage">Percentage</MenuItem>
                  </TextField>
                  <TextField
                    label={form.growthMode === 'percentage' ? 'Growth percentage' : 'Growth amount'}
                    value={form.growthValue}
                    onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('growthValue', Number(event.target.value))}
                    type="number"
                    inputProps={{ step: form.growthMode === 'percentage' ? 0.1 : 0.01 }}
                    fullWidth
                  />
                </>
              ) : null}
            </Stack>

            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                label="Start date"
                value={form.startDate}
                onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('startDate', event.target.value)}
                type="date"
                InputLabelProps={{ shrink: true }}
                fullWidth
              />
              <TextField
                label="End date"
                value={form.endDate}
                onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('endDate', event.target.value)}
                type="date"
                InputLabelProps={{ shrink: true }}
                fullWidth
              />
            </Stack>

            <TextField
              label="Description"
              value={form.description}
              onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('description', event.target.value)}
              multiline
              minRows={2}
              fullWidth
            />

            <TextField
              label="Notes"
              value={form.notes}
              onChange={(event: ChangeEvent<HTMLInputElement>) => onChange('notes', event.target.value)}
              multiline
              minRows={2}
              fullWidth
            />

            <Stack direction="row" spacing={2}>
              <Button type="submit" variant="contained" disabled={saving}>
                {saving ? <CircularProgress size={20} color="inherit" /> : (editingId === null ? 'Add Sale Assumption' : 'Save Changes')}
              </Button>
              {editingId !== null ? (
                <Button
                  variant="outlined"
                  onClick={onClose}
                  disabled={saving}
                >
                  Done
                </Button>
              ) : null}
            </Stack>
          </Stack>

          <Box sx={{ overflowX: 'auto' }}>
            <Table size="small" sx={{ minWidth: 900 }}>
              <TableHead>
                <TableRow>
                  <TableCell sx={{ width: 180 }}>Amount</TableCell>
                  <TableCell>Name</TableCell>
                  <TableCell>Description</TableCell>
                  <TableCell>Start Date</TableCell>
                  <TableCell>End Date</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {assumptions.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6}>
                      <Typography color="text.secondary">No sale assumptions yet.</Typography>
                    </TableCell>
                  </TableRow>
                ) : (
                  assumptions.map((assumption: InvestmentSaleAssumption) => (
                    <TableRow key={assumption.id} hover>
                      <TableCell>
                        <Stack spacing={0.5}>
                          <Typography>{new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(assumption.amount)}</Typography>
                          <Typography variant="caption" color="text.secondary">
                            {assumption.growthType === 'growing'
                              ? `${assumption.growthMode === 'percentage' ? `${assumption.growthValue}%` : new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(assumption.growthValue)} per ${assumption.growthPeriod}`
                              : 'Fixed'}
                          </Typography>
                        </Stack>
                      </TableCell>
                      <TableCell>{assumption.label}</TableCell>
                      <TableCell>{assumption.description || assumption.notes || 'None'}</TableCell>
                      <TableCell>{assumption.startDate || '-'}</TableCell>
                      <TableCell>{assumption.endDate || '-'}</TableCell>
                      <TableCell align="right">
                        <Stack direction="row" spacing={1} justifyContent="flex-end">
                          <IconButton
                            size="small"
                            aria-label={`Edit ${assumption.label}`}
                            onClick={() => onEdit(assumption)}
                            disabled={deletingId === assumption.id}
                          >
                            <EditOutlinedIcon fontSize="small" />
                          </IconButton>
                          <IconButton
                            size="small"
                            color="error"
                            aria-label={`Delete ${assumption.label}`}
                            onClick={() => onDelete(assumption)}
                            disabled={deletingId === assumption.id}
                          >
                            <DeleteOutlineIcon fontSize="small" />
                          </IconButton>
                        </Stack>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </Box>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={saving}>Close</Button>
      </DialogActions>
    </Dialog>
  );
}
