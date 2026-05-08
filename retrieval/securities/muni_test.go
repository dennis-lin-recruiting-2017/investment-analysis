package securities

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// ── muniTradeFromEntry ────────────────────────────────────────────────────────

func TestMuniTradeFromEntry_Full(t *testing.T) {
	entry := map[string]any{
		"tradeDate":      "2026-05-01",
		"settlementDate": "2026-05-04",
		"price":          float64(98.5),
		"yield":          float64(3.75),
		"quantity":       float64(100000),
		"tradeType":      "Customer Buy",
	}

	tr := muniTradeFromEntry(entry)
	if tr.TradeDate != "2026-05-01" {
		t.Errorf("TradeDate = %q; want 2026-05-01", tr.TradeDate)
	}
	if tr.SettleDate != "2026-05-04" {
		t.Errorf("SettleDate = %q; want 2026-05-04", tr.SettleDate)
	}
	if tr.Price != 98.5 {
		t.Errorf("Price = %v; want 98.5", tr.Price)
	}
	if tr.TradeType != "Customer Buy" {
		t.Errorf("TradeType = %q; want Customer Buy", tr.TradeType)
	}
}

func TestMuniTradeFromEntry_Partial(t *testing.T) {
	entry := map[string]any{
		"tradeDate": "2026-05-01",
		"price":     float64(100.0),
	}
	tr := muniTradeFromEntry(entry)
	if tr.TradeDate == "" {
		t.Error("TradeDate should not be empty")
	}
	if tr.SettleDate != "" {
		t.Errorf("SettleDate should be empty, got %q", tr.SettleDate)
	}
}

// ── SaveMuniBondQuote ─────────────────────────────────────────────────────────

const mockEMMASearchResponse = `[
	{
		"cusip":       "64971WAN0",
		"issuerName":  "State of California",
		"description": "CA GO Bond 3.5% 07/01/2035",
		"maturityDate":"07/01/2035",
		"couponRate":  3.5,
		"taxStatus":   "Federal Tax-Exempt",
		"moodyRating": "Aa2",
		"spRating":    "AA-",
		"fitchRating": "AA",
		"securityId":  "12345"
	}
]`

const mockEMMATradeResponse = `[
	{
		"tradeDate":      "2026-05-01",
		"settlementDate": "2026-05-04",
		"price":          98.5,
		"yield":          3.75,
		"quantity":       100000,
		"tradeType":      "Customer Buy"
	}
]`

