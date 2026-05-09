import { createPlaceholderScript } from './createPlaceholderScript';

export const bondStateMunicipalScript = createPlaceholderScript({
  id: 'bond-state-municipal',
  title: 'Bond - State or Municipal Chat',
  intro: 'This placeholder muni bond interview gives us room to add tax and credit-specific branches later.',
  notesHeading: 'Bond - State or Municipal Chat Notes',
  thesisPrompt: 'What makes this state or municipal bond attractive right now?',
  riskPrompt: 'What is the main credit, call, duration, or tax-sensitive risk to monitor?',
  nextPrompt: 'What follow-up question would most improve the analysis for this bond?',
});
