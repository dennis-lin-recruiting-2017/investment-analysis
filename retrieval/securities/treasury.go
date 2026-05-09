package securities

// US Treasury par yield curve retrieval.
//
// Source: US Department of the Treasury daily yield curve XML feed.
// https://home.treasury.gov/resource-center/data-chart-center/interest-rates/
//
// The feed is published on each business day and covers the full par yield
// curve: 1-month through 30-year nominal constant maturity rates.
// No authentication or API key is required.
//
// Coverage note
//
// FINRA TRACE collects Treasury trade reports from broker-dealers (Rule 6730)
// but does NOT publicly disseminate per-trade prices for Treasuries — that
// data is reserved for regulatory use.  The Treasury's own yield curve feed
// is therefore the best free public source for current Treasury rates.

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"investment-analysis/persistence/model"
)

// TreasuryYieldCurve holds the daily par yield curve snapshot published by the
// US Treasury.  All rate fields are annualised percentages (e.g. 4.55 = 4.55%).
//
// Fields may be nil when the Treasury has not published a rate for that maturity
// on the given date (e.g. the 20-year was discontinued Feb 2002 – Feb 2020;
// the 1- and 2-month bills were added later in the series).
type TreasuryYieldCurve struct {
	Date      string   `json:"date"` // YYYY-MM-DD (trading day)
	Source    string   `json:"source"`
	FetchedAt string   `json:"fetched_at"`
	M1Month   *float64 `json:"1_month,omitempty"`
	M2Month   *float64 `json:"2_month,omitempty"`
	M3Month   *float64 `json:"3_month,omitempty"`
	M6Month   *float64 `json:"6_month,omitempty"`
	Y1        *float64 `json:"1_year,omitempty"`
	Y2        *float64 `json:"2_year,omitempty"`
	Y3        *float64 `json:"3_year,omitempty"`
	Y5        *float64 `json:"5_year,omitempty"`
	Y7        *float64 `json:"7_year,omitempty"`
	Y10       *float64 `json:"10_year,omitempty"`
	Y20       *float64 `json:"20_year,omitempty"`
	Y30       *float64 `json:"30_year,omitempty"`
}

const treasuryYieldBase = "https://home.treasury.gov/resource-center/data-chart-center/interest-rates/pages/xml"

// SaveTreasuryYields fetches the most recent US Treasury par yield curve from
// the Treasury's public XML feed and stores a dated snapshot.
//
// Doc key: treasury-yields|<YYYY-MM-DD>
// The date in the key reflects the actual trading date of the data, not the
// fetch date.  Treasury data is published on business days only; if today is
// a weekend or holiday the snapshot will carry the most recent prior
// business-day date.
//
// Calling this more than once for the same trading day returns "skipped_exists"
// without hitting the network again.
func (c *Client) SaveTreasuryYields(ctx context.Context, out string) error {
	curve, sourceURL, err := c.fetchLatestTreasuryYields(ctx)
	if err != nil {
		_ = c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
			DocKey: "treasury-yields|unknown", DocumentType: model.TreasuryYields,
			Ticker: "TREASURY-YIELDS",
			Status: "error", Message: err.Error(),
		})
		return err
	}

	key := fmt.Sprintf("treasury-yields|%s", curve.Date)

	exists, err := c.store.DocumentExists(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
			DocKey: key, DocumentType: model.TreasuryYields, Ticker: "TREASURY-YIELDS",
			Status:  "skipped_exists",
			Message: "Treasury yield curve already stored for " + curve.Date,
		})
	}

	body, err := json.MarshalIndent(curve, "", "  ")
	if err != nil {
		return err
	}

	doc := model.StoredDocument{
		Key:         key,
		DocumentType:  model.TreasuryYields,
		Ticker:      "TREASURY-YIELDS",
		SourceURL:   sourceURL,
		OutputLabel: out,
		MimeType:    "application/json",
		Body:        body,
	}
	if err := c.store.InsertDocument(ctx, doc); err != nil {
		return err
	}
	return c.store.LogRetrievalAttempt(ctx, model.RetrievalAttempt{
		DocKey: key, DocumentType: model.TreasuryYields, Ticker: "TREASURY-YIELDS",
		Status:  "stored",
		Message: fmt.Sprintf("Treasury yield curve stored for %s", curve.Date),
	})
}

