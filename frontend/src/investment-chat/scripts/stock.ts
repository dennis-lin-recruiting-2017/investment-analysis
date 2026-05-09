import { createPlaceholderScript } from './createPlaceholderScript';

export const stockScript = createPlaceholderScript({
  id: 'stock',
  title: 'Stock Chat',
  intro: 'This placeholder stock interview can grow later into a fuller equity research flow.',
  notesHeading: 'Stock Chat Notes',
  thesisPrompt: 'What is the clearest edge, catalyst, or business quality that makes this stock interesting?',
  riskPrompt: 'What is the main business, valuation, or market risk that could break the thesis?',
  nextPrompt: 'What follow-up question would most improve conviction in this stock?',
});
