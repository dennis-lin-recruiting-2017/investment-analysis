package persistence

import (
	"fmt"

	sqlitegorm "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenFile opens (or creates) a SQLite database at filename and returns a
// *GormStore backed by it.
//
// Only the tables required by RetrievalStore (documents, retrieval_attempts)
// are created.  This makes OpenFile suitable for the retrieval package's
// clients, which only need to fetch and store documents:
//
//	store, err := persistence.OpenFile("my-docs.db")
//	client := retrieval.NewClient("", store)
//	defer store.Close()
//
// For the full application schema (investments, llm_settings, etc.) use
// persistence/sqlite.Open(), which also auto-derives the database path from
// the executable location.
//
// The returned *GormStore satisfies the Store, RetrievalStore, SettingsStore,
// and InvestmentStore interfaces; callers that only need RetrievalStore can
// assign the return value directly:
//
//	var rs RetrievalStore = store
func OpenFile(filename string) (*GormStore, error) {
	db, err := gorm.Open(sqlitegorm.Open(filename), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("persistence.OpenFile %q: %w", filename, err)
	}
	if err := migrateRetrievalSchema(db); err != nil {
		return nil, fmt.Errorf("persistence.OpenFile %q: migrate: %w", filename, err)
	}
	return NewGormStore(db), nil
}

// migrateRetrievalSchema creates the documents and retrieval_attempts tables
// when they do not already exist, then runs AutoMigrate to add any new columns
// introduced by the GORM model definitions.
func migrateRetrievalSchema(db *gorm.DB) error {
	ddl := []string{
		`CREATE TABLE IF NOT EXISTS documents (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			doc_key       TEXT    NOT NULL UNIQUE,
			document_type TEXT    NOT NULL,
			ticker        TEXT    NOT NULL DEFAULT '',
			fiscal_year   INTEGER NOT NULL DEFAULT 0,
			fiscal_quarter INTEGER NOT NULL DEFAULT 0,
			form          TEXT    NOT NULL DEFAULT '',
			source_url    TEXT    NOT NULL DEFAULT '',
			output_label  TEXT    NOT NULL DEFAULT '',
			mime_type     TEXT    NOT NULL DEFAULT '',
			body          BLOB    NOT NULL,
			created_at    TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS retrieval_attempts (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			doc_key        TEXT    NOT NULL,
			document_type  TEXT    NOT NULL,
			ticker         TEXT    NOT NULL DEFAULT '',
			fiscal_year    INTEGER NOT NULL DEFAULT 0,
			fiscal_quarter INTEGER NOT NULL DEFAULT 0,
			status         TEXT    NOT NULL,
			message        TEXT    NOT NULL DEFAULT '',
			created_at     TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_retrieval_attempts_doc_key
			ON retrieval_attempts(doc_key)`,
		`CREATE TABLE IF NOT EXISTS retrieval_settings (
			id                        INTEGER PRIMARY KEY DEFAULT 1,
			playwright_timeout_seconds INTEGER NOT NULL DEFAULT 300,
			updated_at                TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, stmt := range ddl {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	// Schema migration: rename the 'mode' column to 'document_type' in
	// databases that were created before this rename.  Each statement is
	// best-effort and idempotent: on a fresh database the ADD COLUMN will fail
	// (column already exists), and on a fully-migrated database the DROP COLUMN
	// will fail (column already gone).  We run these with a silent logger so
	// the expected failures don't pollute application output.
	silent := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	renameMigrations := []string{
		`ALTER TABLE documents          ADD COLUMN document_type TEXT NOT NULL DEFAULT ''`,
		`UPDATE documents               SET document_type = mode WHERE document_type = ''`,
		`ALTER TABLE documents          DROP COLUMN mode`,
		`ALTER TABLE retrieval_attempts ADD COLUMN document_type TEXT NOT NULL DEFAULT ''`,
		`UPDATE retrieval_attempts      SET document_type = mode WHERE document_type = ''`,
		`ALTER TABLE retrieval_attempts DROP COLUMN mode`,
	}
	for _, stmt := range renameMigrations {
		silent.Exec(stmt) //nolint:errcheck // idempotent; failure means column already in target state
	}

	return db.AutoMigrate(&DocumentRow{}, &RetrievalAttemptRow{}, &RetrievalSettingsRow{})
}
