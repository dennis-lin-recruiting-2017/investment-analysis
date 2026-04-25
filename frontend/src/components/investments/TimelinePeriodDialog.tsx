import {
  Divider,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material';
import type { InvestmentExpense } from '../../lib/api';
import {
  expandExpenseOccurrences,
  formatCurrency,
  formatDisplayDate,
  formatPercent,
  parseLocalDate,
  periodEndDate,
  recurringSummary,
  type GroupedExpense,
  type TimelineOccurrence,
  type ZoomLevel,
} from './investmentDetailShared';

type Props = {
  open: boolean;
  group: GroupedExpense | null;
  zoomLevel: ZoomLevel;
  allExpenses: InvestmentExpense[];
  initialInvestment: number;
  initialInvestmentDate: string;
  onClose: () => void;
};

export default function TimelinePeriodDialog({
  open,
  group,
  zoomLevel,
  allExpenses,
  initialInvestment,
  initialInvestmentDate,
  onClose,
}: Props) {
  const cashFlowOccurrences = group?.occurrences.filter(
    (occurrence: TimelineOccurrence) => occurrence.expense.eventType !== 'deferred-tax'
  ) || [];
  const deferredTaxOccurrences = group?.occurrences.filter(
    (occurrence: TimelineOccurrence) => occurrence.expense.eventType === 'deferred-tax'
  ) || [];

  const annualizedRate = (() => {
    if (!group) return null;
    if (initialInvestment <= 0) return null;

    const cutoff = periodEndDate(group.key, zoomLevel);
    const relevantOccurrences = allExpenses.flatMap((expense: InvestmentExpense) =>
      expandExpenseOccurrences(expense).filter(
        (occurrence: TimelineOccurrence) =>
          occurrence.date <= cutoff && occurrence.expense.eventType !== 'deferred-tax'
      )
    );
    const startDate = parseLocalDate(initialInvestmentDate);
    if (!startDate) return null;
    if (cutoff <= startDate) return null;

    const cumulativeCashFlow = relevantOccurrences.reduce(
      (sum: number, occurrence: TimelineOccurrence) => sum + occurrence.amount,
      0
    );
    const endingValue = initialInvestment + cumulativeCashFlow;
    if (endingValue <= 0) return null;

    const elapsedYears = (cutoff.getTime() - startDate.getTime()) / (365.25 * 24 * 60 * 60 * 1000);
    if (elapsedYears <= 0) return null;

    return Math.pow(endingValue / initialInvestment, 1 / elapsedYears) - 1;
  })();

  return (
    <Dialog
      open={open}
      onClose={onClose}
      fullWidth
      maxWidth="md"
      PaperProps={{
        sx: {
          maxHeight: '50vh',
        },
      }}
    >
      <DialogTitle>{group ? `${group.label} Cash Flows` : 'Cash Flows'}</DialogTitle>
      <DialogContent
        dividers
        sx={{
          overflow: 'hidden',
        }}
      >
        {group ? (
          <Stack
            spacing={2}
            sx={{
              height: '100%',
              minHeight: 0,
            }}
          >
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
              <Typography color="text.secondary">
                Included in analysis: {formatCurrency(cashFlowOccurrences.reduce((sum: number, occurrence: TimelineOccurrence) => sum + occurrence.amount, 0))}
              </Typography>
              {annualizedRate !== null ? (
                <Typography color="text.secondary">
                  Annualized return (through this period): {formatPercent(annualizedRate)}
                </Typography>
              ) : (
                <Typography color="text.secondary">
                  Annualized return: add a positive initial investment and date before this period to calculate it.
                </Typography>
              )}
            </Stack>
            <Box
              sx={{
                overflowX: 'auto',
                overflowY: 'auto',
                minHeight: 0,
                maxHeight: deferredTaxOccurrences.length > 0 ? '18vh' : '26vh',
              }}
            >
              <Table size="small" sx={{ minWidth: 760 }}>
                <TableHead>
                  <TableRow>
                    <TableCell>Amount</TableCell>
                    <TableCell>Name</TableCell>
                    <TableCell>Category</TableCell>
                    <TableCell>Description</TableCell>
                    <TableCell>Date</TableCell>
                    <TableCell>Schedule</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {cashFlowOccurrences
                    .slice()
                    .sort((left: TimelineOccurrence, right: TimelineOccurrence) =>
                      left.date.getTime() - right.date.getTime() || left.expense.label.localeCompare(right.expense.label)
                    )
                    .map((occurrence: TimelineOccurrence) => (
                      <TableRow key={occurrence.key}>
                        <TableCell>{formatCurrency(occurrence.amount)}</TableCell>
                        <TableCell>{occurrence.expense.label}</TableCell>
                        <TableCell>{occurrence.expense.category || 'Uncategorized'}</TableCell>
                        <TableCell>{occurrence.expense.description || occurrence.expense.notes || 'None'}</TableCell>
                        <TableCell>{formatDisplayDate(occurrence.date)}</TableCell>
                        <TableCell>{recurringSummary(occurrence.expense)}</TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>
            </Box>
            {deferredTaxOccurrences.length > 0 ? (
              <>
                <Divider />
                <Stack spacing={1}>
                  <Typography variant="subtitle2">Deferred Tax Events</Typography>
                  <Typography color="text.secondary">
                    Deferred tax total for period: {formatCurrency(
                      deferredTaxOccurrences.reduce((sum: number, occurrence: TimelineOccurrence) => sum + occurrence.amount, 0)
                    )}
                  </Typography>
                  <Box
                    sx={{
                      overflowX: 'auto',
                      overflowY: 'auto',
                      minHeight: 0,
                      maxHeight: '18vh',
                    }}
                  >
                    <Table size="small" sx={{ minWidth: 760 }}>
                      <TableHead>
                        <TableRow>
                          <TableCell>Category</TableCell>
                          <TableCell>Amount</TableCell>
                          <TableCell>Name</TableCell>
                          <TableCell>Description</TableCell>
                          <TableCell>Date</TableCell>
                          <TableCell>Schedule</TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {deferredTaxOccurrences
                          .slice()
                          .sort((left: TimelineOccurrence, right: TimelineOccurrence) =>
                            (left.expense.category || 'Uncategorized').localeCompare(right.expense.category || 'Uncategorized') ||
                            left.date.getTime() - right.date.getTime() ||
                            left.expense.label.localeCompare(right.expense.label)
                          )
                          .map((occurrence: TimelineOccurrence) => (
                            <TableRow key={occurrence.key}>
                              <TableCell>{occurrence.expense.category || 'Uncategorized'}</TableCell>
                              <TableCell>{formatCurrency(occurrence.amount)}</TableCell>
                              <TableCell>{occurrence.expense.label}</TableCell>
                              <TableCell>{occurrence.expense.description || occurrence.expense.notes || 'None'}</TableCell>
                              <TableCell>{formatDisplayDate(occurrence.date)}</TableCell>
                              <TableCell>{recurringSummary(occurrence.expense)}</TableCell>
                            </TableRow>
                          ))}
                      </TableBody>
                    </Table>
                  </Box>
                </Stack>
              </>
            ) : null}
          </Stack>
        ) : null}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Close</Button>
      </DialogActions>
    </Dialog>
  );
}
