package securities

import (
	"context"
	"encoding/json"
	"investment-analysis/util"
	"strings"
	"testing"
)

// ── strField ──────────────────────────────────────────────────────────────────

func TestStrField(t *testing.T) {
	m := map[string]any{
		"present": "hello",
		"empty":   "",
		"num":     42.0,
	}

	if v, ok := strField(m, "present"); !ok || v != "hello" {
		t.Errorf("strField(present) = (%q, %v); want (hello, true)", v, ok)
	}
	if _, ok := strField(m, "missing"); ok {
		t.Error("strField(missing): expected false")
	}
	if _, ok := strField(m, "empty"); ok {
		t.Error("strField(empty): empty string should return false")
	}
	if _, ok := strField(m, "num"); ok {
		t.Error("strField(num): non-string should return false")
	}
}

// ── numField ──────────────────────────────────────────────────────────────────

func TestNumField(t *testing.T) {
	m := map[string]any{
		"float": float64(3.14),
		"int":   42,
		"zero":  float64(0),
		"str":   "nope",
	}

	if v, ok := numField(m, "float"); !ok || v != 3.14 {
		t.Errorf("numField(float) = (%v, %v); want (3.14, true)", v, ok)
	}
	if v, ok := numField(m, "int"); !ok || v != 42 {
		t.Errorf("numField(int) = (%v, %v); want (42, true)", v, ok)
	}
	if _, ok := numField(m, "missing"); ok {
		t.Error("numField(missing): expected false")
	}
	if _, ok := numField(m, "zero"); ok {
		t.Error("numField(zero): zero value should return false")
	}
	if _, ok := numField(m, "str"); ok {
		t.Error("numField(str): non-numeric should return false")
	}
}

// ── bondTradeFromEntry ────────────────────────────────────────────────────────

func TestBondTradeFromEntry_Full(t *testing.T) {
	entry := map[string]any{
		"tradeDate":     "2026-05-01",
		"tradeTime":     "10:30:15",
		"lastSalePrice": float64(95.5),
		"yield":         float64(4.125),
		"quantity":      float64(250000),
		"side":          "S",
	}

	tr := bondTradeFromEntry(entry)
	if tr.TradeDate != "2026-05-01" {
		t.Errorf("TradeDate = %q; want 2026-05-01", tr.TradeDate)
	}
	if tr.Price != 95.5 {
		t.Errorf("Price = %v; want 95.5", tr.Price)
	}
	if tr.Yield != 4.125 {
		t.Errorf("Yield = %v; want 4.125", tr.Yield)
	}
	if tr.Side != "S" {
		t.Errorf("Side = %q; want S", tr.Side)
	}
}

func TestBondTradeFromEntry_FallbackPriceKey(t *testing.T) {
	// Some FINRA feed versions use "price" instead of "lastSalePrice".
	entry := map[string]any{
		"tradeDate": "2026-05-01",
		"price":     float64(98.0),
	}
	tr := bondTradeFromEntry(entry)
	if tr.Price != 98.0 {
		t.Errorf("Price (fallback) = %v; want 98.0", tr.Price)
	}
}

// ── SaveBondQuote ─────────────────────────────────────────────────────────────

const mockFINRAResponse = `{
	"data": [
		{
			"bondSymbol":       "38259P508",
			"issueDescription": "Apple Inc 3.85% Notes Due 2043",
			"tradeDate":        "2026-05-01",
			"tradeTime":        "10:30:15",
			"lastSalePrice":    95.5,
			"yield":            4.125,
			"quantity":         250000,
			"side":             "S",
			"moodyRating":      "Aaa",
			"spRating":         "AA+",
			"couponRate":       3.85,
			"maturityDate":     "2043-05-04"
		}
	]
}`

func TestSaveBondQuote_AlreadyExists(t *testing.T) {
	s := &alwaysExistsStore{mockStore: newMockStore()}
	c := newTestClient(s, &mockTransport{})

	if err := c.SaveBondQuote(util.NewTraceContext(), "38259P508", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	last, _ := s.lastAttempt()
	if last.Status != "skipped_exists" {
		t.Errorf("status = %q; want skipped_exists", last.Status)
	}
}

func TestSaveBondQuote_Success(t *testing.T) {
	store := newMockStore()
	mt := &mockTransport{}
	mt.add("getTradeActivity", 200, mockFINRAResponse)

	c := newTestClient(store, mt)
	if err := c.SaveBondQuote(util.NewTraceContext(), "38259p508", "out"); err != nil {
		t.Fatalf("SaveBondQuote error: %v", err)
	}

	// Verify a document was stored.
	var found []byte
	for k, v := range store.docs {
		if strings.HasPrefix(k, "bond-quote|38259P508|") {
			found = v
			break
		}
	}
	if found == nil {
		t.Fatal("no bond-quote document inserted for CUSIP 38259P508")
	}

	var q BondQuote
	if err := json.Unmarshal(found, &q); err != nil {
		t.Fatalf("stored body is invalid JSON: %v", err)
	}
	if q.CUSIP != "38259P508" {
		t.Errorf("CUSIP = %q; want 38259P508", q.CUSIP)
	}
	if q.LastPrice == nil || *q.LastPrice != 95.5 {
		t.Errorf("LastPrice = %v; want 95.5", q.LastPrice)
	}
	if len(q.RecentTrades) != 1 {
		t.Errorf("RecentTrades length = %d; want 1", len(q.RecentTrades))
	}

	last, _ := store.lastAttempt()
	if last.Status != "stored" {
		t.Errorf("final attempt status = %q; want stored", last.Status)
	}
}

func TestSaveBondQuote_NoTrades(t *testing.T) {
	store := newMockStore()
	mt := &mockTransport{}
	mt.add("getTradeActivity", 200, `{"data":[]}`)

	c := newTestClient(store, mt)
	err := c.SaveBondQuote(util.NewTraceContext(), "BADCUSIP0", "")
	if err == nil {
		t.Fatal("expected error for empty trade data, got nil")
	}
	if !strings.Contains(err.Error(), "no TRACE data") {
		t.Errorf("error %q; expected 'no TRACE data'", err.Error())
	}
}

func TestSaveBondQuote_HTTPError(t *testing.T) {
	store := newMockStore()
	mt := &mockTransport{}
	mt.add("getTradeActivity", 500, "internal server error")

	c := newTestClient(store, mt)
	err := c.SaveBondQuote(util.NewTraceContext(), "38259P508", "")
	if err == nil {
		t.Fatal("expected error on 500 response, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error %q should mention HTTP status 500", err.Error())
	}
}

// ── alwaysExistsStore helper ──────────────────────────────────────────────────

// alwaysExistsStore wraps a mockStore but always reports documents as existing.
type alwaysExistsStore struct{ *mockStore }

func (s *alwaysExistsStore) Exists(_ context.Context, _ string) (bool, error) {
	return true, nil
}
