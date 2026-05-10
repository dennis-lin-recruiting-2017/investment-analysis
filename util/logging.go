package util

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
)

// traceIDCtxKey is the unexported key under which a trace ID is stored
// in a context.Context.
type traceIDCtxKey struct{}

// NewTraceContext returns a fresh context.Background() decorated with
// a newly generated trace ID.  Use this anywhere you would otherwise
// reach for context.Background() — every error log produced under the
// returned context will include the trace ID, making it easy to
// correlate related log entries.
func NewTraceContext() context.Context {
	return WithTraceID(context.Background(), uuid.New().String())
}

// WithTraceID returns a copy of ctx with traceID attached.  Use this
// to seed a request-scoped context (e.g. an HTTP handler's
// r.Context()) with a trace ID supplied externally or freshly minted.
// If ctx is nil it falls back to context.Background().
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, traceIDCtxKey{}, traceID)
}

// TraceID returns the trace ID stored in ctx, or "" if none is set.
func TraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(traceIDCtxKey{}).(string); ok {
		return id
	}
	return ""
}

// LogIfErr is the exported alias of logIfErr.  External packages
// (persistence, llm, retrieval, retrieval/securities) call this so
// trace IDs are uniformly included in their error logs.  See logIfErr
// for usage.
func LogIfErr(ctx context.Context, errp *error, fn string, args ...any) {
	logIfErr(ctx, errp, fn, args...)
}

// logIfErr logs the function name, alternating key/value args, and the
// dereferenced error if it is non-nil.  If ctx carries a trace ID
// (set via WithTraceID or NewTraceContext) it is included as a
// [trace=...] prefix to aid correlation across log entries.
//
// Use as the body of a deferred closure with named returns:
//
//	func (t *T) Foo(ctx context.Context, x int) (out X, err error) {
//	    defer func() { logIfErr(ctx, &err, "T.Foo", "x", x) }()
//	    // ... body
//	}
//
// Pass ctx == nil for callers that don't have a context in scope; the
// trace prefix is omitted in that case.
//
// Sensitive or large values (API keys, prompt/response text, document
// bodies) should be omitted from args by the caller.
func logIfErr(ctx context.Context, errp *error, fn string, args ...any) {
	if errp == nil || *errp == nil {
		return
	}
	var b strings.Builder
	for i := 0; i < len(args); i += 2 {
		if i > 0 {
			b.WriteString(", ")
		}
		if i+1 < len(args) {
			fmt.Fprintf(&b, "%v=%v", args[i], args[i+1])
		} else {
			fmt.Fprintf(&b, "%v", args[i])
		}
	}
	if traceID := TraceID(ctx); traceID != "" {
		log.Printf("[trace=%s] %s(%s): %v", traceID, fn, b.String(), *errp)
	} else {
		log.Printf("%s(%s): %v", fn, b.String(), *errp)
	}
}
