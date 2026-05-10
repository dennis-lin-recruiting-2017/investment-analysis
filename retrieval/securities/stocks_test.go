package securities

import (
	"encoding/json"
	"investment-analysis/util"
	"strings"
	"testing"
)

// ── inferFiscalFromReportDate ─────────────────────────────────────────────────

func TestInferFiscalFromReportDate(t *testing.T) {
	tests := []struct {
		date   string
		wantY  int
		wantQ  int
		wantOK bool
	}{
		// November / December → Q1 of the following year
		{"2023-11-30", 2024, 1, true},
		{"2023-12-31", 2024, 1, true},
		// January → Q1 of the same year
		{"2024-01-31", 2024, 1, true},
		// February – April → Q2
		{"2024-02-29", 2024, 2, true},
		{"2024-03-31", 2024, 2, true},
		{"2024-04-30", 2024, 2, true},
		// May – July → Q3
		{"2024-05-31", 2024, 3, true},
		{"2024-07-31", 2024, 3, true},
		// August – October → Q4
		{"2024-08-31", 2024, 4, true},
		{"2024-10-31", 2024, 4, true},
		// Invalid date
		{"not-a-date", 0, 0, false},
		{"2024-13-01", 0, 0, false},
	}
	for _, tc := range tests {
		y, q, ok := inferFiscalFromReportDate(tc.date)
		if ok != tc.wantOK || y != tc.wantY || q != tc.wantQ {
			t.Errorf("inferFiscalFromReportDate(%q) = (%d, %d, %v); want (%d, %d, %v)",
				tc.date, y, q, ok, tc.wantY, tc.wantQ, tc.wantOK)
		}
	}
}

// ── quarterFromFP ─────────────────────────────────────────────────────────────

func TestQuarterFromFP(t *testing.T) {
	tests := []struct {
		fp   string
		want int
	}{
		{"Q1", 1}, {"q1", 1},
		{"Q2", 2}, {"Q3", 3},
		{"Q4", 4}, {"FY", 4}, {"fy", 4},
		{"", 0}, {"annual", 0},
	}
	for _, tc := range tests {
		if got := quarterFromFP(tc.fp); got != tc.want {
			t.Errorf("quarterFromFP(%q) = %d; want %d", tc.fp, got, tc.want)
		}
	}
}

// ── docKey ────────────────────────────────────────────────────────────────────

func TestDocKey(t *testing.T) {
	got := docKey("sec-financials", "aapl", 2024, 1)
	want := "sec-financials|AAPL|2024|Q1"
	if got != want {
		t.Errorf("docKey = %q; want %q", got, want)
	}
}

// ── valueAt ──────────────────────────────────────────────────────────────────

func TestValueAt(t *testing.T) {
	items := []string{"a", "b", "c"}
	if got := valueAt(items, 1); got != "b" {
		t.Errorf("valueAt in-bounds = %q; want %q", got, "b")
	}
	if got := valueAt(items, 5); got != "" {
		t.Errorf("valueAt out-of-bounds = %q; want empty", got)
	}
	if got := valueAt(items, -1); got != "" {
		t.Errorf("valueAt negative = %q; want empty", got)
	}
}

// ── asInt ─────────────────────────────────────────────────────────────────────

func TestAsInt(t *testing.T) {
	if v, ok := asInt(float64(42)); !ok || v != 42 {
		t.Error("asInt(float64) failed")
	}
	if v, ok := asInt(7); !ok || v != 7 {
		t.Error("asInt(int) failed")
	}
	if _, ok := asInt("seven"); ok {
		t.Error("asInt(string) should return false")
	}
}

// ── digMap ────────────────────────────────────────────────────────────────────

func TestDigMap(t *testing.T) {
	root := map[string]any{
		"a": map[string]any{
			"b": map[string]any{"c": "value"},
		},
	}

	m, ok := digMap(root, "a", "b")
	if !ok || m["c"] != "value" {
		t.Error("digMap: expected successful traversal")
	}

	_, ok = digMap(root, "a", "missing")
	if ok {
		t.Error("digMap: missing key should return false")
	}

	_, ok = digMap(root, "a", "b", "c") // "c" is a string, not a map
	if ok {
		t.Error("digMap: non-map value at end should return false")
	}
}

