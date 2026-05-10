package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned by table methods when a query expects a row
// but finds none.
var ErrNotFound = errors.New("persistence: not found")

// TimestampLayout is the canonical format used by every persisted
// timestamp column (created_at, updated_at, requested_at, etc.).
const TimestampLayout = "2006-01-02 15:04:05"

// CurrentTimestamp returns the current UTC time formatted in
// TimestampLayout.
func CurrentTimestamp() string {
	return time.Now().UTC().Format(TimestampLayout)
}

// NewSortableID returns a UUIDv7 string.  v7 places a unix-millisecond
// timestamp in the leading 48 bits with a monotonic sub-ms counter,
// giving lexicographically sortable IDs convenient as database primary
// keys.
func NewSortableID() string {
	id, err := uuid.NewV7()
	if err != nil {
		// uuid.NewV7 only fails if the system RNG is unavailable, which
		// would also break uuid.NewRandom; fall back to v4 so callers
		// always get a valid string.
		return uuid.New().String()
	}
	return id.String()
}
