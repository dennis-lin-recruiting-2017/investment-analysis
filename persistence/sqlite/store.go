// Package sqlite opens (or creates) a SQLite database, applies the
// application schema, and returns a *persistence.Store.
package sqlite

import (
	"fmt"
	"investment-analysis/util"

	sqlitegorm "github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"investment-analysis/persistence"
	"investment-analysis/persistence/model"
)

// NewStore opens a SQLite database at path (creating it if necessary),
// runs the full application schema migration, and returns a
// *persistence.Store ready for use.
func NewStore(path string) (store *persistence.Store, err error) {
	defer func() { util.LogIfErr(nil, &err, "sqlite.NewStore", "path", path) }()
	db, e := gorm.Open(sqlitegorm.Open(path), &gorm.Config{})
	if e != nil {
		return nil, fmt.Errorf("sqlite.NewStore %q: %w", path, e)
	}
	if e := migrate(db); e != nil {
		return nil, fmt.Errorf("sqlite.NewStore %q: migrate: %w", path, e)
	}
	return persistence.NewStore(db), nil
}

func migrate(db *gorm.DB) (err error) {
	defer func() { util.LogIfErr(nil, &err, "sqlite.migrate") }()
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS llm_settings (
			provider TEXT PRIMARY KEY,
			endpoint TEXT NOT NULL,
			model TEXT NOT NULL,
			api_key TEXT NOT NULL DEFAULT '',
			temperature REAL NOT NULL DEFAULT 0.7,
			system_prompt TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS investments (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			ticker TEXT NOT NULL DEFAULT '',
			asset_class TEXT NOT NULL,
			purchase_price REAL NOT NULL DEFAULT 0,
			coupon_rate REAL NOT NULL DEFAULT 0,
			maturity_date TEXT NOT NULL DEFAULT '',
			callable_date_start TEXT NOT NULL DEFAULT '',
			call_price REAL NOT NULL DEFAULT 0,
			call_date TEXT NOT NULL DEFAULT '',
			thesis TEXT NOT NULL,
			target_allocation TEXT NOT NULL DEFAULT '',
			initial_investment REAL NOT NULL DEFAULT 0,
			initial_investment_date TEXT NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	investmentColumns := map[string]string{
		"ticker":                  "TEXT NOT NULL DEFAULT ''",
		"purchase_price":          "REAL NOT NULL DEFAULT 0",
		"coupon_rate":             "REAL NOT NULL DEFAULT 0",
		"maturity_date":           "TEXT NOT NULL DEFAULT ''",
		"callable_date_start":     "TEXT NOT NULL DEFAULT ''",
		"call_price":              "REAL NOT NULL DEFAULT 0",
		"call_date":               "TEXT NOT NULL DEFAULT ''",
		"target_allocation":       "TEXT NOT NULL DEFAULT ''",
		"initial_investment":      "REAL NOT NULL DEFAULT 0",
		"initial_investment_date": "TEXT NOT NULL DEFAULT ''",
		"notes":                   "TEXT NOT NULL DEFAULT ''",
		"created_at":              "TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP",
		"updated_at":              "TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP",
	}
	for column, definition := range investmentColumns {
		if err := ensureColumn(db, "investments", column, definition); err != nil {
			return err
		}
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS investment_categories (
			id TEXT PRIMARY KEY,
			investment_uuid TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_investment_categories_name ON investment_categories(investment_uuid, name)`).Error; err != nil {
		return err
	}

	categoryColumns := map[string]string{
		"created_at": "TEXT NOT NULL DEFAULT ''",
		"updated_at": "TEXT NOT NULL DEFAULT ''",
	}
	for column, definition := range categoryColumns {
		if err := ensureColumn(db, "investment_categories", column, definition); err != nil {
			return err
		}
	}

	if err := db.Exec(`
		UPDATE investment_categories
		SET created_at = CURRENT_TIMESTAMP
		WHERE created_at IS NULL OR created_at = ''
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		UPDATE investment_categories
		SET updated_at = CURRENT_TIMESTAMP
		WHERE updated_at IS NULL OR updated_at = ''
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS investment_expenses (
			id TEXT PRIMARY KEY,
			investment_uuid TEXT NOT NULL,
			event_type TEXT NOT NULL DEFAULT 'cash-flow',
			flow_type TEXT NOT NULL DEFAULT 'one-time',
			recurrence_interval TEXT NOT NULL DEFAULT '',
			label TEXT NOT NULL,
			amount REAL NOT NULL,
			due_date TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			start_date TEXT NOT NULL DEFAULT '',
			end_date TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT '',
			adjustment_frequency TEXT NOT NULL DEFAULT '',
			adjustment_mode TEXT NOT NULL DEFAULT '',
			adjustment_value REAL NOT NULL DEFAULT 0,
			notes TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS investment_sale_assumptions (
			id TEXT PRIMARY KEY,
			investment_uuid TEXT NOT NULL,
			label TEXT NOT NULL,
			amount REAL NOT NULL,
			growth_type TEXT NOT NULL DEFAULT 'fixed',
			growth_period TEXT NOT NULL DEFAULT '',
			growth_mode TEXT NOT NULL DEFAULT '',
			growth_value REAL NOT NULL DEFAULT 0,
			category TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			start_date TEXT NOT NULL DEFAULT '',
			end_date TEXT NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	saleAssumptionColumns := map[string]string{
		"growth_type":   "TEXT NOT NULL DEFAULT 'fixed'",
		"growth_period": "TEXT NOT NULL DEFAULT ''",
		"growth_mode":   "TEXT NOT NULL DEFAULT ''",
		"growth_value":  "REAL NOT NULL DEFAULT 0",
		"category":      "TEXT NOT NULL DEFAULT ''",
		"description":   "TEXT NOT NULL DEFAULT ''",
		"start_date":    "TEXT NOT NULL DEFAULT ''",
		"end_date":      "TEXT NOT NULL DEFAULT ''",
		"notes":         "TEXT NOT NULL DEFAULT ''",
		"created_at":    "TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP",
		"updated_at":    "TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP",
	}
	for column, definition := range saleAssumptionColumns {
		if err := ensureColumn(db, "investment_sale_assumptions", column, definition); err != nil {
			return err
		}
	}

	expenseColumns := map[string]string{
		"description":          "TEXT NOT NULL DEFAULT ''",
		"event_type":           "TEXT NOT NULL DEFAULT 'cash-flow'",
		"flow_type":            "TEXT NOT NULL DEFAULT 'one-time'",
		"recurrence_interval":  "TEXT NOT NULL DEFAULT ''",
		"start_date":           "TEXT NOT NULL DEFAULT ''",
		"end_date":             "TEXT NOT NULL DEFAULT ''",
		"category":             "TEXT NOT NULL DEFAULT ''",
		"adjustment_frequency": "TEXT NOT NULL DEFAULT ''",
		"adjustment_mode":      "TEXT NOT NULL DEFAULT ''",
		"adjustment_value":     "REAL NOT NULL DEFAULT 0",
		"notes":                "TEXT NOT NULL DEFAULT ''",
		"created_at":           "TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP",
		"updated_at":           "TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP",
	}
	for column, definition := range expenseColumns {
		if err := ensureColumn(db, "investment_expenses", column, definition); err != nil {
			return err
		}
	}

	if err := db.Exec(`
		UPDATE investment_expenses
		SET end_date = due_date
		WHERE (end_date IS NULL OR end_date = '')
		  AND due_date IS NOT NULL
		  AND due_date != ''
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			investment_uuid TEXT NOT NULL DEFAULT '',
			doc_key TEXT NOT NULL UNIQUE,
			document_type TEXT NOT NULL,
			ticker TEXT NOT NULL,
			fiscal_year INTEGER NOT NULL,
			fiscal_quarter INTEGER NOT NULL,
			form TEXT NOT NULL DEFAULT '',
			source_url TEXT NOT NULL DEFAULT '',
			output_label TEXT NOT NULL DEFAULT '',
			mime_type TEXT NOT NULL DEFAULT '',
			body BLOB NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	if err := ensureColumn(db, "documents", "investment_uuid", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}

	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_documents_investment_uuid ON documents(investment_uuid)`).Error; err != nil {
		return err
	}

	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_documents_investment_key ON documents(investment_uuid, doc_key)`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS retrieval_attempts (
			id TEXT PRIMARY KEY,
			investment_uuid TEXT NOT NULL DEFAULT '',
			doc_key TEXT NOT NULL,
			document_type TEXT NOT NULL,
			ticker TEXT NOT NULL,
			fiscal_year INTEGER NOT NULL,
			fiscal_quarter INTEGER NOT NULL,
			status TEXT NOT NULL,
			message TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	if err := ensureColumn(db, "retrieval_attempts", "investment_uuid", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}

	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_retrieval_attempts_doc_key ON retrieval_attempts(doc_key)`).Error; err != nil {
		return err
	}

	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_retrieval_attempts_investment_uuid ON retrieval_attempts(investment_uuid)`).Error; err != nil {
		return err
	}

	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_retrieval_attempts_investment_key ON retrieval_attempts(investment_uuid, doc_key)`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS retrieval_settings (
			id                        INTEGER PRIMARY KEY DEFAULT 1,
			playwright_timeout_seconds INTEGER NOT NULL DEFAULT 300,
			updated_at                TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS inferences (
			uuid                TEXT    NOT NULL PRIMARY KEY,
			investment_uuid     TEXT    NOT NULL,
			inference_key       TEXT    NOT NULL DEFAULT '',
			model               TEXT    NOT NULL DEFAULT '',
			prompt              TEXT    NOT NULL DEFAULT '',
			response            TEXT    NOT NULL DEFAULT '',
			processed_artifact  TEXT    NOT NULL DEFAULT '',
			prompt_tokens       INTEGER NOT NULL DEFAULT 0,
			completion_tokens   INTEGER NOT NULL DEFAULT 0,
			total_tokens        INTEGER NOT NULL DEFAULT 0,
			requested_at        TEXT    NOT NULL DEFAULT '',
			responded_at        TEXT    NOT NULL DEFAULT '',
			elapsed_ms          INTEGER NOT NULL DEFAULT 0,
			created_at          TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	if err := ensureColumn(db, "inferences", "inference_key", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}

	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_inferences_investment_uuid ON inferences(investment_uuid)`).Error; err != nil {
		return err
	}

	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_inferences_investment_key ON inferences(investment_uuid, inference_key)`).Error; err != nil {
		return err
	}

	return db.AutoMigrate(
		&model.LLMSettings{},
		&model.Investment{},
		&model.InvestmentCategory{},
		&model.InvestmentExpense{},
		&model.InvestmentSaleAssumption{},
		&model.Document{},
		&model.RetrievalAttempt{},
		&model.RetrievalSettings{},
		&model.Inference{},
	)
}

func ensureColumn(db *gorm.DB, table string, column string, definition string) (err error) {
	defer func() {
		util.LogIfErr(nil, &err, "sqlite.ensureColumn", "table", table, "column", column)
	}()
	type pragmaColumn struct {
		Name string `gorm:"column:name"`
	}

	var columns []pragmaColumn
	if err := db.Raw(fmt.Sprintf(`PRAGMA table_info(%s)`, table)).Scan(&columns).Error; err != nil {
		return err
	}

	for _, existing := range columns {
		if existing.Name == column {
			return nil
		}
	}

	return db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, column, definition)).Error
}
