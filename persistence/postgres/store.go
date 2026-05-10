// Package postgres opens a Postgres connection, applies the application
// schema, and returns a *persistence.Store.
package postgres

import (
	"fmt"
	"investment-analysis/util"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"investment-analysis/persistence"
	"investment-analysis/persistence/model"
)

// NewStore opens a Postgres connection at dsn, runs the full
// application schema migration, and returns a *persistence.Store.
// If dsn is empty a sensible localhost default is used.
func NewStore(dsn string) (store *persistence.Store, err error) {
	defer func() {
		// dsn may contain credentials — log only whether one was supplied.
		util.LogIfErr(nil, &err, "postgres.NewStore", "dsnSet", dsn != "")
	}()
	if dsn == "" {
		dsn = "postgres://localhost/investment_analysis?sslmode=disable"
	}
	db, e := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if e != nil {
		return nil, fmt.Errorf("postgres.NewStore: %w", e)
	}
	if e := migrate(db); e != nil {
		return nil, fmt.Errorf("postgres.NewStore: migrate: %w", e)
	}
	return persistence.NewStore(db), nil
}

func migrate(db *gorm.DB) (err error) {
	defer func() { util.LogIfErr(nil, &err, "postgres.migrate") }()
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS llm_settings (
			provider TEXT PRIMARY KEY,
			endpoint TEXT NOT NULL,
			model TEXT NOT NULL,
			api_key TEXT NOT NULL DEFAULT '',
			temperature DOUBLE PRECISION NOT NULL DEFAULT 0.7,
			system_prompt TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text
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
			purchase_price DOUBLE PRECISION NOT NULL DEFAULT 0,
			coupon_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
			maturity_date TEXT NOT NULL DEFAULT '',
			callable_date_start TEXT NOT NULL DEFAULT '',
			call_price DOUBLE PRECISION NOT NULL DEFAULT 0,
			call_date TEXT NOT NULL DEFAULT '',
			thesis TEXT NOT NULL,
			target_allocation TEXT NOT NULL DEFAULT '',
			initial_investment DOUBLE PRECISION NOT NULL DEFAULT 0,
			initial_investment_date TEXT NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text
		)
	`).Error; err != nil {
		return err
	}

	investmentStatements := []string{
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS ticker TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS purchase_price DOUBLE PRECISION NOT NULL DEFAULT 0`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS coupon_rate DOUBLE PRECISION NOT NULL DEFAULT 0`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS maturity_date TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS callable_date_start TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS call_price DOUBLE PRECISION NOT NULL DEFAULT 0`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS call_date TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS target_allocation TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS initial_investment DOUBLE PRECISION NOT NULL DEFAULT 0`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS initial_investment_date TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text`,
	}
	for _, statement := range investmentStatements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS investment_categories (
			id TEXT PRIMARY KEY,
			investment_uuid TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text
		)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_investment_categories_name ON investment_categories(investment_uuid, name)`).Error; err != nil {
		return err
	}

	categoryStatements := []string{
		`ALTER TABLE investment_categories ADD COLUMN IF NOT EXISTS created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text`,
		`ALTER TABLE investment_categories ADD COLUMN IF NOT EXISTS updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text`,
	}
	for _, statement := range categoryStatements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS investment_expenses (
			id TEXT PRIMARY KEY,
			investment_uuid TEXT NOT NULL,
			event_type TEXT NOT NULL DEFAULT 'cash-flow',
			flow_type TEXT NOT NULL DEFAULT 'one-time',
			recurrence_interval TEXT NOT NULL DEFAULT '',
			label TEXT NOT NULL,
			amount DOUBLE PRECISION NOT NULL,
			due_date TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			start_date TEXT NOT NULL DEFAULT '',
			end_date TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT '',
			adjustment_frequency TEXT NOT NULL DEFAULT '',
			adjustment_mode TEXT NOT NULL DEFAULT '',
			adjustment_value DOUBLE PRECISION NOT NULL DEFAULT 0,
			notes TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text
		)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS investment_sale_assumptions (
			id TEXT PRIMARY KEY,
			investment_uuid TEXT NOT NULL,
			label TEXT NOT NULL,
			amount DOUBLE PRECISION NOT NULL,
			growth_type TEXT NOT NULL DEFAULT 'fixed',
			growth_period TEXT NOT NULL DEFAULT '',
			growth_mode TEXT NOT NULL DEFAULT '',
			growth_value DOUBLE PRECISION NOT NULL DEFAULT 0,
			category TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			start_date TEXT NOT NULL DEFAULT '',
			end_date TEXT NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text
		)
	`).Error; err != nil {
		return err
	}

	saleAssumptionStatements := []string{
		`ALTER TABLE investment_sale_assumptions ADD COLUMN IF NOT EXISTS growth_type TEXT NOT NULL DEFAULT 'fixed'`,
		`ALTER TABLE investment_sale_assumptions ADD COLUMN IF NOT EXISTS growth_period TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_sale_assumptions ADD COLUMN IF NOT EXISTS growth_mode TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_sale_assumptions ADD COLUMN IF NOT EXISTS growth_value DOUBLE PRECISION NOT NULL DEFAULT 0`,
		`ALTER TABLE investment_sale_assumptions ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_sale_assumptions ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_sale_assumptions ADD COLUMN IF NOT EXISTS start_date TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_sale_assumptions ADD COLUMN IF NOT EXISTS end_date TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_sale_assumptions ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_sale_assumptions ADD COLUMN IF NOT EXISTS created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text`,
		`ALTER TABLE investment_sale_assumptions ADD COLUMN IF NOT EXISTS updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text`,
	}
	for _, statement := range saleAssumptionStatements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}

	expenseStatements := []string{
		`ALTER TABLE investment_expenses ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_expenses ADD COLUMN IF NOT EXISTS event_type TEXT NOT NULL DEFAULT 'cash-flow'`,
		`ALTER TABLE investment_expenses ADD COLUMN IF NOT EXISTS flow_type TEXT NOT NULL DEFAULT 'one-time'`,
		`ALTER TABLE investment_expenses ADD COLUMN IF NOT EXISTS recurrence_interval TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_expenses ADD COLUMN IF NOT EXISTS start_date TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_expenses ADD COLUMN IF NOT EXISTS end_date TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_expenses ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_expenses ADD COLUMN IF NOT EXISTS adjustment_frequency TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_expenses ADD COLUMN IF NOT EXISTS adjustment_mode TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_expenses ADD COLUMN IF NOT EXISTS adjustment_value DOUBLE PRECISION NOT NULL DEFAULT 0`,
		`ALTER TABLE investment_expenses ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investment_expenses ADD COLUMN IF NOT EXISTS created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text`,
		`ALTER TABLE investment_expenses ADD COLUMN IF NOT EXISTS updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text`,
	}
	for _, statement := range expenseStatements {
		if err := db.Exec(statement).Error; err != nil {
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
			body BYTEA NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text
		)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`ALTER TABLE documents ADD COLUMN IF NOT EXISTS investment_uuid TEXT NOT NULL DEFAULT ''`).Error; err != nil {
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
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text
		)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`ALTER TABLE retrieval_attempts ADD COLUMN IF NOT EXISTS investment_uuid TEXT NOT NULL DEFAULT ''`).Error; err != nil {
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
			updated_at                TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP::text
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
			elapsed_ms          BIGINT  NOT NULL DEFAULT 0,
			created_at          TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP::text
		)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`ALTER TABLE inferences ADD COLUMN IF NOT EXISTS inference_key TEXT NOT NULL DEFAULT ''`).Error; err != nil {
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
