import type { ChatAnswerValue, ChatAnswers, ChatHistoryEntry, ChatScript, ChatStep } from './types';

export function defaultFormatAnswer(step: ChatStep, answer: ChatAnswerValue): string {
  if (step.input === 'yesno') {
    return answer ? 'Yes' : 'No';
  }

  if (step.input === 'choice' && step.choices) {
    const match = step.choices.find((choice) => choice.value === answer);
    return match?.label || String(answer);
  }

  return String(answer);
}

export function recordAnswer(
  script: ChatScript,
  step: ChatStep,
  answer: ChatAnswerValue,
  answers: ChatAnswers,
  history: ChatHistoryEntry[]
): {
  answers: ChatAnswers;
  history: ChatHistoryEntry[];
  nextStepId: string | null;
} {
  const nextAnswers = { ...answers, [step.id]: answer };
  const formatter = script.formatAnswer || defaultFormatAnswer;
  const nextHistory = [
    ...history,
    {
      stepId: step.id,
      prompt: step.prompt,
      answerLabel: formatter(step, answer),
    },
  ];

  return {
    answers: nextAnswers,
    history: nextHistory,
    nextStepId: script.getNextStepId(step.id, answer, nextAnswers),
  };
}
