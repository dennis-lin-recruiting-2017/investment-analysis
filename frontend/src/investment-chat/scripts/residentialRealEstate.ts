import { createPlaceholderScript } from './createPlaceholderScript';

export const residentialRealEstateScript = createPlaceholderScript({
  id: 'residential-real-estate',
  title: 'Residential Real Estate Chat',
  intro: 'This placeholder residential real estate interview gives us a place to grow rent, vacancy, financing, and exit questions later.',
  notesHeading: 'Residential Real Estate Chat Notes',
  thesisPrompt: 'What is the main reason this residential property looks attractive?',
  riskPrompt: 'What is the main rent, vacancy, financing, maintenance, or regulatory risk to track?',
  nextPrompt: 'What follow-up question would most improve the analysis for this property?',
});
