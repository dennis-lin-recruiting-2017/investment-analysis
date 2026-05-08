package securities

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// ── parseTreasuryXML ──────────────────────────────────────────────────────────

// minimalTreasuryXML builds a minimal but structurally correct OData Atom
// feed that matches what the US Treasury actually publishes.
func makeTreasuryXML(entries ...string) []byte {
	const header = `<?xml version="1.0" encoding="utf-8" standalone="yes" ?>` +
		`<feed xml:base="https://home.treasury.gov/"` +
		` xmlns:d="http://schemas.microsoft.com/ado/2007/08/dataservices"` +
		` xmlns:m="http://schemas.microsoft.com/ado/2007/08/dataservices/metadata"` +
		` xmlns="http://www.w3.org/2005/Atom">`
	const footer = `</feed>`

	var b strings.Builder
	b.WriteString(header)
	for _, e := range entries {
		b.WriteString(e)
	}
	b.WriteString(footer)
	return []byte(b.String())
}

func treasuryEntry(date, m3, m10, m30 string) string {
	return `<entry><content type="application/xml"><m:properties>` +
		`<d:NEW_DATE m:type="Edm.DateTime">` + date + `T00:00:00</d:NEW_DATE>` +
		`<d:BC_3MONTH m:type="Edm.Double">` + m3 + `</d:BC_3MONTH>` +
		`<d:BC_10YEAR m:type="Edm.Double">` + m10 + `</d:BC_10YEAR>` +
		`<d:BC_30YEAR m:type="Edm.Double">` + m30 + `</d:BC_30YEAR>` +
		`</m:properties></content></entry>`
}

func fullTreasuryEntry(date string) string {
	return `<entry><content type="application/xml"><m:properties>` +
		`<d:NEW_DATE m:type="Edm.DateTime">` + date + `T00:00:00</d:NEW_DATE>` +
		`<d:BC_1MONTH m:type="Edm.Double">5.31</d:BC_1MONTH>` +
		`<d:BC_2MONTH m:type="Edm.Double">5.28</d:BC_2MONTH>` +
		`<d:BC_3MONTH m:type="Edm.Double">5.25</d:BC_3MONTH>` +
		`<d:BC_6MONTH m:type="Edm.Double">5.10</d:BC_6MONTH>` +
		`<d:BC_1YEAR m:type="Edm.Double">4.95</d:BC_1YEAR>` +
		`<d:BC_2YEAR m:type="Edm.Double">4.75</d:BC_2YEAR>` +
		`<d:BC_3YEAR m:type="Edm.Double">4.65</d:BC_3YEAR>` +
		`<d:BC_5YEAR m:type="Edm.Double">4.55</d:BC_5YEAR>` +
		`<d:BC_7YEAR m:type="Edm.Double">4.50</d:BC_7YEAR>` +
		`<d:BC_10YEAR m:type="Edm.Double">4.45</d:BC_10YEAR>` +
		`<d:BC_20YEAR m:type="Edm.Double">4.60</d:BC_20YEAR>` +
		`<d:BC_30YEAR m:type="Edm.Double">4.65</d:BC_30YEAR>` +
		`</m:properties></content></entry>`
}

func TestParseTreasuryXML_SingleEntry_AllFields(t *testing.T) {
	raw := makeTreasuryXML(fullTreasuryEntry("2026-05-01"))
	curves, err := parseTreasuryXML(raw)
	if err != nil {
		t.Fatalf("parseTreasuryXML error: %v", err)
	}
	if len(curves) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(curves))
	}
	c := curves[0]
	if c.Date != "2026-05-01" {
		t.Errorf("Date = %q; want 2026-05-01", c.Date)
	}
	if c.M1Month == nil || *c.M1Month != 5.31 {
		t.Errorf("M1Month = %v; want 5.31", c.M1Month)
	}
	if c.Y10 == nil || *c.Y10 != 4.45 {
		t.Errorf("Y10 = %v; want 4.45", c.Y10)
	}
	if c.Y30 == nil || *c.Y30 != 4.65 {
		t.Errorf("Y30 = %v; want 4.65", c.Y30)
	}
}

func TestParseTreasuryXML_MultipleEntries(t *testing.T) {
	raw := makeTreasuryXML(
		treasuryEntry("2026-05-01", "5.25", "4.45", "4.65"),
		treasuryEntry("2026-05-02", "5.20", "4.40", "4.60"),
	)
	curves, err := parseTreasuryXML(raw)
	if err != nil {
		t.Fatalf("parseTreasuryXML error: %v", err)
	}
	if len(curves) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(curves))
	}
	if curves[0].Date != "2026-05-01" || curves[1].Date != "2026-05-02" {
		t.Errorf("dates = %q, %q; want 2026-05-01, 2026-05-02",
			curves[0].Date, curves[1].Date)
	}
}

