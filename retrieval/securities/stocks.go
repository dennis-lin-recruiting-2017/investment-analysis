package securities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"investment-analysis/persistence"
	"investment-analysis/persistence/model"
)

const DefaultUserAgent = "go-download-web-archive/1.0 (contact: you@example.com)"

type Client struct {
	userAgent  string
	httpClient *http.Client
	store      persistence.RetrievalStore
	// transport overrides the HTTP transport for every http.Client created by
	// this instance (including the EMMA session client).  Nil means
	// http.DefaultTransport.  Set only in tests.
	transport http.RoundTripper
}

type tickerEntry struct {
	CIKStr   int    `json:"cik_str"`
	Ticker   string `json:"ticker"`
	Title    string `json:"title"`
	Exchange string `json:"exchange,omitempty"`
}

type filingRecord struct {
	AccessionNumber string
	PrimaryDocument string
	Form            string
	FilingDate      string
	ReportDate      string
}

func NewClient(userAgent string, store persistence.RetrievalStore) *Client {
	if strings.TrimSpace(userAgent) == "" {
		userAgent = DefaultUserAgent
	}
	return &Client{
		userAgent:  userAgent,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		store:      store,
	}
}

func (c *Client) SaveFinancials(ctx context.Context, ticker string, year, quarter int, out string) error {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	key := docKey("sec-financials", ticker, year, quarter)

	exists, err := c.store.DocumentExists(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.SECFinancials, Ticker: ticker,
			FiscalYear: year, FiscalQuarter: quarter,
			Status: "skipped_exists", Message: "document already exists; retrieval skipped",
		})
	}

	entry, err := c.lookupTicker(ctx, ticker)
	if err != nil {
		_ = c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.SECFinancials, Ticker: ticker,
			FiscalYear: year, FiscalQuarter: quarter,
			Status: "error", Message: err.Error(),
		})
		return err
	}

	factsURL := fmt.Sprintf("https://data.sec.gov/api/xbrl/companyfacts/CIK%010d.json", entry.CIKStr)
	payload, err := c.fetchBytes(ctx, factsURL)
	if err != nil {
		_ = c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.SECFinancials, Ticker: ticker,
			FiscalYear: year, FiscalQuarter: quarter,
			Status: "error", Message: err.Error(),
		})
		return err
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		_ = c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.SECFinancials, Ticker: ticker,
			FiscalYear: year, FiscalQuarter: quarter,
			Status: "error", Message: err.Error(),
		})
		return fmt.Errorf("decode company facts: %w", err)
	}

	result := map[string]any{
		"ticker":         ticker,
		"cik":            fmt.Sprintf("%010d", entry.CIKStr),
		"company_name":   entry.Title,
		"fiscal_year":    year,
		"fiscal_quarter": quarter,
		"source":         factsURL,
		"fetched_at":     time.Now().UTC().Format(time.RFC3339),
		"facts":          extractCommonFacts(raw, year, quarter),
	}
	body, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	doc := model.StoredDocument{
		Key: key, DocumentType: model.SECFinancials, Ticker: ticker,
		FiscalYear: year, FiscalQtr: quarter,
		SourceURL: factsURL, OutputLabel: out,
		MimeType: "application/json", Body: body,
	}
	if err := c.store.InsertDocument(ctx, doc); err != nil {
		return err
	}
	return c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
		DocKey: key, DocumentType: model.SECFinancials, Ticker: ticker,
		FiscalYear: year, FiscalQuarter: quarter,
		Status: "stored", Message: "financials stored in sqlite",
	})
}

