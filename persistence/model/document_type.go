package model

// DocumentType identifies the kind of document stored in the retrieval
// system.  Using a named string type rather than bare strings gives
// compile-time safety and a single place to discover all valid values.
type DocumentType string

const (
	// Securities — structured financial data from public sources.
	SECFinancials  DocumentType = "sec-financials"  // SEC EDGAR company facts (XBRL)
	SECFiling      DocumentType = "sec-filing"      // SEC EDGAR raw filing (10-K, 10-Q, …)
	BondQuote      DocumentType = "bond-quote"      // FINRA TRACE corporate/agency bond trade
	MuniBondQuote  DocumentType = "muni-bond-quote" // EMMA municipal bond trade
	TreasuryYields DocumentType = "treasury-yields" // US Treasury par yield curve

	// Real estate — property listing pages and related documents.
	ListingCommercial DocumentType = "listing-commercial" // commercial property listing

	// General documents.
	AnnualReport DocumentType = "annual-report" // company annual report (any format)
)