func TestParseTreasuryXML_MissingOptionalFields(t *testing.T) {
	// No 1-month or 20-year — should be nil, not zero.
	raw := makeTreasuryXML(treasuryEntry("2026-05-01", "5.25", "4.45", "4.65"))
	curves, _ := parseTreasuryXML(raw)
	if len(curves) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(curves))
	}
	if curves[0].M1Month != nil {
		t.Errorf("M1Month should be nil, got %v", curves[0].M1Month)
	}
	if curves[0].Y20 != nil {
		t.Errorf("Y20 should be nil, got %v", curves[0].Y20)
	}
}

func TestParseTreasuryXML_ZeroRatesTreatedAsNil(t *testing.T) {
	// The Treasury uses 0.00 to signal "not published". Those should be nil.
	raw := makeTreasuryXML(
		`<entry><content type="application/xml"><m:properties>` +
			`<d:NEW_DATE>2026-05-01T00:00:00</d:NEW_DATE>` +
			`<d:BC_10YEAR>4.45</d:BC_10YEAR>` +
			`<d:BC_20YEAR>0.00</d:BC_20YEAR>` + // discontinued → 0
			`</m:properties></content></entry>`,
	)
	curves, _ := parseTreasuryXML(raw)
	if len(curves) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(curves))
	}
	if curves[0].Y20 != nil {
		t.Errorf("Y20 with 0.00 should be nil, got %v", *curves[0].Y20)
	}
	if curves[0].Y10 == nil || *curves[0].Y10 != 4.45 {
		t.Errorf("Y10 = %v; want 4.45", curves[0].Y10)
	}
}

func TestParseTreasuryXML_Empty(t *testing.T) {
	raw := makeTreasuryXML() // no entries
	curves, err := parseTreasuryXML(raw)
	if err != nil {
		t.Fatalf("parseTreasuryXML on empty feed should not error: %v", err)
	}
	if len(curves) != 0 {
		t.Errorf("expected 0 entries, got %d", len(curves))
	}
}

func TestParseTreasuryXML_InvalidXML(t *testing.T) {
	// Malformed XML: parser should return what it has (possibly empty), not panic.
	curves, _ := parseTreasuryXML([]byte(`<feed><entry><BROKEN`))
	// We don't assert an error — the token loop just stops at the broken token.
	// The important thing is no panic and the result is either empty or partial.
	_ = curves
}

// ── setTreasuryYield ──────────────────────────────────────────────────────────

func TestSetTreasuryYield_DateParsing(t *testing.T) {
	var c TreasuryYieldCurve
	setTreasuryYield(&c, "NEW_DATE", "2026-05-01T00:00:00")
	if c.Date != "2026-05-01" {
		t.Errorf("Date = %q; want 2026-05-01", c.Date)
	}
}

func TestSetTreasuryYield_EachMaturity(t *testing.T) {
	fields := []struct {
		name  string
		value string
		get   func(*TreasuryYieldCurve) *float64
	}{
		{"BC_1MONTH", "5.31", func(c *TreasuryYieldCurve) *float64 { return c.M1Month }},
		{"BC_2MONTH", "5.28", func(c *TreasuryYieldCurve) *float64 { return c.M2Month }},
		{"BC_3MONTH", "5.25", func(c *TreasuryYieldCurve) *float64 { return c.M3Month }},
		{"BC_6MONTH", "5.10", func(c *TreasuryYieldCurve) *float64 { return c.M6Month }},
		{"BC_1YEAR", "4.95", func(c *TreasuryYieldCurve) *float64 { return c.Y1 }},
		{"BC_2YEAR", "4.75", func(c *TreasuryYieldCurve) *float64 { return c.Y2 }},
		{"BC_3YEAR", "4.65", func(c *TreasuryYieldCurve) *float64 { return c.Y3 }},
		{"BC_5YEAR", "4.55", func(c *TreasuryYieldCurve) *float64 { return c.Y5 }},
		{"BC_7YEAR", "4.50", func(c *TreasuryYieldCurve) *float64 { return c.Y7 }},
		{"BC_10YEAR", "4.45", func(c *TreasuryYieldCurve) *float64 { return c.Y10 }},
		{"BC_20YEAR", "4.60", func(c *TreasuryYieldCurve) *float64 { return c.Y20 }},
		{"BC_30YEAR", "4.65", func(c *TreasuryYieldCurve) *float64 { return c.Y30 }},
	}
	for _, f := range fields {
		var c TreasuryYieldCurve
		setTreasuryYield(&c, f.name, f.value)
		ptr := f.get(&c)
		if ptr == nil {
			t.Errorf("%s: field is nil after set", f.name)
		}
	}
}

// ── SaveTreasuryYields ────────────────────────────────────────────────────────

func makeTreasuryXMLResponse(date string) string {
	return string(makeTreasuryXML(fullTreasuryEntry(date)))
}

