import type { InvestmentExpense, InvestmentExpenseInput, InvestmentInput } from '../../lib/api';

export type ZoomLevel = 'days' | 'weeks' | 'months' | 'years';
export type FlowTab = 'one-time' | 'recurring';

export type GroupedExpense = {
  key: string;
  label: string;
  amount: number;
  occurrences: TimelineOccurrence[];
};

export type TimelineOccurrence = {
  key: string;
  label: string;
  date: Date;
  amount: number;
  expense: InvestmentExpense;
};

export const initialExpenseForm: InvestmentExpenseInput = {
  eventType: 'cash-flow',
  flowType: 'one-time',
  recurrenceInterval: '',
  label: '',
  amount: 0,
  description: '',
  startDate: '',
  endDate: '',
  category: '',
  adjustmentFrequency: '',
  adjustmentMode: '',
  adjustmentValue: 0,
  notes: '',
};

export const initialInvestmentForm: InvestmentInput = {
  name: '',
  ticker: '',
  assetClass: '',
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

export function parseLocalDate(value: string): Date | null {
  if (!value) {
    return null;
  }
  const date = new Date(`${value}T00:00:00`);
  return Number.isNaN(date.getTime()) ? null : date;
}

export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 0,
  }).format(amount);
}

export function formatDisplayDate(value: Date | string): string {
  const date = typeof value === 'string' ? parseLocalDate(value) : value;
  if (!date) {
    return '-';
  }
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

export function formatPeriod(date: Date, zoomLevel: ZoomLevel): string {
  switch (zoomLevel) {
    case 'days':
      return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    case 'weeks': {
      const weekStart = new Date(date);
      weekStart.setDate(date.getDate() - date.getDay());
      return `Week of ${weekStart.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}`;
    }
    case 'months':
      return date.toLocaleDateString('en-US', { month: 'short', year: 'numeric' });
    case 'years':
      return date.toLocaleDateString('en-US', { year: 'numeric' });
  }
}

export function buildGroupKey(date: Date, zoomLevel: ZoomLevel): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');

  switch (zoomLevel) {
    case 'days':
      return `${year}-${month}-${day}`;
    case 'weeks': {
      const weekStart = new Date(date);
      weekStart.setDate(date.getDate() - date.getDay());
      const weekMonth = String(weekStart.getMonth() + 1).padStart(2, '0');
      const weekDay = String(weekStart.getDate()).padStart(2, '0');
      return `${weekStart.getFullYear()}-${weekMonth}-${weekDay}`;
    }
    case 'months':
      return `${year}-${month}`;
    case 'years':
      return `${year}`;
  }
}

export function effectiveEndDate(expense: InvestmentExpense): string {
  return expense.endDate || expense.dueDate;
}

function cloneDate(date: Date): Date {
  return new Date(date.getTime());
}

function daysInMonth(year: number, month: number): number {
  return new Date(year, month + 1, 0).getDate();
}

function buildAnchoredDate(year: number, month: number, day: number): Date {
  return new Date(year, month, Math.min(day, daysInMonth(year, month)));
}

function occurrenceDateForStep(startDate: Date, interval: InvestmentExpense['recurrenceInterval'], step: number): Date {
  const year = startDate.getFullYear();
  const month = startDate.getMonth();
  const day = startDate.getDate();

  switch (interval) {
    case 'daily': {
      const next = cloneDate(startDate);
      next.setDate(day + step);
      return next;
    }
    case 'weekly': {
      const next = cloneDate(startDate);
      next.setDate(day + step * 7);
      return next;
    }
    case 'monthly': {
      const targetMonth = month + step;
      const targetYear = year + Math.floor(targetMonth / 12);
      const normalizedMonth = ((targetMonth % 12) + 12) % 12;
      return buildAnchoredDate(targetYear, normalizedMonth, day);
    }
    case 'annually':
      return buildAnchoredDate(year + step, month, day);
    default:
      return cloneDate(startDate);
  }
}

function monthsBetween(start: Date, end: Date): number {
  let months = (end.getFullYear() - start.getFullYear()) * 12 + (end.getMonth() - start.getMonth());
  if (end.getDate() < start.getDate()) {
    months -= 1;
  }
  return Math.max(0, months);
}

function yearsBetween(start: Date, end: Date): number {
  let years = end.getFullYear() - start.getFullYear();
  if (
    end.getMonth() < start.getMonth() ||
    (end.getMonth() === start.getMonth() && end.getDate() < start.getDate())
  ) {
    years -= 1;
  }
  return Math.max(0, years);
}

function adjustedAmount(expense: InvestmentExpense, occurrenceDate: Date, startDate: Date): number {
  if (!expense.adjustmentFrequency || !expense.adjustmentMode || expense.adjustmentValue === 0) {
    return expense.amount;
  }

  const periods = expense.adjustmentFrequency === 'month'
    ? monthsBetween(startDate, occurrenceDate)
    : yearsBetween(startDate, occurrenceDate);

  if (periods <= 0) {
    return expense.amount;
  }

  if (expense.adjustmentMode === 'amount') {
    return expense.amount + expense.adjustmentValue * periods;
  }

  return expense.amount * Math.pow(1 + expense.adjustmentValue / 100, periods);
}

