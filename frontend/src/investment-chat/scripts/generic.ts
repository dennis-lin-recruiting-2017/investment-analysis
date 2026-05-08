import { createPlaceholderScript } from './createPlaceholderScript';

export const genericInvestmentScript = createPlaceholderScript({
  id: 'generic-investment',
  title: 'Investment Chat',
  intro: 'Let’s capture a little more context while the investment is fresh.',
  notesHeading: 'Investment Chat Notes',
  thesisPrompt: 'What is the strongest part of the thesis that is not already written down?',
  riskPrompt: 'What is the main risk or invalidation trigger to track?',
  nextPrompt: 'What should the next round of analysis focus on?',
});