func (c *Client) SaveFiling(ctx context.Context, ticker string, year, quarter int, filingForm, out string) error {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	key := docKey("sec-filing", ticker, year, quarter)

	exists, err := c.store.DocumentExists(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.SECFiling, Ticker: ticker,
			FiscalYear: year, FiscalQuarter: quarter,
			Status: "skipped_exists", Message: "document already exists; retrieval skipped",
		})
	}

	entry, err := c.lookupTicker(ctx, ticker)
	if err != nil {
		_ = c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.SECFiling, Ticker: ticker,
			FiscalYear: year, FiscalQuarter: quarter,
			Status: "error", Message: err.Error(),
		})
		return err
	}

	record, archiveURL, err := c.resolveFiling(ctx, entry.CIKStr, year, quarter)
	if err != nil {
		_ = c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.SECFiling, Ticker: ticker,
			FiscalYear: year, FiscalQuarter: quarter,
			Status: "error", Message: err.Error(),
		})
		return err
	}

	body, mimeType, err := c.fetchWithContentType(ctx, archiveURL)
	if err != nil {
		_ = c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.SECFiling, Ticker: ticker,
			FiscalYear: year, FiscalQuarter: quarter,
			Status: "error", Message: err.Error(),
		})
		return err
	}

	if strings.TrimSpace(filingForm) != "" && strings.ToLower(strings.TrimSpace(filingForm)) != "auto" && !strings.EqualFold(record.Form, filingForm) {
		msg := fmt.Sprintf("resolved form %s did not match requested %s", record.Form, filingForm)
		_ = c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.SECFiling, Ticker: ticker,
			FiscalYear: year, FiscalQuarter: quarter,
			Status: "error", Message: msg,
		})
		return errors.New(msg)
	}

	doc := model.StoredDocument{
		Key: key, DocumentType: model.SECFiling, Ticker: ticker,
		FiscalYear: year, FiscalQtr: quarter,
		Form: record.Form, SourceURL: archiveURL, OutputLabel: out,
		MimeType: mimeType, Body: body,
	}
	if err := c.store.InsertDocument(ctx, doc); err != nil {
		return err
	}
	return c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
		DocKey: key, DocumentType: model.SECFiling, Ticker: ticker,
		FiscalYear: year, FiscalQuarter: quarter,
		Status: "stored", Message: fmt.Sprintf("stored %s from %s", record.Form, archiveURL),
	})
}

func (c *Client) DumpDocumentToFile(ctx context.Context, key, path string) error {
	body, err := c.store.GetDocumentBody(ctx, key)
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}

func (c *Client) lookupTicker(ctx context.Context, ticker string) (tickerEntry, error) {
	payload, err := c.fetchBytes(ctx, "https://www.sec.gov/files/company_tickers.json")
	if err != nil {
		return tickerEntry{}, fmt.Errorf("fetch ticker map: %w", err)
	}
	var all map[string]tickerEntry
	if err := json.Unmarshal(payload, &all); err != nil {
		return tickerEntry{}, fmt.Errorf("decode ticker map: %w", err)
	}
	for _, v := range all {
		if strings.EqualFold(v.Ticker, ticker) {
			return v, nil
		}
	}
	return tickerEntry{}, fmt.Errorf("ticker %s not found in SEC ticker map", ticker)
}

func (c *Client) resolveFiling(ctx context.Context, cik, year, quarter int) (filingRecord, string, error) {
	submissionsURL := fmt.Sprintf("https://data.sec.gov/submissions/CIK%010d.json", cik)
	payload, err := c.fetchBytes(ctx, submissionsURL)
	if err != nil {
		return filingRecord{}, "", fmt.Errorf("fetch submissions: %w", err)
	}
	var submissions struct {
		Filings struct {
			Recent struct {
				AccessionNumber []string `json:"accessionNumber"`
				PrimaryDocument []string `json:"primaryDocument"`
				Form            []string `json:"form"`
				FilingDate      []string `json:"filingDate"`
				ReportDate      []string `json:"reportDate"`
			} `json:"recent"`
		} `json:"filings"`
	}
	if err := json.Unmarshal(payload, &submissions); err != nil {
		return filingRecord{}, "", fmt.Errorf("decode submissions: %w", err)
	}
	var candidates []filingRecord
	for i, form := range submissions.Filings.Recent.Form {
		if form != "10-Q" && form != "10-K" {
			continue
		}
		reportDate := valueAt(submissions.Filings.Recent.ReportDate, i)
		if reportDate == "" {
			continue
		}
		fiscalYear, fiscalQuarter, ok := inferFiscalFromReportDate(reportDate)
		if !ok {
			continue
		}
		if fiscalYear != year || fiscalQuarter != quarter {
			continue
		}
		candidates = append(candidates, filingRecord{
			AccessionNumber: valueAt(submissions.Filings.Recent.AccessionNumber, i),
			PrimaryDocument: valueAt(submissions.Filings.Recent.PrimaryDocument, i),
			Form:            form,
			FilingDate:      valueAt(submissions.Filings.Recent.FilingDate, i),
			ReportDate:      reportDate,
		})
	}
	if len(candidates) == 0 {
		return filingRecord{}, "", fmt.Errorf("no 10-Q or 10-K found for requested fiscal period %d Q%d", year, quarter)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Form != candidates[j].Form {
			if quarter == 4 {
				return candidates[i].Form == "10-K"
			}
			return candidates[i].Form == "10-Q"
		}
		return candidates[i].FilingDate > candidates[j].FilingDate
	})
	chosen := candidates[0]
	archiveURL := fmt.Sprintf("https://www.sec.gov/Archives/edgar/data/%d/%s/%s",
		cik, strings.ReplaceAll(chosen.AccessionNumber, "-", ""), chosen.PrimaryDocument)
	return chosen, archiveURL, nil
}