export function expandExpenseOccurrences(expense: InvestmentExpense): TimelineOccurrence[] {
  if (expense.flowType !== 'recurring') {
    const date = parseLocalDate(effectiveEndDate(expense));
    if (!date) {
      return [];
    }

    return [{
      key: `${expense.id}-${date.toISOString()}`,
      label: expense.label,
      date,
      amount: expense.amount,
      expense,
    }];
  }

  const startDate = parseLocalDate(expense.startDate);
  const endDate = parseLocalDate(effectiveEndDate(expense));
  if (!startDate || !endDate || !expense.recurrenceInterval) {
    return [];
  }

  const occurrences: TimelineOccurrence[] = [];
  let step = 0;

  while (step < 20000) {
    const date = occurrenceDateForStep(startDate, expense.recurrenceInterval, step);
    if (date > endDate) {
      break;
    }
    occurrences.push({
      key: `${expense.id}-${date.toISOString()}`,
      label: expense.label,
      date,
      amount: adjustedAmount(expense, date, startDate),
      expense,
    });
    step += 1;
  }

  return occurrences;
}

export function groupExpenses(expenses: InvestmentExpense[], zoomLevel: ZoomLevel): GroupedExpense[] {
  const grouped = new Map<string, GroupedExpense>();

  expenses.forEach((expense: InvestmentExpense) => {
    expandExpenseOccurrences(expense).forEach((occurrence: TimelineOccurrence) => {
      const key = buildGroupKey(occurrence.date, zoomLevel);
      const existing = grouped.get(key);
      if (existing) {
        existing.amount += occurrence.amount;
        existing.occurrences.push(occurrence);
        return;
      }

      grouped.set(key, {
        key,
        label: formatPeriod(occurrence.date, zoomLevel),
        amount: occurrence.amount,
        occurrences: [occurrence],
      });
    });
  });

  return Array.from(grouped.values()).sort((left: GroupedExpense, right: GroupedExpense) =>
    left.key.localeCompare(right.key)
  );
}

export function recurringSummary(expense: InvestmentExpense): string {
  if (expense.flowType !== 'recurring') {
    return 'One-time';
  }

  const parts: string[] = ['Recurring'];
  if (expense.recurrenceInterval) {
    parts.push(expense.recurrenceInterval);
  }
  if (expense.adjustmentFrequency && expense.adjustmentMode) {
    const formattedValue = expense.adjustmentMode === 'percentage'
      ? `${expense.adjustmentValue}%`
      : formatCurrency(expense.adjustmentValue);
    parts.push(`adjust ${formattedValue} per ${expense.adjustmentFrequency}`);
  }
  return parts.join(' / ');
}

export function computeXIRR(cashflows: { date: Date; amount: number }[]): number | null {
  if (cashflows.length < 2) return null;

  const sorted = [...cashflows].sort((a, b) => a.date.getTime() - b.date.getTime());
  const t0 = sorted[0].date.getTime();
  const msPerYear = 365.25 * 24 * 3600 * 1000;
  const times = sorted.map(cf => (cf.date.getTime() - t0) / msPerYear);
  const amounts = sorted.map(cf => cf.amount);

  const hasNegative = amounts.some(a => a < 0);
  const hasPositive = amounts.some(a => a > 0);
  if (!hasNegative || !hasPositive) return null;

  const npv = (rate: number) =>
    amounts.reduce((sum, amount, i) => sum + amount / Math.pow(1 + rate, times[i]), 0);

  const dnpv = (rate: number) =>
    amounts.reduce((sum, amount, i) => sum - times[i] * amount / Math.pow(1 + rate, times[i] + 1), 0);

  const tryNewton = (initial: number): number | null => {
    let rate = initial;
    for (let i = 0; i < 200; i += 1) {
      const f = npv(rate);
      const df = dnpv(rate);
      if (Math.abs(df) < 1e-12) break;
      const newRate = rate - f / df;
      if (Math.abs(newRate - rate) < 1e-8 && rate > -1) return rate;
      rate = Math.max(newRate, -0.9999);
    }
    return null;
  };

  for (const guess of [0.1, 0.0, -0.1, 0.5, -0.5, -0.9]) {
    const result = tryNewton(guess);
    if (result !== null && result > -1) return result;
  }

  let lowerBound = -0.9999;
  let upperBound = 100;
  if (npv(lowerBound) * npv(upperBound) > 0) return null;
  for (let i = 0; i < 300; i += 1) {
    const midpoint = (lowerBound + upperBound) / 2;
    if (upperBound - lowerBound < 1e-8) return midpoint;
    if (npv(midpoint) * npv(lowerBound) <= 0) {
      upperBound = midpoint;
    } else {
      lowerBound = midpoint;
    }
  }
  return (lowerBound + upperBound) / 2;
}

export function periodEndDate(key: string, zoomLevel: ZoomLevel): Date {
  const parts = key.split('-').map(Number);
  switch (zoomLevel) {
    case 'days':
      return new Date(parts[0], parts[1] - 1, parts[2], 23, 59, 59, 999);
    case 'weeks':
      return new Date(parts[0], parts[1] - 1, parts[2] + 6, 23, 59, 59, 999);
    case 'months':
      return new Date(parts[0], parts[1], 0, 23, 59, 59, 999);
    case 'years':
      return new Date(parts[0], 11, 31, 23, 59, 59, 999);
  }
}

export function formatPercent(rate: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'percent',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(rate);
}

export function sortExpenses(expenses: InvestmentExpense[]): InvestmentExpense[] {
  return [...expenses].sort((left: InvestmentExpense, right: InvestmentExpense) =>
    effectiveEndDate(left).localeCompare(effectiveEndDate(right)) || left.label.localeCompare(right.label)
  );
}
