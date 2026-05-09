package securities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"investment-analysis/persistence/model"
)

// BondQuote is the structured snapshot stored for a CUSIP quote lookup.
type BondQuote struct {
	CUSIP         string      `json:"cusip"`
	Source        string      `json:"source"`
	FetchedAt     string      `json:"fetched_at"`
	Description   string      `json:"description,omitempty"`
	CouponRate    *float64    `json:"coupon_rate,omitempty"`
	MaturityDate  string      `json:"maturity_date,omitempty"`
	LastTradeDate string      `json:"last_trade_date,omitempty"`
	LastPrice     *float64    `json:"last_price,omitempty"`
	LastYield     *float64    `json:"last_yield,omitempty"`
	MoodyRating   string      `json:"moody_rating,omitempty"`
	SPRating      string      `json:"sp_rating,omitempty"`
	RecentTrades  []BondTrade `json:"recent_trades,omitempty"`
}

// BondTrade is one TRACE-disseminated trade record.
type BondTrade struct {
	TradeDate string  `json:"trade_date"`
	TradeTime string  `json:"trade_time,omitempty"`
	Price     float64 `json:"price"`
	Yield     float64 `json:"yield,omitempty"`
	// Quantity is the par-value amount traded (USD).
	Quantity float64 `json:"quantity,omitempty"`
	// Side is "B" (customer buy), "S" (customer sell), or "D" (inter-dealer).
	Side string `json:"side,omitempty"`
}

// finraTraceBase is the FINRA Market Data / TRACE dissemination API.
// No authentication is required; data are publicly mandated by FINRA Rule 6750.
// Reference: https://www.finra.org/rules-guidance/rulebooks/finra-rules/6750
const finraTraceBase = "https://mds.finra.org/bd/getTradeActivity"

// SaveBondQuote fetches the latest TRACE trade data for the bond identified by
// cusip from FINRA and stores a dated snapshot.
//
// Because bond prices change intraday a date-scoped doc key is used:
//
//	bond-quote|<CUSIP>|<YYYY-MM-DD>
//
// Calling this more than once on the same calendar day (UTC) returns
// "skipped_exists" without hitting the network again.  Pass a new date (or
// delete the stored row) to force a refresh.
//
// Coverage: corporate bonds, agency bonds, and asset-backed securities that
// are TRACE-eligible.  Municipal bonds are not covered by FINRA TRACE; use
// EMMA (emma.msrb.org) for those.
func (c *Client) SaveBondQuote(ctx context.Context, cusip, out string) error {
	cusip = strings.ToUpper(strings.TrimSpace(cusip))
	today := time.Now().UTC().Format("2006-01-02")
	key := fmt.Sprintf("bond-quote|%s|%s", cusip, today)

	exists, err := c.store.DocumentExists(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.BondQuote, Ticker: cusip,
			Status: "skipped_exists", Message: "bond quote already stored for " + today,
		})
	}

	quote, sourceURL, err := c.fetchFINRABondQuote(ctx, cusip)
	if err != nil {
		_ = c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.BondQuote, Ticker: cusip,
			Status: "error", Message: err.Error(),
		})
		return err
	}

	body, err := json.MarshalIndent(quote, "", "  ")
	if err != nil {
		return err
	}

	doc := model.StoredDocument{
		Key:         key,
		DocumentType:  model.BondQuote,
		Ticker:      cusip,
		SourceURL:   sourceURL,
		OutputLabel: out,
		MimeType:    "application/json",
		Body:        body,
	}
	if err := c.store.InsertDocument(ctx, doc); err != nil {
		return err
	}
	return c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
		DocKey: key, DocumentType: model.BondQuote, Ticker: cusip,
		Status:  "stored",
		Message: fmt.Sprintf("FINRA TRACE quote stored for %s on %s", cusip, today),
	})
}

// fetchFINRABondQuote calls the FINRA TRACE dissemination API and returns a
// BondQuote together with the URL that was fetched.
//
// The API returns the most recent TRACE-reported trades for the CUSIP.  The
// first (most recent) trade is used to populate the top-level summary fields;
// all trades are retained in RecentTrades.
func (c *Client) fetchFINRABondQuote(ctx context.Context, cusip string) (BondQuote, string, error) {
	// type=trd  → trade records (as opposed to reference data).
	// No date range is supplied so the API returns the most recent trades.
	sourceURL := fmt.Sprintf("%s?symbol=%s&type=trd", finraTraceBase, cusip)

	raw, err := c.fetchBytes(ctx, sourceURL)
	if err != nil {
		return BondQuote{}, sourceURL, fmt.Errorf("fetch FINRA TRACE for CUSIP %s: %w", cusip, err)
	}

	// FINRA wraps results in { "data": [...] }.
	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return BondQuote{}, sourceURL, fmt.Errorf("decode FINRA TRACE response for %s: %w", cusip, err)
	}

	if len(envelope.Data) == 0 {
		return BondQuote{}, sourceURL,
			fmt.Errorf("no TRACE data found for CUSIP %s (bond may not be TRACE-eligible or CUSIP is incorrect)", cusip)
	}

	quote := BondQuote{
		CUSIP:     cusip,
		Source:    sourceURL,
		FetchedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Populate summary fields from the most recent trade entry.
	first := envelope.Data[0]
	quote.LastTradeDate, _ = strField(first, "tradeDate")
	quote.Description, _ = strField(first, "issueDescription")
	if quote.Description == "" {
		quote.Description, _ = strField(first, "description")
	}
	quote.MaturityDate, _ = strField(first, "maturityDate")
	quote.MoodyRating, _ = strField(first, "moodyRating")
	quote.SPRating, _ = strField(first, "spRating")

	if p, ok := numField(first, "lastSalePrice"); ok {
		quote.LastPrice = &p
	} else if p, ok := numField(first, "price"); ok {
		quote.LastPrice = &p
	}
	if y, ok := numField(first, "yield"); ok {
		quote.LastYield = &y
	}
	if cr, ok := numField(first, "couponRate"); ok {
		quote.CouponRate = &cr
	}

	// Collect all trade records.
	for _, entry := range envelope.Data {
		quote.RecentTrades = append(quote.RecentTrades, bondTradeFromEntry(entry))
	}

	return quote, sourceURL, nil
}

// bondTradeFromEntry converts one raw FINRA TRACE map entry into a BondTrade.
func bondTradeFromEntry(m map[string]any) BondTrade {
	t := BondTrade{}
	t.TradeDate, _ = strField(m, "tradeDate")
	t.TradeTime, _ = strField(m, "tradeTime")

	// FINRA may use "lastSalePrice" or "price" depending on feed version.
	for _, k := range []string{"lastSalePrice", "price"} {
		if v, ok := numField(m, k); ok {
			t.Price = v
			break
		}
	}
	t.Yield, _ = numField(m, "yield")
	t.Quantity, _ = numField(m, "quantity")

	// "side" values: "B" = customer buy, "S" = customer sell, "D" = inter-dealer.
	for _, k := range []string{"side", "tradeType", "buyOrSell"} {
		if v, ok := strField(m, k); ok {
			t.Side = v
			break
		}
	}
	return t
}

// strField safely extracts a string value from a loosely-typed JSON map.
func strField(m map[string]any, key string) (string, bool) {
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}

// numField safely extracts a numeric value from a loosely-typed JSON map.
func numField(m map[string]any, key string) (float64, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		return x, x != 0
	case int:
		return float64(x), x != 0
	}
	return 0, false
}
