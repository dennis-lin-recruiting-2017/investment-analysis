import { createPlaceholderScript } from './createPlaceholderScript';

export const bondCorporateScript = createPlaceholderScript({
  id: 'bond-corporate',
  title: 'Bond - Corporate Chat',
  intro: 'This placeholder corporate bond interview gives us a clean place to expand issuer and covenant questions later.',
  notesHeading: 'Bond - Corporate Chat Notes',
  thesisPrompt: 'What makes this corporate bond the right risk-return tradeoff?',
  riskPrompt: 'What is the main issuer, refinancing, spread, or call risk to track?',
  nextPrompt: 'What follow-up question would most improve the analysis for this corporate bond?',
});
