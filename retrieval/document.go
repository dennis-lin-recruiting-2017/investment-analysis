package retrieval

// Package retrieval provides a general-purpose HTTP document downloader that
// stores raw response bodies in the application's persistence store.
//
// For structured financial data (SEC EDGAR filings, FINRA TRACE bond quotes,
// EMMA municipal bonds, US Treasury yield curves) use the dedicated methods in
// the retrieval/securities sub-package instead.

import (
	"context"
	"fmt"
	"investment-analysis/util"
	"io"
	"net/http"
	"strings"
	"time"

	"investment-analysis/persistence/model"
)

const DefaultUserAgent = "go-download-web-archive/1.0 (contact: you@example.com)"

// Client downloads documents and stores them via the documents and
// retrieval-attempts tables.  Use NewClient for plain HTTP fetching or
// NewPlaywrightClient to drive a real Firefox browser (required for
// JS-heavy or bot-protected pages).  The two store arguments accept any
// implementation of DocumentsStore / AttemptsStore — typically a
// *model.Store's Documents and RetrievalAttempts fields.
type Client struct {
	userAgent  string
	httpClient *http.Client
	docs       DocumentsStore
	attempts   AttemptsStore
	pw         *playwrightFetcher // nil → use httpClient
}

// NewClient returns a Client that fetches pages with a plain HTTP request.
// If userAgent is blank the package DefaultUserAgent is used.
func NewClient(userAgent string, docs DocumentsStore, attempts AttemptsStore) *Client {
	if strings.TrimSpace(userAgent) == "" {
		userAgent = DefaultUserAgent
	}
	return &Client{
		userAgent:  userAgent,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		docs:       docs,
		attempts:   attempts,
	}
}

// NewPlaywrightClient returns a Client that fetches pages by driving a real
// Firefox browser via Playwright.  Use this for sites that require JavaScript
// execution or bot-detection bypass (e.g. Akamai-protected pages).
//
// timeoutSeconds is the maximum total time allowed for a single page fetch
// (navigation + DOM stability wait combined).  Pass the value from the
// retrieval_settings table (store.RetrievalSettings.Get) or use
// model.DefaultRetrievalSettings().PlaywrightTimeoutSeconds
// for the default of 300 s.
//
// The caller must call Close() when done to shut down the browser process.
func NewPlaywrightClient(userAgent string, docs DocumentsStore, attempts AttemptsStore, timeoutSeconds int) (c *Client, err error) {
	defer func() {
		util.LogIfErr(nil, &err, "retrieval.NewPlaywrightClient", "userAgent", userAgent, "timeoutSeconds", timeoutSeconds)
	}()
	if strings.TrimSpace(userAgent) == "" {
		userAgent = DefaultUserAgent
	}
	pw, e := newPlaywrightFetcher(timeoutSeconds)
	if e != nil {
		return nil, e
	}
	return &Client{
		userAgent: userAgent,
		docs:      docs,
		attempts:  attempts,
		pw:        pw,
	}, nil
}

// Close releases resources held by the client.  It must be called when using
// NewPlaywrightClient; it is safe but a no-op for plain HTTP clients.
func (c *Client) Close() (err error) {
	defer func() { util.LogIfErr(nil, &err, "retrieval.Client.Close") }()
	if c.pw != nil {
		return c.pw.close()
	}
	return nil
}

// SaveDocument fetches url and stores the raw response body in the persistence
// store under key.
//
// If a document with key already exists the function logs "skipped_exists" and
// returns nil without hitting the network again.
//
// mode is a short identifier for the kind of document (e.g. "press-release",
// "annual-report"); it is stored alongside the document for later retrieval.
// outputLabel is an optional human-readable label.
func (c *Client) SaveDocument(ctx context.Context, url, key string, documentType model.DocumentType, outputLabel string) (err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "retrieval.Client.SaveDocument",
			"url", url, "key", key, "documentType", documentType, "outputLabel", outputLabel)
	}()
	exists, err := c.docs.Exists(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return c.attempts.Log(ctx, model.RetrievalAttempt{
			DocKey:       key,
			DocumentType: documentType,
			Status:       "skipped_exists",
			Message:      "document already exists; retrieval skipped",
		})
	}

	body, mimeType, err := c.fetch(ctx, url)
	if err != nil {
		_ = c.attempts.Log(ctx, model.RetrievalAttempt{
			DocKey:       key,
			DocumentType: documentType,
			Status:       "error",
			Message:      err.Error(),
		})
		return err
	}

	doc := model.Document{
		DocKey:       key,
		DocumentType: documentType,
		SourceURL:    url,
		OutputLabel:  outputLabel,
		MimeType:     mimeType,
		Body:         body,
	}
	if err := c.docs.Insert(ctx, &doc); err != nil {
		return err
	}
	return c.attempts.Log(ctx, model.RetrievalAttempt{
		DocKey:       key,
		DocumentType: documentType,
		Status:       "stored",
		Message:      fmt.Sprintf("stored %d bytes from %s at %s", len(body), url, time.Now().UTC().Format(time.RFC3339)),
	})
}

// fetch retrieves url and returns the raw body and MIME type.
// When the client was created with NewPlaywrightClient it drives a real
// Firefox browser; otherwise it issues a plain HTTP GET.
func (c *Client) fetch(ctx context.Context, url string) (body []byte, mimeType string, err error) {
	defer func() { util.LogIfErr(ctx, &err, "retrieval.Client.fetch", "url", url) }()
	if c.pw != nil {
		return c.pw.fetch(ctx, url)
	}
	return c.httpFetch(ctx, url)
}

// httpFetch performs a single GET request and returns the body and Content-Type.
func (c *Client) httpFetch(ctx context.Context, url string) (body []byte, mimeType string, err error) {
	defer func() { util.LogIfErr(ctx, &err, "retrieval.Client.httpFetch", "url", url) }()
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
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, "", fmt.Errorf("HTTP %s from %s: %s",
			resp.Status, url, strings.TrimSpace(string(snippet)))
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	mimeType = resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return body, mimeType, nil
}
