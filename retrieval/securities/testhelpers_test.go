package securities

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"investment-analysis/persistence/model"
	"investment-analysis/retrieval"
)

// ── mock store ────────────────────────────────────────────────────────────────

// mockStore satisfies both retrieval.DocumentsStore and
// retrieval.AttemptsStore using in-memory maps; the same instance can
// be passed for both arguments to NewClient.
type mockStore struct {
	docs     map[string][]byte
	attempts []model.RetrievalAttempt

	// Inject errors to simulate storage failures.
	existsErr  error
	insertErr  error
	logErr     error
	getBodyErr error
}

func newMockStore() *mockStore {
	return &mockStore{docs: make(map[string][]byte)}
}

func (s *mockStore) Exists(_ context.Context, key string) (bool, error) {
	if s.existsErr != nil {
		return false, s.existsErr
	}
	_, ok := s.docs[key]
	return ok, nil
}

func (s *mockStore) Insert(_ context.Context, doc *model.Document) error {
	if s.insertErr != nil {
		return s.insertErr
	}
	s.docs[doc.DocKey] = doc.Body
	return nil
}

func (s *mockStore) Log(_ context.Context, a model.RetrievalAttempt) error {
	s.attempts = append(s.attempts, a)
	return s.logErr
}

func (s *mockStore) GetBody(_ context.Context, key string) ([]byte, error) {
	if s.getBodyErr != nil {
		return nil, s.getBodyErr
	}
	body, ok := s.docs[key]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", key)
	}
	return body, nil
}

func (s *mockStore) lastAttempt() (model.RetrievalAttempt, bool) {
	if len(s.attempts) == 0 {
		return model.RetrievalAttempt{}, false
	}
	return s.attempts[len(s.attempts)-1], true
}

// ── mock HTTP transport ───────────────────────────────────────────────────────

// mockRoute maps a URL substring to a fixed response.
type mockRoute struct {
	contains   string
	statusCode int
	body       string
}

// mockTransport intercepts outbound HTTP requests by matching URL substrings.
// Routes are evaluated in registration order; the first match wins.
type mockTransport struct {
	routes []mockRoute
	calls  []string // recorded request URLs, for assertions
}

func (m *mockTransport) add(contains string, status int, body string) {
	m.routes = append(m.routes, mockRoute{contains, status, body})
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rawURL := req.URL.String()
	m.calls = append(m.calls, rawURL)
	for _, r := range m.routes {
		if strings.Contains(rawURL, r.contains) {
			return &http.Response{
				StatusCode: r.statusCode,
				Status:     fmt.Sprintf("%d %s", r.statusCode, http.StatusText(r.statusCode)),
				Body:       io.NopCloser(strings.NewReader(r.body)),
				Header:     make(http.Header),
			}, nil
		}
	}
	return nil, fmt.Errorf("mockTransport: no route matched %q", rawURL)
}

// ── client factory ────────────────────────────────────────────────────────────

// combinedStore is satisfied by anything that implements both halves of
// the retrieval contract.  *mockStore and *alwaysExistsStore qualify.
type combinedStore interface {
	retrieval.DocumentsStore
	retrieval.AttemptsStore
}

// newTestClient wires a Client to use t for all outbound HTTP and s for
// storage (used as both the documents and attempts store).  Both fields
// are set so that even EMMA's internal http.Client (which reads
// c.transport) is intercepted.  s accepts any implementation of the
// combined interface so test-local wrappers (e.g. alwaysExistsStore,
// funcTransport) can be passed without a concrete-type cast.
func newTestClient(s combinedStore, t http.RoundTripper) *Client {
	c := NewClient("test-agent/1.0", s, s)
	c.httpClient = &http.Client{Transport: t}
	c.transport = t
	return c
}