func TestSaveMuniBondQuote_AlreadyExists(t *testing.T) {
	store2 := &alwaysExistsStore{mockStore: newMockStore()}
	c := newTestClient(store2, &mockTransport{})

	if err := c.SaveMuniBondQuote(context.Background(), "64971WAN0", "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	last, _ := store2.lastAttempt()
	if last.Status != "skipped_exists" {
		t.Errorf("status = %q; want skipped_exists", last.Status)
	}
}

func TestSaveMuniBondQuote_NoSession_ReferenceDataOnly(t *testing.T) {
	store := newMockStore()
	mt := &mockTransport{}
	mt.add("QuickSearch/SearchAhead", 200, mockEMMASearchResponse)

	c := newTestClient(store, mt)
	if err := c.SaveMuniBondQuote(context.Background(), "64971WAN0", "", "out"); err != nil {
		t.Fatalf("SaveMuniBondQuote error: %v", err)
	}

	// Find the stored document.
	var found []byte
	for k, v := range store.docs {
		if strings.HasPrefix(k, "muni-bond-quote|64971WAN0|") {
			found = v
			break
		}
	}
	if found == nil {
		t.Fatal("no muni-bond-quote document inserted")
	}

	var q MuniBondQuote
	if err := json.Unmarshal(found, &q); err != nil {
		t.Fatalf("stored body is invalid JSON: %v", err)
	}
	if q.CUSIP != "64971WAN0" {
		t.Errorf("CUSIP = %q; want 64971WAN0", q.CUSIP)
	}
	if q.Issuer != "State of California" {
		t.Errorf("Issuer = %q; want State of California", q.Issuer)
	}
	if q.MoodyRating != "Aa2" {
		t.Errorf("MoodyRating = %q; want Aa2", q.MoodyRating)
	}
	if q.CouponRate == nil || *q.CouponRate != 3.5 {
		t.Errorf("CouponRate = %v; want 3.5", q.CouponRate)
	}
	// No session → trade prices not fetched; note must be set.
	if q.Note == "" {
		t.Error("Note should explain that trade prices require a session")
	}
	if len(q.RecentTrades) != 0 {
		t.Errorf("RecentTrades should be empty without session, got %d", len(q.RecentTrades))
	}
}

func TestSaveMuniBondQuote_WithSession_FullData(t *testing.T) {
	store := newMockStore()
	mt := &mockTransport{}
	mt.add("QuickSearch/SearchAhead", 200, mockEMMASearchResponse)
	mt.add("JsonGetRecentTrades", 200, mockEMMATradeResponse)

	c := newTestClient(store, mt)
	if err := c.SaveMuniBondQuote(context.Background(), "64971WAN0", "fake-session-id", ""); err != nil {
		t.Fatalf("SaveMuniBondQuote error: %v", err)
	}

	var found []byte
	for k, v := range store.docs {
		if strings.HasPrefix(k, "muni-bond-quote|64971WAN0|") {
			found = v
			break
		}
	}
	if found == nil {
		t.Fatal("no muni-bond-quote document inserted")
	}

	var q MuniBondQuote
	if err := json.Unmarshal(found, &q); err != nil {
		t.Fatalf("stored body is invalid JSON: %v", err)
	}
	if len(q.RecentTrades) != 1 {
		t.Errorf("RecentTrades length = %d; want 1", len(q.RecentTrades))
	}
	if q.LastPrice == nil || *q.LastPrice != 98.5 {
		t.Errorf("LastPrice = %v; want 98.5", q.LastPrice)
	}
	// No error, so Note should be empty.
	if q.Note != "" {
		t.Errorf("Note should be empty when session is provided, got %q", q.Note)
	}
}

func TestSaveMuniBondQuote_WithSession_TradesFail_StillStoresRefData(t *testing.T) {
	// Trade endpoint returns 403 — reference data should still be stored with a note.
	store := newMockStore()
	mt := &mockTransport{}
	mt.add("QuickSearch/SearchAhead", 200, mockEMMASearchResponse)
	mt.add("JsonGetRecentTrades", 403, "Forbidden")

	c := newTestClient(store, mt)
	if err := c.SaveMuniBondQuote(context.Background(), "64971WAN0", "expired-session", ""); err != nil {
		t.Fatalf("SaveMuniBondQuote should not fail when trade fetch fails: %v", err)
	}

	var found []byte
	for k, v := range store.docs {
		if strings.HasPrefix(k, "muni-bond-quote|64971WAN0|") {
			found = v
			break
		}
	}
	if found == nil {
		t.Fatal("reference data should still be stored even when trades fail")
	}

	var q MuniBondQuote
	if err := json.Unmarshal(found, &q); err != nil {
		t.Fatalf("stored body is invalid JSON: %v", err)
	}
	if q.Issuer == "" {
		t.Error("Issuer should be set from reference data")
	}
	if q.Note == "" {
		t.Error("Note should explain the trade-fetch failure")
	}
}

func TestSaveMuniBondQuote_CUSIPNotFound(t *testing.T) {
	store := newMockStore()
	mt := &mockTransport{}
	mt.add("QuickSearch/SearchAhead", 200, `[]`) // empty results

	c := newTestClient(store, mt)
	err := c.SaveMuniBondQuote(context.Background(), "000000000", "", "")
	if err == nil {
		t.Fatal("expected error for CUSIP not found in EMMA, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error %q should mention 'not found'", err.Error())
	}
}
