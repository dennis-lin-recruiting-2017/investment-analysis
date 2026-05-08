import { bondUsTreasuryScript } from './bondUsTreasury';
import { bondCorporateScript } from './bondCorporate';
import { bondStateMunicipalScript } from './bondStateMunicipal';
import { commercialRealEstateScript } from './commercialRealEstate';
import { etfScript } from './etf';
import { fundScript } from './fund';
import { genericInvestmentScript } from './generic';
import { otherInvestmentScript } from './other';
import { privateInvestmentScript } from './privateInvestment';
import { residentialRealEstateScript } from './residentialRealEstate';
import { stockScript } from './stock';
import type { ChatScript } from '../types';

export function scriptForAssetClass(assetClass: string): ChatScript {
  switch (assetClass) {
    case 'Stock':
      return stockScript;
    case 'ETF':
      return etfScript;
    case 'Bond - US Treasury':
      return bondUsTreasuryScript;
    case 'Bond - State or Municipal':
      return bondStateMunicipalScript;
    case 'Bond - Corporate':
      return bondCorporateScript;
    case 'Fund':
      return fundScript;
    case 'Residential Real Estate':
      return residentialRealEstateScript;
    case 'Commercial Real Estate':
      return commercialRealEstateScript;
    case 'Private investment':
      return privateInvestmentScript;
    case 'Other':
      return otherInvestmentScript;
    default:
      return genericInvestmentScript;
  }
}
