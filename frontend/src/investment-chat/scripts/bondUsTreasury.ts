import type { ChatAnswers, ChatScript, ChatStep } from '../types';

function choiceLabel(step: ChatStep, value: unknown): string {
  return step.choices?.find((choice) => choice.value === value)?.label || String(value || '-');
}

const steps: Record<string, ChatStep> = {
  thesis_anchor: {
    id: 'thesis_anchor',
    prompt: 'What is the main reason this Treasury belongs in the portfolio right now?',
    input: 'textarea',
  },
  portfolio_role: {
    id: 'portfolio_role',
    prompt: 'What role should this Treasury position play in the portfolio?',
    input: 'choice',
    choices: [
      { label: 'Liquidity reserve', value: 'liquidity_reserve' },
      { label: 'Duration expression', value: 'duration_expression' },
      { label: 'Capital preservation', value: 'capital_preservation' },
      { label: 'Income anchor', value: 'income_anchor' },
      { label: 'Ladder position', value: 'ladder_position' },
    ],
  },
  duration_view: {
    id: 'duration_view',
    prompt: 'What is your duration or rate view behind this purchase?',
    input: 'textarea',
  },
  expected_holding_period: {
    id: 'expected_holding_period',
    prompt: 'How long do you expect to hold it?',
    input: 'choice',
    choices: [
      { label: 'Less than 1 year', value: 'lt_1_year' },
      { label: '1 to 3 years', value: '1_to_3_years' },
      { label: 'To maturity', value: 'to_maturity' },
    ],
  },
  liquidity_need: {
    id: 'liquidity_need',
    prompt: 'Could this position need to be sold early for liquidity?',
    input: 'yesno',
  },
  liquidity_context: {
    id: 'liquidity_context',
    prompt: 'What would most likely trigger an early sale?',
    input: 'textarea',
  },
  reinvestment_risk: {
    id: 'reinvestment_risk',
    prompt: 'What reinvestment or rollover risk matters most here?',
    input: 'textarea',
  },
  review_trigger: {
    id: 'review_trigger',
    prompt: 'What market move or portfolio change would make you revisit this Treasury first?',
    input: 'textarea',
    submitLabel: 'Finish Chat',
  },
};

export const bondUsTreasuryScript: ChatScript = {
  id: 'bond-us-treasury',
  title: 'Bond - US Treasury Chat',
  completeLabel: 'Finish Chat',
  initialStepId: 'thesis_anchor',
  intro: 'Let’s capture the portfolio role, rate view, expected holding period, and the conditions that would change your mind on this Treasury.',
  resolveStep: (stepId: string) => steps[stepId] || null,
  getNextStepId: (currentStepId: string, answer: string | number | boolean) => {
    switch (currentStepId) {
      case 'thesis_anchor':
        return 'portfolio_role';
      case 'portfolio_role':
        return 'duration_view';
      case 'duration_view':
        return 'expected_holding_period';
      case 'expected_holding_period':
        return 'liquidity_need';
      case 'liquidity_need':
        return answer ? 'liquidity_context' : 'reinvestment_risk';
      case 'liquidity_context':
        return 'reinvestment_risk';
      case 'reinvestment_risk':
        return 'review_trigger';
      case 'review_trigger':
        return null;
      default:
        return null;
    }
  },
  buildSummary: (answers: ChatAnswers) => {
    const lines = [
      'Treasury Chat Notes',
      `Reason in portfolio: ${answers.thesis_anchor || '-'}`,
      `Portfolio role: ${choiceLabel(steps.portfolio_role, answers.portfolio_role)}`,
      `Duration/rate view: ${answers.duration_view || '-'}`,
      `Expected holding period: ${choiceLabel(steps.expected_holding_period, answers.expected_holding_period)}`,
      `May need early sale: ${answers.liquidity_need ? 'Yes' : 'No'}`,
    ];

    if (answers.liquidity_context) {
      lines.push(`Early sale trigger: ${answers.liquidity_context}`);
    }

    lines.push(`Reinvestment risk: ${answers.reinvestment_risk || '-'}`);
    lines.push(`Review trigger: ${answers.review_trigger || '-'}`);
    return lines.join('\n');
  },
};
