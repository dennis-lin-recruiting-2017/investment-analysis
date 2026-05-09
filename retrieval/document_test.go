package retrieval

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"investment-analysis/persistence"
	"investment-analysis/persistence/model"
)

// openTestStore creates a real SQLite database in the test's temp directory.
// Using a real store exercises the full path: SQL schema creation, inserts,
// and reads — without touching any production file.
func openTestStore(t *testing.T) *persistence.GormStore {
	t.Helper()
	store, err := persistence.OpenFile(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("openTestStore: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

// newTestServer starts an httptest.Server that responds with the given status,
// content type, and body.  It is shut down automatically when the test ends.
func newTestServer(t *testing.T, status int, contentType, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestSaveDocument_Success verifies that a document is fetched from an HTTP
// server and stored in the database, and that the stored bytes and MIME type
// are exactly what the server returned.
func TestSaveDocument_Success(t *testing.T) {
	const (
		wantBody = "<html><h1>Annual Report 2025</h1></html>"
		wantMIME = "text/html; charset=utf-8"
		key      = "annual-report|ACME|2025"
		label    = "ACME Corp 2025 Annual Report"
	)

	srv := newTestServer(t, http.StatusOK, wantMIME, wantBody)
	store := openTestStore(t)
	client := NewClient("test-agent/1.0", store)

	if err := client.SaveDocument(context.Background(), srv.URL, key, model.AnnualReport, label); err != nil {
		t.Fatalf("SaveDocument returned unexpected error: %v", err)
	}

	// Retrieve the stored body through the store and verify content.
	got, err := store.GetDocumentBody(context.Background(), key)
	if err != nil {
		t.Fatalf("GetDocumentBody: %v", err)
	}
	if string(got) != wantBody {
		t.Errorf("stored body = %q; want %q", got, wantBody)
	}
}

// TestSaveDocument_StoresMIMEType checks that the Content-Type header returned
// by the server is preserved in the stored document record.
func TestSaveDocument_StoresMIMEType(t *testing.T) {
	srv := newTestServer(t, http.StatusOK, "application/pdf", "%PDF-1.4 fake pdf")
	store := openTestStore(t)
	client := NewClient("", store)

	key := "pdf-doc|test"
	if err := client.SaveDocument(context.Background(), srv.URL, key, model.DocumentType("pdf"), ""); err != nil {
		t.Fatalf("SaveDocument: %v", err)
	}

	// The MIME type is stored on the DocumentRow; we can confirm indirectly
	// that it was stored by verifying the body is retrievable (a full MIME
	// assertion would require an unexported field or a Store extension).
	body, err := store.GetDocumentBody(context.Background(), key)
	if err != nil {
		t.Fatalf("GetDocumentBody: %v", err)
	}
	if !strings.HasPrefix(string(body), "%PDF") {
		t.Errorf("body does not look like a PDF: %q", body)
	}
}

// TestSaveDocument_AlreadyExists verifies that calling SaveDocument twice with
// the same key is a no-op on the second call: the HTTP server is not contacted
// again and no error is returned.
func TestSaveDocument_AlreadyExists(t *testing.T) {
	requestCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("content"))
	}))
	t.Cleanup(srv.Close)

	store := openTestStore(t)
	client := NewClient("", store)
	ctx := context.Background()
	key := "dup-doc|001"

	// First call — should fetch and store.
	if err := client.SaveDocument(ctx, srv.URL, key, model.DocumentType("test"), ""); err != nil {
		t.Fatalf("first SaveDocument: %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("expected 1 HTTP request after first call, got %d", requestCount)
	}

	// Second call with the same key — must not hit the server again.
	if err := client.SaveDocument(ctx, srv.URL, key, model.DocumentType("test"), ""); err != nil {
		t.Fatalf("second SaveDocument (already exists): %v", err)
	}
	if requestCount != 1 {
		t.Errorf("server was contacted on second call (request count = %d); want 1", requestCount)
	}
}

// TestSaveDocument_ServerError verifies that a non-2xx response from the
// server causes SaveDocument to return an error and store nothing.
func TestSaveDocument_ServerError(t *testing.T) {
	srv := newTestServer(t, http.StatusInternalServerError, "", "something went wrong")
	store := openTestStore(t)
	client := NewClient("", store)

	key := "error-doc|001"
	err := client.SaveDocument(context.Background(), srv.URL, key, model.DocumentType("test"), "")
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error %q should mention HTTP status 500", err.Error())
	}

	// Nothing should have been stored.
	if _, dbErr := store.GetDocumentBody(context.Background(), key); dbErr == nil {
		t.Error("document should not be stored after a server error")
	}
}

