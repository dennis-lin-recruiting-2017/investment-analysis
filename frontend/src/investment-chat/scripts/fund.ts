import { createPlaceholderScript } from './createPlaceholderScript';

export const fundScript = createPlaceholderScript({
  id: 'fund',
  title: 'Fund Chat',
  intro: 'This placeholder fund interview gives us a simple structure for manager, strategy, and fee questions later.',
  notesHeading: 'Fund Chat Notes',
  thesisPrompt: 'What makes this fund or manager worth allocating to?',
  riskPrompt: 'What is the main strategy, manager, liquidity, or fee risk to watch?',
  nextPrompt: 'What follow-up question would most improve the analysis for this fund?',
});
