import type { ChatAnswers, ChatScript } from '../types';

export const genericInvestmentScript: ChatScript = {
  id: 'generic-investment',
  title: 'Investment Chat',
  intro: 'Let’s capture a little more context while the investment is fresh.',
  initialStepId: 'main_thesis',
  resolveStep: (stepId: string) => {
    switch (stepId) {
      case 'main_thesis':
        return {
          id: stepId,
          prompt: 'What is the strongest part of the thesis that is not already written down?',
          input: 'textarea',
        };
      case 'main_risk':
        return {
          id: stepId,
          prompt: 'What is the main risk or invalidation trigger to track?',
          input: 'textarea',
        };
      case 'next_question':
        return {
          id: stepId,
          prompt: 'What should the next round of analysis focus on?',
          input: 'textarea',
        };
      default:
        return null;
    }
  },
  getNextStepId: (currentStepId: string) => {
    if (currentStepId === 'main_thesis') return 'main_risk';
    if (currentStepId === 'main_risk') return 'next_question';
    return null;
  },
  buildSummary: (answers: ChatAnswers) =>
    [
      'Investment Chat Notes',
      `Thesis: ${answers.main_thesis || '-'}`,
      `Main risk: ${answers.main_risk || '-'}`,
      `Next question: ${answers.next_question || '-'}`,
    ].join('\n'),
};
