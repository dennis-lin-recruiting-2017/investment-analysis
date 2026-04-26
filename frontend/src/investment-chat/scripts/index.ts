import { bondUsTreasuryScript } from './bondUsTreasury';
import { commercialRealEstateScript } from './commercialRealEstate';
import { genericInvestmentScript } from './generic';
import type { ChatScript } from '../types';

export function scriptForAssetClass(assetClass: string): ChatScript {
  switch (assetClass) {
    case 'Bond - US Treasury':
      return bondUsTreasuryScript;
    case 'Commercial Real Estate':
      return commercialRealEstateScript;
    default:
      return genericInvestmentScript;
  }
}
