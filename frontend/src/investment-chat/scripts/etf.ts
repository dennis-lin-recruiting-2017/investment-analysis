import { createPlaceholderScript } from './createPlaceholderScript';

export const etfScript = createPlaceholderScript({
  id: 'etf',
  title: 'ETF Chat',
  intro: 'This placeholder ETF interview gives us a slot for portfolio-role questions that can get richer later.',
  notesHeading: 'ETF Chat Notes',
  thesisPrompt: 'Why is this ETF the right instrument for the exposure you want?',
  riskPrompt: 'What tracking, concentration, cost, or timing risk matters most for this ETF?',
  nextPrompt: 'What should the next analysis step focus on for this ETF allocation?',
});