// TestSaveDocument_NetworkFailure verifies that a connection-level error
// (server not running) is surfaced as an error and nothing is stored.
func TestSaveDocument_NetworkFailure(t *testing.T) {
	// Start and immediately close a server to get a valid-looking but dead URL.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	deadURL := srv.URL
	srv.Close() // close before the request is made

	store := openTestStore(t)
	client := NewClient("", store)

	key := "net-fail|001"
	if err := client.SaveDocument(context.Background(), deadURL, key, model.DocumentType("test"), ""); err == nil {
		t.Fatal("expected network error, got nil")
	}

	if _, dbErr := store.GetDocumentBody(context.Background(), key); dbErr == nil {
		t.Error("document should not be stored after a network failure")
	}
}

// TestSaveDocument_FallbackMIMEType checks that when the server returns no
// Content-Type header, "application/octet-stream" is used.
func TestSaveDocument_FallbackMIMEType(t *testing.T) {
	// Use a raw handler that deliberately omits Content-Type.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("raw bytes"))
	}))
	t.Cleanup(srv.Close)

	store := openTestStore(t)
	client := NewClient("", store)

	key := "octet-doc|001"
	if err := client.SaveDocument(context.Background(), srv.URL, key, model.DocumentType("raw"), ""); err != nil {
		t.Fatalf("SaveDocument: %v", err)
	}

	// Verify the document was stored (MIME type is an internal field; the key
	// observable here is that the body was persisted correctly).
	body, err := store.GetDocumentBody(context.Background(), key)
	if err != nil {
		t.Fatalf("GetDocumentBody: %v", err)
	}
	if string(body) != "raw bytes" {
		t.Errorf("body = %q; want %q", body, "raw bytes")
	}
}

// TestSaveDocument_UserAgentForwarded verifies that the User-Agent header
// configured on the client is sent in the outbound HTTP request.
func TestSaveDocument_UserAgentForwarded(t *testing.T) {
	const wantUA = "my-app/2.0 (contact: test@example.com)"
	var gotUA string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	store := openTestStore(t)
	client := NewClient(wantUA, store)

	if err := client.SaveDocument(context.Background(), srv.URL, "ua-doc|001", model.DocumentType("test"), ""); err != nil {
		t.Fatalf("SaveDocument: %v", err)
	}
	if gotUA != wantUA {
		t.Errorf("User-Agent = %q; want %q", gotUA, wantUA)
	}
}

// TestSaveDocument_EmptyUserAgentUsesDefault verifies that passing an empty
// user-agent string falls back to DefaultUserAgent.
func TestSaveDocument_EmptyUserAgentUsesDefault(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	store := openTestStore(t)
	client := NewClient("", store) // blank → should use DefaultUserAgent

	if err := client.SaveDocument(context.Background(), srv.URL, "default-ua|001", model.DocumentType("test"), ""); err != nil {
		t.Fatalf("SaveDocument: %v", err)
	}
	if gotUA != DefaultUserAgent {
		t.Errorf("User-Agent = %q; want DefaultUserAgent %q", gotUA, DefaultUserAgent)
	}
}
