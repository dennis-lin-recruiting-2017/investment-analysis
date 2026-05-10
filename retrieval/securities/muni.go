package securities

// Municipal bond quote retrieval via EMMA (Electronic Municipal Market Access),
// the MSRB's public disclosure system for municipal securities.
//
// # Data sources
//
//   - Reference data (issuer, coupon, maturity, ratings, call provisions):
//     EMMA QuickSearch — public, no authentication required.
//     POST https://emma.msrb.org/QuickSearch/SearchAhead
//
//   - Recent trade data (price, yield, quantity, trade type):
//     EMMA trade endpoint — requires an active MSRB session cookie.
//     POST https://emma.msrb.org/TradeData/JsonGetRecentTrades
//
// # Authentication
//
// EMMA trade data is freely available to registered users.
// To enable trade-price retrieval, call SaveMuniBondQuote with a non-empty
// emmaSession value obtained from the "ASP.NET_SessionId" cookie after
// logging in at https://emma.msrb.org with a free MSRB account.
// If emmaSession is empty the function stores reference-data only and
// records a note explaining why trade prices are absent.
//
// # CUSIP scope
//
// This function covers municipal securities (GO bonds, revenue bonds,
// municipal notes, etc.).  Corporate, agency, and ABS bonds are served
// by SaveBondQuote via FINRA TRACE.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"investment-analysis/util"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"investment-analysis/persistence/model"
)

// MuniBondQuote is the structured snapshot stored for an EMMA quote lookup.
type MuniBondQuote struct {
	CUSIP         string      `json:"cusip"`
	Source        string      `json:"source"`
	FetchedAt     string      `json:"fetched_at"`
	Issuer        string      `json:"issuer,omitempty"`
	Description   string      `json:"description,omitempty"`
	CouponRate    *float64    `json:"coupon_rate,omitempty"`
	MaturityDate  string      `json:"maturity_date,omitempty"`
	CallDate      string      `json:"call_date,omitempty"`
	CallPrice     *float64    `json:"call_price,omitempty"`
	TaxStatus     string      `json:"tax_status,omitempty"`
	MoodyRating   string      `json:"moody_rating,omitempty"`
	SPRating      string      `json:"sp_rating,omitempty"`
	FitchRating   string      `json:"fitch_rating,omitempty"`
	SecurityID    string      `json:"emma_security_id,omitempty"`
	LastTradeDate string      `json:"last_trade_date,omitempty"`
	LastPrice     *float64    `json:"last_price,omitempty"`
	LastYield     *float64    `json:"last_yield,omitempty"`
	RecentTrades  []MuniTrade `json:"recent_trades,omitempty"`
	// Note is populated when trade prices could not be fetched
	// (e.g. no MSRB session provided).
	Note string `json:"note,omitempty"`
}

// MuniTrade is one EMMA-disseminated municipal bond trade record.
type MuniTrade struct {
	TradeDate  string  `json:"trade_date"`
	SettleDate string  `json:"settle_date,omitempty"`
	Price      float64 `json:"price"`
	Yield      float64 `json:"yield,omitempty"`
	// Quantity is the par-value amount traded (USD).
	Quantity float64 `json:"quantity,omitempty"`
	// TradeType is "Customer Buy", "Customer Sell", or "Inter-Dealer".
	TradeType string `json:"trade_type,omitempty"`
}

const emmaBase = "https://emma.msrb.org"