// fetchLatestTreasuryYields queries the Treasury XML feed for the current
// month (and the previous month as a fallback) and returns the most recent
// trading-day entry together with the source URL.
func (c *Client) fetchLatestTreasuryYields(ctx context.Context) (TreasuryYieldCurve, string, error) {
	now := time.Now().UTC()

	// Try the current month first; if it has no entries yet (e.g. the first
	// calendar day of the month is a holiday) fall back to the prior month.
	candidates := []time.Time{now, now.AddDate(0, -1, 0)}

	var all []TreasuryYieldCurve
	var lastURL string

	for _, t := range candidates {
		ym := t.Format("200601") // e.g. "202605"
		sourceURL := fmt.Sprintf(
			"%s?data=daily_treasury_yield_curve&field_tdr_date_value=%s",
			treasuryYieldBase, ym,
		)
		lastURL = sourceURL

		raw, err := c.fetchBytes(ctx, sourceURL)
		if err != nil {
			continue
		}

		entries, err := parseTreasuryXML(raw)
		if err != nil || len(entries) == 0 {
			continue
		}

		all = append(all, entries...)
		break // found data; no need to check previous month
	}

	if len(all) == 0 {
		return TreasuryYieldCurve{}, lastURL,
			fmt.Errorf("no Treasury yield curve data found for %s or the prior month", now.Format("2006-01"))
	}

	// Pick the entry with the lexicographically largest date (YYYY-MM-DD).
	latest := all[0]
	for _, e := range all[1:] {
		if e.Date > latest.Date {
			latest = e
		}
	}
	latest.Source = lastURL
	latest.FetchedAt = time.Now().UTC().Format(time.RFC3339)

	return latest, lastURL, nil
}

// ── XML parsing ───────────────────────────────────────────────────────────────

// parseTreasuryXML parses the OData Atom XML feed published by the US Treasury.
//
// The feed uses two XML namespace prefixes:
//
//	d: http://schemas.microsoft.com/ado/2007/08/dataservices   (data values)
//	m: http://schemas.microsoft.com/ado/2007/08/dataservices/metadata
//
// Rather than carrying verbose namespace URIs in every struct tag, this
// function uses a streaming token decoder and matches only on the local
// element name, which is unambiguous within the feed's well-known schema.
func parseTreasuryXML(data []byte) ([]TreasuryYieldCurve, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))

	var curves []TreasuryYieldCurve
	var cur *TreasuryYieldCurve
	var activeField string

	for {
		tok, err := dec.Token()
		if err != nil {
			break // io.EOF or fatal parse error — return what we have
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "entry":
				c := TreasuryYieldCurve{}
				cur = &c
			case "NEW_DATE",
				"BC_1MONTH", "BC_2MONTH", "BC_3MONTH", "BC_6MONTH",
				"BC_1YEAR", "BC_2YEAR", "BC_3YEAR", "BC_5YEAR", "BC_7YEAR",
				"BC_10YEAR", "BC_20YEAR", "BC_30YEAR":
				if cur != nil {
					activeField = t.Name.Local
				}
			}

		case xml.CharData:
			if cur != nil && activeField != "" {
				setTreasuryYield(cur, activeField, strings.TrimSpace(string(t)))
				activeField = ""
			}

		case xml.EndElement:
			if t.Name.Local == "entry" && cur != nil {
				if cur.Date != "" {
					curves = append(curves, *cur)
				}
				cur = nil
			}
		}
	}

	return curves, nil
}

// setTreasuryYield assigns a parsed XML value to the corresponding field of c.
func setTreasuryYield(c *TreasuryYieldCurve, name, value string) {
	if value == "" {
		return
	}

	if name == "NEW_DATE" {
		// Treasury format: "2026-05-01T00:00:00" — keep the date part only.
		if len(value) >= 10 {
			c.Date = value[:10]
		}
		return
	}

	f, err := strconv.ParseFloat(value, 64)
	if err != nil || f == 0 {
		// 0.00 means the rate was not published for this date/maturity.
		return
	}

	switch name {
	case "BC_1MONTH":
		c.M1Month = &f
	case "BC_2MONTH":
		c.M2Month = &f
	case "BC_3MONTH":
		c.M3Month = &f
	case "BC_6MONTH":
		c.M6Month = &f
	case "BC_1YEAR":
		c.Y1 = &f
	case "BC_2YEAR":
		c.Y2 = &f
	case "BC_3YEAR":
		c.Y3 = &f
	case "BC_5YEAR":
		c.Y5 = &f
	case "BC_7YEAR":
		c.Y7 = &f
	case "BC_10YEAR":
		c.Y10 = &f
	case "BC_20YEAR":
		c.Y20 = &f
	case "BC_30YEAR":
		c.Y30 = &f
	}
}
