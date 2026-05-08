export type Provider = 'lm-studio' | 'ollama' | 'external-endpoint';

export type LLMSettings = {
  provider: Provider;
  endpoint: string;
  model: string;
  apiKey: string;
  temperature: number;
  systemPrompt: string;
};

export type Investment = {
  uuid: string;
  name: string;
  ticker: string;
  assetClass: string;
  purchasePrice: number;
  couponRate: number;
  maturityDate: string;
  callableDateStart: string;
  callPrice: number;
  callDate: string;
  thesis: string;
  targetAllocation: string;
  initialInvestment: number;
  initialInvestmentDate: string;
  notes: string;
  categories?: string[];
  expenses?: InvestmentExpense[];
  saleAssumptions?: InvestmentSaleAssumption[];
  createdAt?: string;
  updatedAt?: string;
};

export type InvestmentSaleAssumption = {
  id: number;
  investmentUuid?: string;
  label: string;
  amount: number;
  growthType: 'fixed' | 'growing';
  growthPeriod: '' | 'month' | 'year';
  growthMode: '' | 'amount' | 'percentage';
  growthValue: number;
  description: string;
  startDate: string;
  endDate: string;
  notes: string;
  createdAt?: string;
  updatedAt?: string;
};

export type InvestmentExpense = {
  id: number;
  investmentUuid?: string;
  eventType: 'cash-flow' | 'deferred-tax';
  flowType: 'one-time' | 'recurring';
  recurrenceInterval: '' | 'daily' | 'weekly' | 'monthly' | 'annually';
  label: string;
  amount: number;
  description: string;
  startDate: string;
  endDate: string;
  dueDate: string;
  category: string;
  adjustmentFrequency: '' | 'month' | 'year';
  adjustmentMode: '' | 'amount' | 'percentage';
  adjustmentValue: number;
  notes: string;
  createdAt?: string;
  updatedAt?: string;
};

export type InvestmentInput = Omit<Investment, 'uuid' | 'expenses' | 'createdAt' | 'updatedAt'>;
export type InvestmentExpenseInput = Omit<InvestmentExpense, 'id' | 'investmentUuid' | 'createdAt' | 'updatedAt' | 'dueDate'>;
export type InvestmentSaleAssumptionInput = Omit<InvestmentSaleAssumption, 'id' | 'investmentUuid' | 'createdAt' | 'updatedAt'>;

declare global {
  interface Window {
    APP_CONFIG?: {
      apiBase?: string;
    };
  }
}

const apiBase = (window.APP_CONFIG?.apiBase || '/api').replace(/\/$/, '');

async function parseResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Request failed with ${response.status}`);
  }
  return response.json() as Promise<T>;
}

export async function getSettings(provider: Provider): Promise<LLMSettings> {
  const response = await fetch(`${apiBase}/settings/${provider}`);
  return parseResponse<LLMSettings>(response);
}

export async function listSettings(): Promise<LLMSettings[]> {
  const response = await fetch(`${apiBase}/settings`);
  return parseResponse<LLMSettings[]>(response);
}

export async function saveSettings(provider: Provider, payload: LLMSettings): Promise<LLMSettings> {
  const response = await fetch(`${apiBase}/settings/${provider}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return parseResponse<LLMSettings>(response);
}

export async function deleteSettings(provider: Provider): Promise<void> {
  const response = await fetch(`${apiBase}/settings/${provider}`, { method: 'DELETE' });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Delete failed with ${response.status}`);
  }
}

export async function listInvestments(): Promise<Investment[]> {
  const response = await fetch(`${apiBase}/investments`);
  return parseResponse<Investment[]>(response);
}

export async function getInvestment(uuid: string): Promise<Investment> {
  const response = await fetch(`${apiBase}/investments/${uuid}`);
  return parseResponse<Investment>(response);
}

export async function saveInvestment(payload: InvestmentInput): Promise<Investment> {
  const response = await fetch(`${apiBase}/investments`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return parseResponse<Investment>(response);
}

export async function updateInvestment(uuid: string, payload: InvestmentInput): Promise<Investment> {
  const response = await fetch(`${apiBase}/investments/${uuid}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return parseResponse<Investment>(response);
}

export async function deleteInvestment(uuid: string): Promise<void> {
  const response = await fetch(`${apiBase}/investments/${uuid}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Delete failed with ${response.status}`);
  }
}

export async function listInvestmentExpenses(uuid: string): Promise<InvestmentExpense[]> {
  const response = await fetch(`${apiBase}/investments/${uuid}/expenses`);
  return parseResponse<InvestmentExpense[]>(response);
}

export async function saveInvestmentExpense(uuid: string, payload: InvestmentExpenseInput): Promise<InvestmentExpense> {
  const response = await fetch(`${apiBase}/investments/${uuid}/expenses`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return parseResponse<InvestmentExpense>(response);
}

export async function updateInvestmentExpense(uuid: string, expenseId: number, payload: InvestmentExpenseInput): Promise<InvestmentExpense> {
  const response = await fetch(`${apiBase}/investments/${uuid}/expenses/${expenseId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return parseResponse<InvestmentExpense>(response);
}

export async function deleteInvestmentExpense(uuid: string, expenseId: number): Promise<void> {
  const response = await fetch(`${apiBase}/investments/${uuid}/expenses/${expenseId}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Delete failed with ${response.status}`);
  }
}

export async function listInvestmentCategories(uuid: string): Promise<string[]> {
  const response = await fetch(`${apiBase}/investments/${uuid}/categories`);
  return parseResponse<string[]>(response);
}

export async function createInvestmentCategory(uuid: string, name: string): Promise<string[]> {
  const response = await fetch(`${apiBase}/investments/${uuid}/categories`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  });
  return parseResponse<string[]>(response);
}

export async function updateInvestmentCategory(uuid: string, currentName: string, name: string): Promise<string[]> {
  const response = await fetch(`${apiBase}/investments/${uuid}/categories/${encodeURIComponent(currentName)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  });
  return parseResponse<string[]>(response);
}

export async function deleteInvestmentCategory(uuid: string, name: string): Promise<string[]> {
  const response = await fetch(`${apiBase}/investments/${uuid}/categories/${encodeURIComponent(name)}`, {
    method: 'DELETE',
  });
  return parseResponse<string[]>(response);
}

export async function saveInvestmentSaleAssumption(uuid: string, payload: InvestmentSaleAssumptionInput): Promise<InvestmentSaleAssumption> {
  const response = await fetch(`${apiBase}/investments/${uuid}/sale-assumptions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return parseResponse<InvestmentSaleAssumption>(response);
}

export async function updateInvestmentSaleAssumption(uuid: string, assumptionId: number, payload: InvestmentSaleAssumptionInput): Promise<InvestmentSaleAssumption> {
  const response = await fetch(`${apiBase}/investments/${uuid}/sale-assumptions/${assumptionId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return parseResponse<InvestmentSaleAssumption>(response);
}

export async function deleteInvestmentSaleAssumption(uuid: string, assumptionId: number): Promise<void> {
  const response = await fetch(`${apiBase}/investments/${uuid}/sale-assumptions/${assumptionId}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Delete failed with ${response.status}`);
  }
}