// SaveMuniBondQuote fetches reference and (optionally) trade data for the
// municipal bond identified by cusip from EMMA and stores a dated snapshot.
//
// emmaSession is the value of the "ASP.NET_SessionId" cookie obtained after
// logging in at https://emma.msrb.org (free MSRB account).  Pass an empty
// string to retrieve reference data only; the stored document will include a
// Note field explaining that trade prices require a session.
//
// A date-scoped doc key is used:
//
//	muni-bond-quote|<CUSIP>|<YYYY-MM-DD>
//
// Calling this more than once on the same calendar day (UTC) returns
// "skipped_exists" without hitting the network again.
func (c *Client) SaveMuniBondQuote(ctx context.Context, cusip, emmaSession, out string) (err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "securities.Client.SaveMuniBondQuote",
			"cusip", cusip, "sessionSet", emmaSession != "", "out", out)
	}()
	cusip = strings.ToUpper(strings.TrimSpace(cusip))
	today := time.Now().UTC().Format("2006-01-02")
	key := fmt.Sprintf("muni-bond-quote|%s|%s", cusip, today)

	exists, err := c.docs.Exists(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return c.attempts.Log(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.MuniBondQuote, Ticker: cusip,
			Status: "skipped_exists", Message: "muni bond quote already stored for " + today,
		})
	}

	quote, sourceURL, err := c.fetchEMMABondQuote(ctx, cusip, emmaSession)
	if err != nil {
		_ = c.attempts.Log(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.MuniBondQuote, Ticker: cusip,
			Status: "error", Message: err.Error(),
		})
		return err
	}

	body, err := json.MarshalIndent(quote, "", "  ")
	if err != nil {
		return err
	}

	doc := model.Document{
		DocKey:       key,
		DocumentType: model.MuniBondQuote,
		Ticker:       cusip,
		SourceURL:    sourceURL,
		OutputLabel:  out,
		MimeType:     "application/json",
		Body:         body,
	}
	if err := c.docs.Insert(ctx, &doc); err != nil {
		return err
	}

	msg := fmt.Sprintf("EMMA quote stored for %s on %s", cusip, today)
	if quote.Note != "" {
		msg += " (reference data only — " + quote.Note + ")"
	}
	return c.attempts.Log(ctx, model.RetrievalAttempt{
		DocKey: key, DocumentType: model.MuniBondQuote, Ticker: cusip,
		Status: "stored", Message: msg,
	})
}

// fetchEMMABondQuote queries EMMA for the given CUSIP and returns a
// MuniBondQuote together with the primary source URL used.
//
// It always calls the public SearchAhead endpoint for reference data, then
// attempts the authenticated trade endpoint if emmaSession is non-empty.
func (c *Client) fetchEMMABondQuote(ctx context.Context, cusip, emmaSession string) (_ MuniBondQuote, _ string, err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "securities.Client.fetchEMMABondQuote",
			"cusip", cusip, "sessionSet", emmaSession != "")
	}()
	ec := c.newEMMAClient(emmaSession)

	quote := MuniBondQuote{
		CUSIP:     cusip,
		Source:    emmaBase,
		FetchedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// ── Step 1: reference data via QuickSearch (always public) ──────────────
	searchURL := emmaBase + "/QuickSearch/SearchAhead"
	if err := ec.fetchSearchAhead(ctx, searchURL, cusip, &quote); err != nil {
		// Reference data failure is fatal — without it we have nothing useful.
		return MuniBondQuote{}, searchURL, fmt.Errorf("EMMA QuickSearch for %s: %w", cusip, err)
	}

	// ── Step 2: trade data (requires MSRB session cookie) ───────────────────
	tradeURL := emmaBase + "/TradeData/JsonGetRecentTrades"
	if emmaSession == "" {
		quote.Note = "trade prices not fetched; provide an MSRB session cookie " +
			"(ASP.NET_SessionId from https://emma.msrb.org) to enable them"
	} else {
		if err := ec.fetchRecentTrades(ctx, tradeURL, cusip, &quote); err != nil {
			// Non-fatal: store reference data + the error note.
			quote.Note = "trade prices unavailable: " + err.Error()
		}
	}

	quote.Source = searchURL
	return quote, searchURL, nil
}

// ── EMMA HTTP client ──────────────────────────────────────────────────────────

// emmaHTTPClient wraps an *http.Client pre-loaded with the EMMA session cookie
// and the browser-like headers that EMMA expects.
type emmaHTTPClient struct {
	hc        *http.Client
	userAgent string
}

func (c *Client) newEMMAClient(sessionID string) *emmaHTTPClient {
	jar, _ := cookiejar.New(nil) // error only when Options is non-nil pointer; safe to ignore

	if sessionID != "" {
		emmURL, _ := url.Parse(emmaBase)
		jar.SetCookies(emmURL, []*http.Cookie{
			{Name: "ASP.NET_SessionId", Value: sessionID},
		})
	}

	return &emmaHTTPClient{
		hc: &http.Client{
			Timeout:   30 * time.Second,
			Jar:       jar,
			Transport: c.transport, // nil → http.DefaultTransport; overridden in tests
		},
		userAgent: c.userAgent,
	}
}

