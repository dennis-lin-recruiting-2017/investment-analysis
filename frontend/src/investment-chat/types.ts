export type ChatInputType = 'text' | 'textarea' | 'number' | 'date' | 'choice' | 'yesno';

export type ChatAnswerValue = string | number | boolean;

export type ChatAnswers = Record<string, ChatAnswerValue>;

export type ChatStep = {
  id: string;
  prompt: string;
  input: ChatInputType;
  helpText?: string;
  placeholder?: string;
  choices?: Array<{ label: string; value: string }>;
  submitLabel?: string;
};

export type ChatHistoryEntry = {
  stepId: string;
  prompt: string;
  answerLabel: string;
};

export type ChatScript = {
  id: string;
  title: string;
  initialStepId: string;
  intro?: string;
  completeLabel?: string;
  resolveStep: (stepId: string, answers: ChatAnswers) => ChatStep | null;
  getNextStepId: (currentStepId: string, answer: ChatAnswerValue, answers: ChatAnswers) => string | null;
  formatAnswer?: (step: ChatStep, answer: ChatAnswerValue) => string;
  buildSummary: (answers: ChatAnswers) => string;
};