func (c *Client) fetchBytes(ctx context.Context, url string) ([]byte, error) {
	body, _, err := c.fetchWithContentType(ctx, url)
	return body, err
}

func (c *Client) fetchWithContentType(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept-Encoding", "identity")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, "", fmt.Errorf("unexpected HTTP status %s from %s: %s", resp.Status, url, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return body, mimeType, nil
}

func inferFiscalFromReportDate(reportDate string) (int, int, bool) {
	t, err := time.Parse("2006-01-02", reportDate)
	if err != nil {
		return 0, 0, false
	}
	m := t.Month()
	y := t.Year()
	switch {
	case m == 11 || m == 12:
		return y + 1, 1, true
	case m == 1:
		return y, 1, true
	case m >= 2 && m <= 4:
		return y, 2, true
	case m >= 5 && m <= 7:
		return y, 3, true
	default:
		return y, 4, true
	}
}

func extractCommonFacts(raw map[string]any, year, quarter int) map[string]any {
	usGaap, _ := digMap(raw, "facts", "us-gaap")
	if usGaap == nil {
		return map[string]any{}
	}
	candidates := map[string][]string{
		"revenue":                   {"Revenues", "RevenueFromContractWithCustomerExcludingAssessedTax", "SalesRevenueNet"},
		"gross_profit":              {"GrossProfit"},
		"operating_income":          {"OperatingIncomeLoss"},
		"net_income":                {"NetIncomeLoss"},
		"sg_and_a":                  {"SellingGeneralAndAdministrativeExpense"},
		"research_and_development":  {"ResearchAndDevelopmentExpense"},
		"cash_and_cash_equivalents": {"CashAndCashEquivalentsAtCarryingValue"},
		"total_assets":              {"Assets"},
		"total_liabilities":         {"Liabilities"},
		"stockholders_equity":       {"StockholdersEquity"},
		"operating_cash_flow":       {"NetCashProvidedByUsedInOperatingActivities"},
	}
	out := map[string]any{}
	for label, tags := range candidates {
		for _, tag := range tags {
			if fact := bestFactForPeriod(usGaap, tag, year, quarter); fact != nil {
				out[label] = fact
				break
			}
		}
	}
	return out
}

func bestFactForPeriod(usGaap map[string]any, tag string, year, quarter int) map[string]any {
	tagMap, ok := usGaap[tag].(map[string]any)
	if !ok {
		return nil
	}
	units, ok := tagMap["units"].(map[string]any)
	if !ok {
		return nil
	}
	for _, unit := range []string{"USD", "USD/shares", "shares"} {
		arr, ok := units[unit].([]any)
		if !ok {
			continue
		}
		var best map[string]any
		for _, item := range arr {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if fy, ok := asInt(m["fy"]); ok && fy != year {
				continue
			}
			if fp, ok := m["fp"].(string); ok && quarterFromFP(fp) != quarter {
				continue
			}
			best = m
		}
		if best != nil {
			return best
		}
	}
	return nil
}

func quarterFromFP(fp string) int {
	switch strings.ToUpper(strings.TrimSpace(fp)) {
	case "Q1":
		return 1
	case "Q2":
		return 2
	case "Q3":
		return 3
	case "Q4", "FY":
		return 4
	default:
		return 0
	}
}

func asInt(v any) (int, bool) {
	switch x := v.(type) {
	case float64:
		return int(x), true
	case int:
		return x, true
	default:
		return 0, false
	}
}

func digMap(root map[string]any, path ...string) (map[string]any, bool) {
	cur := any(root)
	for _, p := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[p]
		if !ok {
			return nil, false
		}
	}
	m, ok := cur.(map[string]any)
	return m, ok
}

func docKey(mode, ticker string, year, quarter int) string {
	return fmt.Sprintf("%s|%s|%d|Q%d", mode, strings.ToUpper(strings.TrimSpace(ticker)), year, quarter)
}

func valueAt(items []string, i int) string {
	if i < 0 || i >= len(items) {
		return ""
	}
	return items[i]
}