// ── bestFactForPeriod ─────────────────────────────────────────────────────────

func TestBestFactForPeriod(t *testing.T) {
	entry := map[string]any{"fy": float64(2024), "fp": "Q1", "val": float64(100)}
	wrongYear := map[string]any{"fy": float64(2023), "fp": "Q1", "val": float64(50)}
	wrongQtr := map[string]any{"fy": float64(2024), "fp": "Q2", "val": float64(75)}

	usGaap := map[string]any{
		"Revenues": map[string]any{
			"units": map[string]any{
				"USD": []any{wrongYear, wrongQtr, entry},
			},
		},
	}

	got := bestFactForPeriod(usGaap, "Revenues", 2024, 1)
	if got == nil {
		t.Fatal("bestFactForPeriod: expected a match, got nil")
	}
	if got["val"] != float64(100) {
		t.Errorf("bestFactForPeriod val = %v; want 100", got["val"])
	}

	// Tag not in usGaap
	if bestFactForPeriod(usGaap, "NonExistentTag", 2024, 1) != nil {
		t.Error("bestFactForPeriod: unknown tag should return nil")
	}

	// Year mismatch
	if bestFactForPeriod(usGaap, "Revenues", 2025, 1) != nil {
		t.Error("bestFactForPeriod: year mismatch should return nil")
	}
}

// ── extractCommonFacts ────────────────────────────────────────────────────────

func TestExtractCommonFacts(t *testing.T) {
	raw := map[string]any{
		"facts": map[string]any{
			"us-gaap": map[string]any{
				"NetIncomeLoss": map[string]any{
					"units": map[string]any{
						"USD": []any{
							map[string]any{"fy": float64(2024), "fp": "Q1", "val": float64(9999)},
						},
					},
				},
			},
		},
	}

	facts := extractCommonFacts(raw, 2024, 1)
	if _, ok := facts["net_income"]; !ok {
		t.Error("extractCommonFacts: expected net_income key")
	}

	// Missing us-gaap section returns empty map, not nil
	empty := extractCommonFacts(map[string]any{}, 2024, 1)
	if len(empty) != 0 {
		t.Error("extractCommonFacts: empty raw should return empty map")
	}
}

// ── SaveFinancials ────────────────────────────────────────────────────────────

const (
	mockTickerJSON = `{
		"0": {"cik_str": 320193, "ticker": "AAPL", "title": "Apple Inc."}
	}`

	mockCompanyFactsJSON = `{
		"entityName": "Apple Inc.",
		"facts": {
			"us-gaap": {
				"Revenues": {
					"units": {
						"USD": [
							{"fy": 2024, "fp": "Q1", "val": 119575000000,
							 "accn": "0000320193-24-000005", "form": "10-Q", "filed": "2024-02-02"}
						]
					}
				}
			}
		}
	}`
)