func (ec *emmaHTTPClient) post(ctx context.Context, endpoint string, reqBody any) (_ []byte, err error) {
	defer func() { util.LogIfErr(ctx, &err, "securities.emmaHTTPClient.post", "endpoint", endpoint) }()
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", ec.userAgent)
	req.Header.Set("Referer", emmaBase+"/TradeData/Search")

	resp, err := ec.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("EMMA returned %s — the session cookie may be expired or missing", resp.Status)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("EMMA returned unexpected status %s from %s", resp.Status, endpoint)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ── SearchAhead (reference data) ─────────────────────────────────────────────

// fetchSearchAhead calls the EMMA QuickSearch endpoint and populates the
// reference fields of quote (issuer, coupon, maturity, ratings, etc.).
//
// The endpoint accepts a free-text query; sending the CUSIP directly yields
// an exact match in the first result when the CUSIP is known to EMMA.
func (ec *emmaHTTPClient) fetchSearchAhead(ctx context.Context, endpoint, cusip string, quote *MuniBondQuote) (err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "securities.emmaHTTPClient.fetchSearchAhead", "endpoint", endpoint, "cusip", cusip)
	}()
	raw, err := ec.post(ctx, endpoint, map[string]string{"searchText": cusip})
	if err != nil {
		return err
	}

	// EMMA SearchAhead returns a JSON array of result objects.
	var results []map[string]any
	if err := json.Unmarshal(raw, &results); err != nil {
		return fmt.Errorf("decode SearchAhead response: %w", err)
	}
	if len(results) == 0 {
		return fmt.Errorf("CUSIP %s not found in EMMA SearchAhead — verify the CUSIP is a municipal security", cusip)
	}

	// Pick the result whose CUSIP field exactly matches (case-insensitive).
	// If none match exactly, fall back to the first result.
	hit := results[0]
	for _, r := range results {
		if c, _ := strField(r, "cusip"); strings.EqualFold(c, cusip) {
			hit = r
			break
		}
	}

	quote.Issuer, _ = strField(hit, "issuerName")
	if quote.Issuer == "" {
		quote.Issuer, _ = strField(hit, "issuer")
	}
	quote.Description, _ = strField(hit, "description")
	if quote.Description == "" {
		quote.Description, _ = strField(hit, "label")
	}
	quote.MaturityDate, _ = strField(hit, "maturityDate")
	quote.TaxStatus, _ = strField(hit, "taxStatus")
	quote.MoodyRating, _ = strField(hit, "moodyRating")
	quote.SPRating, _ = strField(hit, "spRating")
	quote.FitchRating, _ = strField(hit, "fitchRating")
	quote.SecurityID, _ = strField(hit, "securityId")
	quote.CallDate, _ = strField(hit, "callDate")

	if v, ok := numField(hit, "couponRate"); ok {
		quote.CouponRate = &v
	}
	if v, ok := numField(hit, "callPrice"); ok {
		quote.CallPrice = &v
	}

	return nil
}

// ── JsonGetRecentTrades (trade data) ─────────────────────────────────────────

// fetchRecentTrades calls the EMMA trade endpoint and populates the trade
// fields of quote.  Requires a valid MSRB session cookie in the client jar.
func (ec *emmaHTTPClient) fetchRecentTrades(ctx context.Context, endpoint, cusip string, quote *MuniBondQuote) (err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "securities.emmaHTTPClient.fetchRecentTrades", "endpoint", endpoint, "cusip", cusip)
	}()
	raw, err := ec.post(ctx, endpoint, map[string]string{"cusip": cusip})
	if err != nil {
		return err
	}

	// The trade endpoint returns a JSON array of trade objects.
	var trades []map[string]any
	if err := json.Unmarshal(raw, &trades); err != nil {
		return fmt.Errorf("decode trade response: %w", err)
	}

	for _, t := range trades {
		trade := muniTradeFromEntry(t)
		quote.RecentTrades = append(quote.RecentTrades, trade)
	}

	if len(quote.RecentTrades) > 0 {
		first := quote.RecentTrades[0]
		quote.LastTradeDate = first.TradeDate
		if first.Price != 0 {
			p := first.Price
			quote.LastPrice = &p
		}
		if first.Yield != 0 {
			y := first.Yield
			quote.LastYield = &y
		}
	}

	return nil
}

// muniTradeFromEntry converts one raw EMMA trade map entry into a MuniTrade.
func muniTradeFromEntry(m map[string]any) MuniTrade {
	t := MuniTrade{}
	t.TradeDate, _ = strField(m, "tradeDate")
	t.SettleDate, _ = strField(m, "settlementDate")
	t.TradeType, _ = strField(m, "tradeType")

	for _, k := range []string{"price", "lastSalePrice"} {
		if v, ok := numField(m, k); ok {
			t.Price = v
			break
		}
	}
	t.Yield, _ = numField(m, "yield")
	t.Quantity, _ = numField(m, "quantity")

	return t
}
