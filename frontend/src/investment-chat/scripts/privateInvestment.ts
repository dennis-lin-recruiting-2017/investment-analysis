import { createPlaceholderScript } from './createPlaceholderScript';

export const privateInvestmentScript = createPlaceholderScript({
  id: 'private-investment',
  title: 'Private Investment Chat',
  intro: 'This placeholder private investment interview gives us a stable spot for deal-specific diligence questions later.',
  notesHeading: 'Private Investment Chat Notes',
  thesisPrompt: 'What is the main reason this private investment belongs in the pipeline?',
  riskPrompt: 'What is the biggest underwriting, governance, liquidity, or execution risk to monitor?',
  nextPrompt: 'What follow-up question would most improve confidence in this private investment?',
});
