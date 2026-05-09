import type { ChatAnswers, ChatScript } from '../types';

export function createPlaceholderScript(options: {
  id: string;
  title: string;
  intro: string;
  notesHeading: string;
  thesisPrompt: string;
  riskPrompt: string;
  nextPrompt: string;
}): ChatScript {
  return {
    id: options.id,
    title: options.title,
    intro: options.intro,
    initialStepId: 'main_thesis',
    resolveStep: (stepId: string) => {
      switch (stepId) {
        case 'main_thesis':
          return {
            id: stepId,
            prompt: options.thesisPrompt,
            input: 'textarea',
          };
        case 'main_risk':
          return {
            id: stepId,
            prompt: options.riskPrompt,
            input: 'textarea',
          };
        case 'next_question':
          return {
            id: stepId,
            prompt: options.nextPrompt,
            input: 'textarea',
            submitLabel: 'Finish Chat',
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
        options.notesHeading,
        `Thesis: ${answers.main_thesis || '-'}`,
        `Main risk: ${answers.main_risk || '-'}`,
        `Next question: ${answers.next_question || '-'}`,
      ].join('\n'),
  };
}
