package retrieval

import (
	"context"

	"investment-analysis/persistence/model"
)

// DocumentsStore is the small contract retrieval needs to read and
// write the documents table.  *model.DocumentsTable satisfies it.
type DocumentsStore interface {
	Exists(ctx context.Context, key string) (bool, error)
	Insert(ctx context.Context, doc *model.Document) error
	GetBody(ctx context.Context, key string) ([]byte, error)
}

// AttemptsStore is the small contract retrieval needs to log document
// retrieval outcomes.  *model.RetrievalAttemptsTable satisfies it.
type AttemptsStore interface {
	Log(ctx context.Context, attempt model.RetrievalAttempt) error
}