func TestSaveFinancials_AlreadyExists(t *testing.T) {
	store := newMockStore()
	key := "sec-financials|AAPL|2024|Q1"
	store.docs[key] = []byte(`{}`) // pre-seed so DocumentExists returns true

	c := newTestClient(store, &mockTransport{})
	if err := c.SaveFinancials(util.NewTraceContext(), "AAPL", 2024, 1, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	last, _ := store.lastAttempt()
	if last.Status != "skipped_exists" {
		t.Errorf("status = %q; want skipped_exists", last.Status)
	}
}

func TestSaveFinancials_Success(t *testing.T) {
	store := newMockStore()
	mt := &mockTransport{}
	mt.add("company_tickers.json", 200, mockTickerJSON)
	mt.add("companyfacts", 200, mockCompanyFactsJSON)

	c := newTestClient(store, mt)
	if err := c.SaveFinancials(util.NewTraceContext(), "aapl", 2024, 1, "test-out"); err != nil {
		t.Fatalf("SaveFinancials error: %v", err)
	}

	key := "sec-financials|AAPL|2024|Q1"
	if _, ok := store.docs[key]; !ok {
		t.Error("document was not inserted into store")
	}

	last, _ := store.lastAttempt()
	if last.Status != "stored" {
		t.Errorf("final attempt status = %q; want stored", last.Status)
	}

	// Verify the stored JSON is parseable and contains expected fields.
	var result map[string]any
	if err := json.Unmarshal(store.docs[key], &result); err != nil {
		t.Fatalf("stored body is invalid JSON: %v", err)
	}
	if result["ticker"] != "AAPL" {
		t.Errorf("stored ticker = %v; want AAPL", result["ticker"])
	}
}

func TestSaveFinancials_TickerNotFound(t *testing.T) {
	store := newMockStore()
	mt := &mockTransport{}
	// Return a ticker map that does NOT contain MSFT.
	mt.add("company_tickers.json", 200, `{"0":{"cik_str":320193,"ticker":"AAPL","title":"Apple"}}`)

	c := newTestClient(store, mt)
	err := c.SaveFinancials(util.NewTraceContext(), "MSFT", 2024, 1, "")
	if err == nil {
		t.Fatal("expected an error for unknown ticker, got nil")
	}
	if !strings.Contains(err.Error(), "MSFT") {
		t.Errorf("error %q does not mention MSFT", err.Error())
	}
}

// ── SaveFiling ────────────────────────────────────────────────────────────────

const mockSubmissionsJSON = `{
	"filings": {
		"recent": {
			"accessionNumber": ["0000320193-24-000123"],
			"primaryDocument": ["aapl20240101_10q.htm"],
			"form":            ["10-Q"],
			"filingDate":      ["2024-02-02"],
			"reportDate":      ["2024-01-31"]
		}
	}
}`

func TestSaveFiling_AlreadyExists(t *testing.T) {
	store := newMockStore()
	key := "sec-filing|AAPL|2024|Q1"
	store.docs[key] = []byte(`{}`)

	c := newTestClient(store, &mockTransport{})
	if err := c.SaveFiling(util.NewTraceContext(), "AAPL", 2024, 1, "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	last, _ := store.lastAttempt()
	if last.Status != "skipped_exists" {
		t.Errorf("status = %q; want skipped_exists", last.Status)
	}
}

func TestSaveFiling_Success(t *testing.T) {
	store := newMockStore()
	mt := &mockTransport{}
	mt.add("company_tickers.json", 200, mockTickerJSON)
	mt.add("submissions/CIK", 200, mockSubmissionsJSON)
	mt.add("Archives/edgar", 200, "<html>10-Q content</html>")

	c := newTestClient(store, mt)
	if err := c.SaveFiling(util.NewTraceContext(), "AAPL", 2024, 1, "10-Q", ""); err != nil {
		t.Fatalf("SaveFiling error: %v", err)
	}

	key := "sec-filing|AAPL|2024|Q1"
	if _, ok := store.docs[key]; !ok {
		t.Error("filing document was not inserted into store")
	}

	last, _ := store.lastAttempt()
	if last.Status != "stored" {
		t.Errorf("final attempt status = %q; want stored", last.Status)
	}
}

func TestSaveFiling_FormMismatch(t *testing.T) {
	store := newMockStore()
	mt := &mockTransport{}
	mt.add("company_tickers.json", 200, mockTickerJSON)
	mt.add("submissions/CIK", 200, mockSubmissionsJSON) // filing is 10-Q
	mt.add("Archives/edgar", 200, "<html>content</html>")

	c := newTestClient(store, mt)
	// Request 10-K but the only available filing is 10-Q.
	err := c.SaveFiling(util.NewTraceContext(), "AAPL", 2024, 1, "10-K", "")
	if err == nil {
		t.Fatal("expected form-mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "10-K") {
		t.Errorf("error %q should mention the requested form", err.Error())
	}
}
