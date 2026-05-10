import { useEffect, useMemo, useState, type FormEvent, type SyntheticEvent } from 'react';
import {
  IconButton,
  Alert,
  Box,
  Button,
  ButtonBase,
  CircularProgress,
  Divider,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from '@mui/material';
import AddCircleOutlineIcon from '@mui/icons-material/AddCircleOutline';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import EditOutlinedIcon from '@mui/icons-material/EditOutlined';
import { useParams } from 'react-router-dom';
import {
  createInvestmentCategory,
  deleteInvestmentCategory,
  deleteInvestmentExpense,
  deleteInvestmentSaleAssumption,
  getInvestment,
  saveInvestmentExpense,
  saveInvestmentSaleAssumption,
  updateInvestment,
  updateInvestmentCategory,
  updateInvestmentExpense,
  updateInvestmentSaleAssumption,
  type Investment,
  type InvestmentExpense,
  type InvestmentExpenseInput,
  type InvestmentInput,
  type InvestmentSaleAssumption,
  type InvestmentSaleAssumptionInput,
} from '../lib/api';
import CategoryManagerDialog from '../components/investments/CategoryManagerDialog';
import DeleteCategoryDialog from '../components/investments/DeleteCategoryDialog';
import InvestmentAnalysisDialog from '../components/investments/InvestmentAnalysisDialog';
import InvestmentDetailsDialog from '../components/investments/InvestmentDetailsDialog';
import PaymentFlowDialog from '../components/investments/PaymentFlowDialog';
import SalesAssumptionsDialog from '../components/investments/SalesAssumptionsDialog';
import TimelinePeriodDialog from '../components/investments/TimelinePeriodDialog';
import {
  effectiveEndDate,
  formatCurrency,
  groupExpenses,
  initialExpenseForm,
  initialInvestmentForm,
  recurringSummary,
  sortExpenses,
  type FlowTab,
  type GroupedExpense,
  type TimelineOccurrence,
  type ZoomLevel,
} from '../components/investments/investmentDetailShared';

const initialSaleAssumptionForm: InvestmentSaleAssumptionInput = {
  label: '',
  amount: 0,
  growthType: 'fixed',
  growthPeriod: '',
  growthMode: '',
  growthValue: 0,
  description: '',
  startDate: '',
  endDate: '',
  notes: '',
};

function TimelineChart({
  expenses,
  zoomLevel,
  onZoomChange,
  onGroupSelect,
}: {
  expenses: InvestmentExpense[];
  zoomLevel: ZoomLevel;
  onZoomChange: (zoomLevel: ZoomLevel) => void;
  onGroupSelect: (group: GroupedExpense) => void;
}) {
  const groupedExpenses = useMemo(() => groupExpenses(expenses, zoomLevel), [expenses, zoomLevel]);
  const maxPositiveAmount = useMemo(
    () => groupedExpenses.reduce((highest: number, item: GroupedExpense) => Math.max(highest, item.amount > 0 ? item.amount : 0), 0),
    [groupedExpenses]
  );
  const maxNegativeAmount = useMemo(
    () => groupedExpenses.reduce((highest: number, item: GroupedExpense) => Math.max(highest, item.amount < 0 ? Math.abs(item.amount) : 0), 0),
    [groupedExpenses]
  );

  return (
    <Stack
      spacing={2}
      sx={{
        minHeight: 374,
        maxHeight: 374,
      }}
    >
      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        spacing={2}
        alignItems={{ xs: 'flex-start', sm: 'center' }}
        justifyContent="space-between"
      >
        <div>
          <Typography variant="h6">Payment Timeline</Typography>
          <Typography color="text.secondary">
            Use the controls to switch between days, weeks, months, and years.
          </Typography>
        </div>
        <ToggleButtonGroup
          exclusive
          size="small"
          value={zoomLevel}
          disabled={groupedExpenses.length === 0}
          onChange={(_event, value: ZoomLevel | null) => {
            if (value) {
              onZoomChange(value);
            }
          }}
        >
          <ToggleButton value="days">Days</ToggleButton>
          <ToggleButton value="weeks">Weeks</ToggleButton>
          <ToggleButton value="months">Months</ToggleButton>
          <ToggleButton value="years">Years</ToggleButton>
        </ToggleButtonGroup>
      </Stack>

      <Box
        sx={{
          overflowX: 'auto',
          overflowY: 'hidden',
          border: '1px solid',
          borderColor: 'divider',
          borderRadius: 1,
          p: 2,
          minHeight: 296,
          maxHeight: 296,
        }}
      >
        {groupedExpenses.length === 0 ? (
          <Stack justifyContent="center" alignItems="flex-start" sx={{ minHeight: 280 }}>
            <Typography color="text.secondary">
              Add cash flows below to see them plotted on the timeline.
            </Typography>
          </Stack>
        ) : (
          <Stack direction="row" spacing={2} alignItems="flex-start" sx={{ minHeight: 280, width: 'max-content' }}>
            {groupedExpenses.map((group: GroupedExpense) => {
            const positiveHeight = group.amount > 0 && maxPositiveAmount > 0
              ? Math.max((group.amount / maxPositiveAmount) * 90, 12)
              : 0;
            const negativeHeight = group.amount < 0 && maxNegativeAmount > 0
              ? Math.max((Math.abs(group.amount) / maxNegativeAmount) * 90, 12)
              : 0;
            const labels = Array.from(new Set(group.occurrences.map((occurrence: TimelineOccurrence) => occurrence.expense.label)));
            const caption = labels.join(', ');

            return (
              <Stack key={group.key} spacing={1} alignItems="center" sx={{ width: 96, flexShrink: 0 }}>
                <ButtonBase
                  onClick={() => onGroupSelect(group)}
                  title={caption}
                  sx={{
                    width: '100%',
                    borderRadius: 1,
                    display: 'block',
                    textAlign: 'inherit',
                    p: 0.5,
                  }}
                >
                  <Typography variant="caption" color="text.secondary" textAlign="center">
                    {formatCurrency(group.amount)}
                  </Typography>
                  <Box
                    sx={{
                      width: '100%',
                      height: 220,
                      position: 'relative',
                    }}
                  >
                    <Box
                      sx={{
                        position: 'absolute',
                        left: 0,
                        right: 0,
                        top: '50%',
                        borderTop: '1px solid',
                        borderColor: 'divider',
                      }}
                    />
                    {group.amount > 0 ? (
                      <Box
                        sx={{
                          position: 'absolute',
                          left: '12%',
                          right: '12%',
                          bottom: '50%',
                          height: positiveHeight,
                          minHeight: 12,
                          borderRadius: 1,
                          backgroundColor: 'primary.main',
                        }}
                      />
                    ) : null}
                    {group.amount < 0 ? (
                      <Box
                        sx={{
                          position: 'absolute',
                          left: '12%',
                          right: '12%',
                          top: '50%',
                          height: negativeHeight,
                          minHeight: 12,
                          borderRadius: 1,
                          backgroundColor: 'error.main',
                        }}
                      />
                    ) : null}
                  </Box>
                  <Typography variant="caption" textAlign="center">
                    {group.label}
                  </Typography>
                  <Typography variant="caption" color="text.secondary" textAlign="center">
                    {group.occurrences.length === 1 ? group.occurrences[0].expense.label : `${group.occurrences.length} cash flows`}
                  </Typography>
                </ButtonBase>
              </Stack>
            );
          })}
          </Stack>
        )}
      </Box>
    </Stack>
  );
}

export default function InvestmentDetail() {
  const { uuid } = useParams();
  const [investment, setInvestment] = useState<Investment | null>(null);
  const [investmentForm, setInvestmentForm] = useState<InvestmentInput>(initialInvestmentForm);
  const [initialInvestmentInput, setInitialInvestmentInput] = useState('0');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expenseForm, setExpenseForm] = useState<InvestmentExpenseInput>(initialExpenseForm);
  const [amountInput, setAmountInput] = useState('0');
  const [categories, setCategories] = useState<string[]>([]);
  const [expenseError, setExpenseError] = useState<string | null>(null);
  const [savingExpense, setSavingExpense] = useState(false);
  const [deletingExpenseId, setDeletingExpenseId] = useState<string | null>(null);
  const [zoomLevel, setZoomLevel] = useState<ZoomLevel>('months');
  const [showAddFlow, setShowAddFlow] = useState(false);
  const [showCategoryManager, setShowCategoryManager] = useState(false);
  const [showInvestmentDetails, setShowInvestmentDetails] = useState(false);
  const [savingInvestmentDetails, setSavingInvestmentDetails] = useState(false);
  const [investmentDetailsError, setInvestmentDetailsError] = useState<string | null>(null);
  const [flowTab, setFlowTab] = useState<FlowTab>('one-time');
  const [editingExpenseId, setEditingExpenseId] = useState<string | null>(null);
  const [categoryDraft, setCategoryDraft] = useState('');
  const [editingCategoryName, setEditingCategoryName] = useState<string | null>(null);
  const [categoryError, setCategoryError] = useState<string | null>(null);
  const [savingCategory, setSavingCategory] = useState(false);
  const [pendingDeleteCategory, setPendingDeleteCategory] = useState<string | null>(null);
  const [selectedCategories, setSelectedCategories] = useState<string[]>([]);
  const [selectedTimelineGroup, setSelectedTimelineGroup] = useState<GroupedExpense | null>(null);
  const [showSalesAssumptions, setShowSalesAssumptions] = useState(false);
  const [showInvestmentAnalysis, setShowInvestmentAnalysis] = useState(false);
  const [saleAssumptionForm, setSaleAssumptionForm] = useState<InvestmentSaleAssumptionInput>(initialSaleAssumptionForm);
  const [saleAssumptionAmountInput, setSaleAssumptionAmountInput] = useState('0');
  const [saleAssumptionError, setSaleAssumptionError] = useState<string | null>(null);
  const [savingSaleAssumption, setSavingSaleAssumption] = useState(false);
  const [editingSaleAssumptionId, setEditingSaleAssumptionId] = useState<string | null>(null);
  const [deletingSaleAssumptionId, setDeletingSaleAssumptionId] = useState<string | null>(null);

  useEffect(() => {
    if (!uuid) {
      setError('Invalid investment id.');
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);
    getInvestment(uuid)
      .then((result: Investment) => {
      setInvestment({
          ...result,
          expenses: sortExpenses(result.expenses || []),
        });
        setInvestmentForm({
          name: result.name,
          ticker: result.ticker,
          assetClass: result.assetClass,
          purchasePrice: result.purchasePrice || 0,
          couponRate: result.couponRate || 0,
          maturityDate: result.maturityDate || '',
          callableDateStart: result.callableDateStart || '',
          callPrice: result.callPrice || 0,
          callDate: result.callDate || '',
          thesis: result.thesis,
          targetAllocation: result.targetAllocation,
          initialInvestment: result.initialInvestment || 0,
          initialInvestmentDate: result.initialInvestmentDate || '',
          notes: result.notes,
        });
        setInitialInvestmentInput(String(result.initialInvestment || 0));
        setCategories(result.categories || []);
      })
      .catch((err: unknown) => setError(err instanceof Error ? err.message : 'Failed to load investment'))
      .finally(() => setLoading(false));
  }, [uuid]);

  const handleExpenseChange = <K extends keyof InvestmentExpenseInput>(
    key: K,
    value: InvestmentExpenseInput[K]
  ) => {
    setExpenseError(null);
    setExpenseForm((current: InvestmentExpenseInput) => ({ ...current, [key]: value }));
  };

  const handleInvestmentChange = <K extends keyof InvestmentInput>(
    key: K,
    value: InvestmentInput[K]
  ) => {
    setInvestmentDetailsError(null);
    setInvestmentForm((current: InvestmentInput) => ({ ...current, [key]: value }));
  };

  const handleSaleAssumptionChange = <K extends keyof InvestmentSaleAssumptionInput>(
    key: K,
    value: InvestmentSaleAssumptionInput[K]
  ) => {
    setSaleAssumptionError(null);
    setSaleAssumptionForm((current: InvestmentSaleAssumptionInput) => ({ ...current, [key]: value }));
  };

  const resetDialog = () => {
    setShowAddFlow(false);
    setFlowTab('one-time');
    setEditingExpenseId(null);
    setExpenseForm(initialExpenseForm);
    setAmountInput('0');
    setExpenseError(null);
  };

  const resetCategoryEditor = () => {
    setEditingCategoryName(null);
    setCategoryDraft('');
    setCategoryError(null);
  };

  const resetSaleAssumptionDialog = () => {
    setShowSalesAssumptions(false);
    setSaleAssumptionForm(initialSaleAssumptionForm);
    setSaleAssumptionAmountInput('0');
    setSaleAssumptionError(null);
    setEditingSaleAssumptionId(null);
  };

  const openSalesAssumptionsDialog = () => {
    setSaleAssumptionError(null);
    setShowSalesAssumptions(true);
  };

  const openInvestmentDetailsDialog = () => {
    if (!investment) {
      return;
    }
    setInvestmentDetailsError(null);
    setInvestmentForm({
      name: investment.name,
      ticker: investment.ticker,
      assetClass: investment.assetClass,
      purchasePrice: investment.purchasePrice || 0,
      couponRate: investment.couponRate || 0,
      maturityDate: investment.maturityDate || '',
      callableDateStart: investment.callableDateStart || '',
      callPrice: investment.callPrice || 0,
      callDate: investment.callDate || '',
      thesis: investment.thesis,
      targetAllocation: investment.targetAllocation,
      initialInvestment: investment.initialInvestment || 0,
      initialInvestmentDate: investment.initialInvestmentDate || '',
      notes: investment.notes,
    });
    setInitialInvestmentInput(String(investment.initialInvestment || 0));
    setShowInvestmentDetails(true);
  };

  const handleInvestmentDetailsSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!investment) {
      return;
    }

    const payload: InvestmentInput = {
      ...investmentForm,
      initialInvestment: Number(initialInvestmentInput),
    };

    if (!Number.isFinite(payload.initialInvestment)) {
      setInvestmentDetailsError('Initial investment must be a valid number.');
      return;
    }

    setSavingInvestmentDetails(true);
    setInvestmentDetailsError(null);
    try {
      const updated = await updateInvestment(investment.uuid, payload);
      setInvestment((current: Investment | null) => ({
        ...updated,
        expenses: updated.expenses?.length ? sortExpenses(updated.expenses) : (current?.expenses || []),
        categories: updated.categories?.length ? updated.categories : (current?.categories || []),
        saleAssumptions: updated.saleAssumptions?.length ? updated.saleAssumptions : (current?.saleAssumptions || []),
      }));
      setInvestmentForm({
        name: updated.name,
        ticker: updated.ticker,
        assetClass: updated.assetClass,
        purchasePrice: updated.purchasePrice || 0,
        couponRate: updated.couponRate || 0,
        maturityDate: updated.maturityDate || '',
        callableDateStart: updated.callableDateStart || '',
        callPrice: updated.callPrice || 0,
        callDate: updated.callDate || '',
        thesis: updated.thesis,
        targetAllocation: updated.targetAllocation,
        initialInvestment: updated.initialInvestment || 0,
        initialInvestmentDate: updated.initialInvestmentDate || '',
        notes: updated.notes,
      });
      setInitialInvestmentInput(String(updated.initialInvestment || 0));
      setShowInvestmentDetails(false);
    } catch (err) {
      setInvestmentDetailsError(err instanceof Error ? err.message : 'Failed to save investment details');
    } finally {
      setSavingInvestmentDetails(false);
    }
  };

  const openCreateDialog = () => {
    setFlowTab('one-time');
    setEditingExpenseId(null);
    setExpenseForm(initialExpenseForm);
    setAmountInput('0');
    setExpenseError(null);
    setShowAddFlow(true);
  };

  const openEditDialog = (expense: InvestmentExpense) => {
    setFlowTab(expense.flowType);
    setEditingExpenseId(expense.id);
    setExpenseForm({
      eventType: expense.eventType,
      flowType: expense.flowType,
      recurrenceInterval: expense.recurrenceInterval,
      label: expense.label,
      amount: expense.amount,
      description: expense.description,
      startDate: expense.startDate,
      endDate: expense.endDate,
      category: expense.category,
      adjustmentFrequency: expense.adjustmentFrequency,
      adjustmentMode: expense.adjustmentMode,
      adjustmentValue: expense.adjustmentValue,
      notes: expense.notes,
    });
    setAmountInput(String(expense.amount));
    setExpenseError(null);
    setShowAddFlow(true);
  };

  const handleFlowTabChange = (_event: SyntheticEvent, value: FlowTab) => {
    if (!value) {
      return;
    }
    setFlowTab(value);
    setExpenseError(null);
    setExpenseForm((current: InvestmentExpenseInput) => ({
      ...current,
      flowType: value,
      recurrenceInterval: value === 'one-time' ? '' : (current.recurrenceInterval || 'monthly'),
      startDate: value === 'one-time' && !current.startDate ? current.endDate : current.startDate,
      adjustmentFrequency: value === 'one-time' ? '' : current.adjustmentFrequency,
      adjustmentMode: value === 'one-time' ? '' : current.adjustmentMode,
      adjustmentValue: value === 'one-time' ? 0 : current.adjustmentValue,
    }));
  };

  const handleExpenseSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!investment) {
      return;
    }

    const payload: InvestmentExpenseInput = {
      ...expenseForm,
      amount: Number(amountInput),
      flowType: flowTab,
      recurrenceInterval: flowTab === 'recurring' ? expenseForm.recurrenceInterval : '',
      startDate: flowTab === 'one-time' ? (expenseForm.endDate || expenseForm.startDate) : expenseForm.startDate,
      adjustmentFrequency: flowTab === 'recurring' ? expenseForm.adjustmentFrequency : '',
      adjustmentMode: flowTab === 'recurring' ? expenseForm.adjustmentMode : '',
      adjustmentValue: flowTab === 'recurring' ? expenseForm.adjustmentValue : 0,
    };

    if (!Number.isFinite(payload.amount)) {
      setExpenseError('Amount must be a valid number.');
      return;
    }

    setSavingExpense(true);
    setExpenseError(null);

    try {
      const savedExpense = editingExpenseId === null
        ? await saveInvestmentExpense(investment.uuid, payload)
        : await updateInvestmentExpense(investment.uuid, editingExpenseId, payload);
      setInvestment((current: Investment | null) =>
        current
          ? {
              ...current,
              expenses: sortExpenses(
                editingExpenseId === null
                  ? [...(current.expenses || []), savedExpense]
                  : (current.expenses || []).map((expense: InvestmentExpense) =>
                      expense.id === editingExpenseId ? savedExpense : expense
                    )
              ),
            }
          : current
      );
      resetDialog();
    } catch (err) {
      setExpenseError(err instanceof Error ? err.message : 'Failed to save cash flow');
    } finally {
      setSavingExpense(false);
    }
  };

  const handleDeleteExpense = async (expense: InvestmentExpense) => {
    if (!investment) {
      return;
    }
    setDeletingExpenseId(expense.id);
    setExpenseError(null);
    try {
      await deleteInvestmentExpense(investment.uuid, expense.id);
      setInvestment((current: Investment | null) =>
        current
          ? {
              ...current,
              expenses: sortExpenses((current.expenses || []).filter((item: InvestmentExpense) => item.id !== expense.id)),
            }
          : current
      );
      if (editingExpenseId === expense.id) {
        resetDialog();
      }
    } catch (err) {
      setExpenseError(err instanceof Error ? err.message : 'Failed to delete cash flow');
    } finally {
      setDeletingExpenseId(null);
    }
  };

  const openEditSaleAssumption = (assumption: InvestmentSaleAssumption) => {
    setEditingSaleAssumptionId(assumption.id);
    setSaleAssumptionForm({
      label: assumption.label,
      amount: assumption.amount,
      growthType: assumption.growthType,
      growthPeriod: assumption.growthPeriod,
      growthMode: assumption.growthMode,
      growthValue: assumption.growthValue,
      description: assumption.description,
      startDate: assumption.startDate,
      endDate: assumption.endDate,
      notes: assumption.notes,
    });
    setSaleAssumptionAmountInput(String(assumption.amount));
    setSaleAssumptionError(null);
    setShowSalesAssumptions(true);
  };

  const handleSaleAssumptionSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!investment) {
      return;
    }

    const payload: InvestmentSaleAssumptionInput = {
      ...saleAssumptionForm,
      amount: Number(saleAssumptionAmountInput),
    };
    if (!Number.isFinite(payload.amount)) {
      setSaleAssumptionError('Amount must be a valid number.');
      return;
    }

    setSavingSaleAssumption(true);
    setSaleAssumptionError(null);
    try {
      const saved = editingSaleAssumptionId === null
        ? await saveInvestmentSaleAssumption(investment.uuid, payload)
        : await updateInvestmentSaleAssumption(investment.uuid, editingSaleAssumptionId, payload);
      setInvestment((current: Investment | null) =>
        current
          ? {
              ...current,
              saleAssumptions: editingSaleAssumptionId === null
                ? [...(current.saleAssumptions || []), saved]
                : (current.saleAssumptions || []).map((assumption: InvestmentSaleAssumption) =>
                    assumption.id === editingSaleAssumptionId ? saved : assumption
                  ),
            }
          : current
      );
      setSaleAssumptionForm(initialSaleAssumptionForm);
      setSaleAssumptionAmountInput('0');
      setEditingSaleAssumptionId(null);
    } catch (err) {
      setSaleAssumptionError(err instanceof Error ? err.message : 'Failed to save sale assumption');
    } finally {
      setSavingSaleAssumption(false);
    }
  };

  const handleDeleteSaleAssumption = async (assumption: InvestmentSaleAssumption) => {
    if (!investment) {
      return;
    }
    setDeletingSaleAssumptionId(assumption.id);
    setSaleAssumptionError(null);
    try {
      await deleteInvestmentSaleAssumption(investment.uuid, assumption.id);
      setInvestment((current: Investment | null) =>
        current
          ? {
              ...current,
              saleAssumptions: (current.saleAssumptions || []).filter((item: InvestmentSaleAssumption) => item.id !== assumption.id),
            }
          : current
      );
      if (editingSaleAssumptionId === assumption.id) {
        setSaleAssumptionForm(initialSaleAssumptionForm);
        setSaleAssumptionAmountInput('0');
        setEditingSaleAssumptionId(null);
      }
    } catch (err) {
      setSaleAssumptionError(err instanceof Error ? err.message : 'Failed to delete sale assumption');
    } finally {
      setDeletingSaleAssumptionId(null);
    }
  };

  const handleCategorySubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!investment) {
      return;
    }
    setSavingCategory(true);
    setCategoryError(null);
    try {
      const trimmedName = categoryDraft.trim();
      const nextCategories = editingCategoryName === null
        ? await createInvestmentCategory(investment.uuid, trimmedName)
        : await updateInvestmentCategory(investment.uuid, editingCategoryName, trimmedName);
      setCategories(nextCategories);
      setInvestment((current: Investment | null) =>
        current
          ? {
              ...current,
              categories: nextCategories,
              expenses: editingCategoryName === null
                ? current.expenses
                : (current.expenses || []).map((expense: InvestmentExpense) =>
                    expense.category === editingCategoryName ? { ...expense, category: trimmedName } : expense
                  ),
            }
          : current
      );
      setExpenseForm((current: InvestmentExpenseInput) => ({
        ...current,
        category: editingCategoryName !== null && current.category === editingCategoryName ? trimmedName : current.category,
      }));
      resetCategoryEditor();
    } catch (err) {
      setCategoryError(err instanceof Error ? err.message : 'Failed to save category');
    } finally {
      setSavingCategory(false);
    }
  };

  const handleConfirmDeleteCategory = async () => {
    if (!investment || !pendingDeleteCategory) {
      return;
    }

    setSavingCategory(true);
    setCategoryError(null);
    try {
      const categoryToDelete = pendingDeleteCategory;
      const nextCategories = await deleteInvestmentCategory(investment.uuid, categoryToDelete);
      setCategories(nextCategories);
      setInvestment((current: Investment | null) =>
        current
          ? {
              ...current,
              categories: nextCategories,
              expenses: (current.expenses || []).map((expense: InvestmentExpense) =>
                expense.category === categoryToDelete ? { ...expense, category: '' } : expense
              ),
            }
          : current
      );
      setExpenseForm((current: InvestmentExpenseInput) => ({
        ...current,
        category: current.category === categoryToDelete ? '' : current.category,
      }));
      if (editingCategoryName === categoryToDelete) {
        resetCategoryEditor();
      }
      setPendingDeleteCategory(null);
    } catch (err) {
      setCategoryError(err instanceof Error ? err.message : 'Failed to delete category');
    } finally {
      setSavingCategory(false);
    }
  };

  if (loading) {
    return (
      <Stack spacing={2}>
        <Typography variant="h4">Investment</Typography>
        <Stack direction="row" spacing={2} alignItems="center">
          <CircularProgress size={24} />
          <Typography>Loading investment...</Typography>
        </Stack>
      </Stack>
    );
  }

  if (error || !investment) {
    return (
      <Stack spacing={2}>
        <Typography variant="h4">Investment</Typography>
        <Alert severity="error">{error || 'Investment not found.'}</Alert>
      </Stack>
    );
  }

  const expenses = investment.expenses || [];
  const filteredExpenses = selectedCategories.length === 0
    ? expenses
    : expenses.filter((expense: InvestmentExpense) => {
        const category = expense.category || 'Uncategorized';
        return selectedCategories.includes(category);
      });
  const categoryOptions = ['Uncategorized', ...categories];

  return (
    <Stack
      spacing={2}
      sx={{
        height: 'calc(100vh - 112px)',
        maxHeight: 'calc(100vh - 112px)',
        overflow: 'hidden',
      }}
    >
      <div>
        <Typography variant="h4" gutterBottom>
          {investment.name}
        </Typography>
        <Typography color="text.secondary">
          {investment.assetClass}
          {investment.ticker ? ` / ${investment.ticker}` : ''}
        </Typography>
      </div>

      <PaymentFlowDialog
        open={showAddFlow}
        saving={savingExpense}
        error={expenseError}
        editingExpenseId={editingExpenseId}
        flowTab={flowTab}
        expenseForm={expenseForm}
        amountInput={amountInput}
        categories={categories}
        onClose={resetDialog}
        onSubmit={handleExpenseSubmit}
        onTabChange={handleFlowTabChange}
        onExpenseChange={handleExpenseChange}
        setAmountInput={setAmountInput}
        clearError={() => setExpenseError(null)}
      />

      <InvestmentDetailsDialog
        open={showInvestmentDetails}
        saving={savingInvestmentDetails}
        error={investmentDetailsError}
        form={investmentForm}
        initialInvestmentInput={initialInvestmentInput}
        onClose={() => setShowInvestmentDetails(false)}
        onSubmit={handleInvestmentDetailsSubmit}
        onChange={handleInvestmentChange}
        setInitialInvestmentInput={setInitialInvestmentInput}
        clearError={() => setInvestmentDetailsError(null)}
      />

      <InvestmentAnalysisDialog
        open={showInvestmentAnalysis}
        onClose={() => setShowInvestmentAnalysis(false)}
      />

      <TimelinePeriodDialog
        open={selectedTimelineGroup !== null}
        group={selectedTimelineGroup}
        zoomLevel={zoomLevel}
        allExpenses={investment.expenses || []}
        initialInvestment={investment.initialInvestment}
        initialInvestmentDate={investment.initialInvestmentDate}
        onClose={() => setSelectedTimelineGroup(null)}
      />

      <SalesAssumptionsDialog
        open={showSalesAssumptions}
        saving={savingSaleAssumption}
        deletingId={deletingSaleAssumptionId}
        editingId={editingSaleAssumptionId}
        error={saleAssumptionError}
        form={saleAssumptionForm}
        amountInput={saleAssumptionAmountInput}
        assumptions={investment.saleAssumptions || []}
        onClose={resetSaleAssumptionDialog}
        onSubmit={handleSaleAssumptionSubmit}
        onChange={handleSaleAssumptionChange}
        onAmountInputChange={setSaleAssumptionAmountInput}
        onEdit={openEditSaleAssumption}
        onDelete={(assumption: InvestmentSaleAssumption) => void handleDeleteSaleAssumption(assumption)}
        clearError={() => setSaleAssumptionError(null)}
      />

      <CategoryManagerDialog
        open={showCategoryManager}
        saving={savingCategory}
        error={categoryError}
        categories={categories}
        categoryDraft={categoryDraft}
        editingCategoryName={editingCategoryName}
        onClose={() => setShowCategoryManager(false)}
        onSubmit={handleCategorySubmit}
        setCategoryDraft={setCategoryDraft}
        clearError={() => setCategoryError(null)}
        onStartEdit={(category: string) => {
          setEditingCategoryName(category);
          setCategoryDraft(category);
          setCategoryError(null);
        }}
        onCancelEdit={resetCategoryEditor}
        onDelete={setPendingDeleteCategory}
      />

      <DeleteCategoryDialog
        open={pendingDeleteCategory !== null}
        saving={savingCategory}
        categoryName={pendingDeleteCategory}
        onClose={() => setPendingDeleteCategory(null)}
        onConfirm={() => void handleConfirmDeleteCategory()}
      />

      <Paper variant="outlined" sx={{ p: 2, minHeight: 0, overflowY: 'auto', flex: 1 }}>
        <Stack spacing={2}>
          <TimelineChart
            expenses={filteredExpenses}
            zoomLevel={zoomLevel}
            onZoomChange={setZoomLevel}
            onGroupSelect={setSelectedTimelineGroup}
          />

          <Divider />

          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
            <Button
              variant="outlined"
              size="small"
              onClick={openInvestmentDetailsDialog}
            >
              Investment Details
            </Button>
            <Button
              variant="outlined"
              size="small"
              onClick={() => setShowInvestmentAnalysis(true)}
            >
              Investment Analysis
            </Button>
            <Button
              variant="outlined"
              size="small"
              onClick={openSalesAssumptionsDialog}
            >
              Sales Assumptions
            </Button>
            <Button
              variant="outlined"
              size="small"
              onClick={() => {
                resetCategoryEditor();
                setShowCategoryManager(true);
              }}
            >
              Manage Categories
            </Button>
          </Stack>

          <Stack spacing={1}>
            <Typography variant="subtitle2" color="text.secondary">Filter Categories</Typography>
            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
              {categoryOptions.map((category: string) => {
                const active = selectedCategories.includes(category);
                return (
                  <Button
                    key={category}
                    size="small"
                    variant={active ? 'contained' : 'outlined'}
                    onClick={() =>
                      setSelectedCategories((current: string[]) =>
                        current.includes(category)
                          ? current.filter((item: string) => item !== category)
                          : [...current, category]
                      )
                    }
                  >
                    {category}
                  </Button>
                );
              })}
              {selectedCategories.length > 0 ? (
                <Button size="small" onClick={() => setSelectedCategories([])}>
                  Clear Filters
                </Button>
              ) : null}
            </Stack>
          </Stack>

          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} justifyContent="space-between" alignItems={{ xs: 'flex-start', sm: 'center' }}>
            <Typography variant="h6">Payment Flow</Typography>
            <Button
              variant="outlined"
              size="small"
              startIcon={<AddCircleOutlineIcon />}
              onClick={openCreateDialog}
            >
              Add Cash Flow
            </Button>
          </Stack>

          <Box sx={{ overflowX: 'auto' }}>
            <Table size="small" sx={{ minWidth: 900 }}>
              <TableHead>
                <TableRow>
                  <TableCell sx={{ width: 180 }}>Amount</TableCell>
                  <TableCell>Name</TableCell>
                  <TableCell>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <span>Category</span>
                      <Button
                        size="small"
                        onClick={() => {
                          resetCategoryEditor();
                          setShowCategoryManager(true);
                        }}
                      >
                        Manage
                      </Button>
                    </Stack>
                  </TableCell>
                  <TableCell>Description</TableCell>
                  <TableCell>Start Date</TableCell>
                  <TableCell>End Date</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filteredExpenses.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7}>
                      <Typography color="text.secondary">
                        {expenses.length === 0 ? 'No payment flows yet.' : 'No payment flows match the selected categories.'}
                      </Typography>
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredExpenses.map((expense: InvestmentExpense) => (
                    <TableRow key={expense.id} hover>
                      <TableCell>
                        <Stack spacing={0.5}>
                          <Typography>{formatCurrency(expense.amount)}</Typography>
                          <Typography variant="caption" color="text.secondary">
                            {recurringSummary(expense)}
                          </Typography>
                        </Stack>
                      </TableCell>
                      <TableCell>{expense.label}</TableCell>
                      <TableCell>{expense.category || 'Uncategorized'}</TableCell>
                      <TableCell>{expense.description || expense.notes || 'None'}</TableCell>
                      <TableCell>{expense.startDate || '-'}</TableCell>
                      <TableCell>{effectiveEndDate(expense) || '-'}</TableCell>
                      <TableCell align="right">
                        <Stack direction="row" spacing={1} justifyContent="flex-end">
                          <IconButton
                            size="small"
                            aria-label={`Edit ${expense.label}`}
                            onClick={() => openEditDialog(expense)}
                            disabled={deletingExpenseId === expense.id}
                          >
                            <EditOutlinedIcon fontSize="small" />
                          </IconButton>
                          <IconButton
                            size="small"
                            color="error"
                            aria-label={`Delete ${expense.label}`}
                            onClick={() => void handleDeleteExpense(expense)}
                            disabled={deletingExpenseId === expense.id}
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
      </Paper>
    </Stack>
  );
}