func TestSaveTreasuryYields_AlreadyExists(t *testing.T) {
	// SaveTreasuryYields fetches data first to determine the date key, so the
	// mock transport must return valid XML even for the already-exists case.
	s := &alwaysExistsStore{mockStore: newMockStore()}
	mt := &mockTransport{}
	mt.add("daily_treasury_yield_curve", 200, makeTreasuryXMLResponse("2026-05-01"))

	c := newTestClient(s, mt)
	if err := c.SaveTreasuryYields(context.Background(), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	last, _ := s.lastAttempt()
	if last.Status != "skipped_exists" {
		t.Errorf("status = %q; want skipped_exists", last.Status)
	}
}

func TestSaveTreasuryYields_Success(t *testing.T) {
	store := newMockStore()
	mt := &mockTransport{}
	mt.add("daily_treasury_yield_curve", 200, makeTreasuryXMLResponse("2026-05-01"))

	c := newTestClient(store, mt)
	if err := c.SaveTreasuryYields(context.Background(), "out"); err != nil {
		t.Fatalf("SaveTreasuryYields error: %v", err)
	}

	var found []byte
	for k, v := range store.docs {
		if strings.HasPrefix(k, "treasury-yields|") {
			found = v
			break
		}
	}
	if found == nil {
		t.Fatal("no treasury-yields document inserted")
	}

	var curve TreasuryYieldCurve
	if err := json.Unmarshal(found, &curve); err != nil {
		t.Fatalf("stored body is invalid JSON: %v", err)
	}
	if curve.Date != "2026-05-01" {
		t.Errorf("Date = %q; want 2026-05-01", curve.Date)
	}
	if curve.Y10 == nil || *curve.Y10 != 4.45 {
		t.Errorf("Y10 = %v; want 4.45", curve.Y10)
	}

	last, _ := store.lastAttempt()
	if last.Status != "stored" {
		t.Errorf("final attempt status = %q; want stored", last.Status)
	}
}

func TestSaveTreasuryYields_FallbackToPreviousMonth(t *testing.T) {
	// Current month returns empty feed; previous month has data.
	store := newMockStore()
	mt := &mockTransport{}

	callCount := 0
	mt.routes = nil // use a custom handler
	customMT := &funcTransport{fn: func(req *http.Request) (string, int, string) {
		callCount++
		url := req.URL.String()
		if strings.Contains(url, "daily_treasury_yield_curve") {
			if callCount == 1 {
				// First call (current month) → empty
				return url, 200, string(makeTreasuryXML())
			}
			// Second call (previous month) → has data
			return url, 200, makeTreasuryXMLResponse("2026-04-30")
		}
		return url, 404, "not found"
	}}

	c := newTestClient(store, customMT)
	if err := c.SaveTreasuryYields(context.Background(), ""); err != nil {
		t.Fatalf("SaveTreasuryYields fallback error: %v", err)
	}

	var found []byte
	for k, v := range store.docs {
		if strings.HasPrefix(k, "treasury-yields|") {
			found = v
			break
		}
	}
	if found == nil {
		t.Fatal("no document inserted after fallback to previous month")
	}

	var curve TreasuryYieldCurve
	json.Unmarshal(found, &curve) //nolint:errcheck
	if curve.Date != "2026-04-30" {
		t.Errorf("Date = %q; want 2026-04-30 (from previous month fallback)", curve.Date)
	}
}

func TestSaveTreasuryYields_NoData(t *testing.T) {
	store := newMockStore()
	mt := &mockTransport{}
	// Both months return empty feeds.
	mt.add("daily_treasury_yield_curve", 200, string(makeTreasuryXML()))

	c := newTestClient(store, mt)
	err := c.SaveTreasuryYields(context.Background(), "")
	if err == nil {
		t.Fatal("expected error when both months have no data, got nil")
	}
	if !strings.Contains(err.Error(), "no Treasury yield curve data") {
		t.Errorf("error %q should mention 'no Treasury yield curve data'", err.Error())
	}
}

func TestSaveTreasuryYields_PicksMostRecentEntry(t *testing.T) {
	// Feed contains two entries; the most recent should be stored.
	store := newMockStore()
	mt := &mockTransport{}
	xml := makeTreasuryXML(
		treasuryEntry("2026-05-01", "5.25", "4.45", "4.65"),
		treasuryEntry("2026-05-02", "5.20", "4.40", "4.60"),
	)
	mt.add("daily_treasury_yield_curve", 200, string(xml))

	c := newTestClient(store, mt)
	if err := c.SaveTreasuryYields(context.Background(), ""); err != nil {
		t.Fatalf("SaveTreasuryYields error: %v", err)
	}

	expectedKey := "treasury-yields|2026-05-02"
	if _, ok := store.docs[expectedKey]; !ok {
		t.Errorf("expected document with key %q, got keys: %v", expectedKey, keys(store.docs))
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func keys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// funcTransport is a test transport that delegates to a function.
type funcTransport struct {
	fn func(*http.Request) (url string, status int, body string)
}

func (f *funcTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	_, status, body := f.fn(req)
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}, nil
}
