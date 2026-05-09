import { createPlaceholderScript } from './createPlaceholderScript';

export const otherInvestmentScript = createPlaceholderScript({
  id: 'other-investment',
  title: 'Other Investment Chat',
  intro: 'This placeholder interview covers anything that does not fit a more specific asset-class flow yet.',
  notesHeading: 'Other Investment Chat Notes',
  thesisPrompt: 'What makes this investment worth tracking even though it falls outside the usual categories?',
  riskPrompt: 'What is the main unknown, structural risk, or diligence gap to keep in view?',
  nextPrompt: 'What follow-up question would most improve the analysis for this investment?',
});
