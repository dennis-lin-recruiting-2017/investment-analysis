package postgres

import (
	"os"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"investment-analysis/persistence"
)

type Store struct {
	*persistence.GormStore
}

func Open() (*Store, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://localhost/investment_analysis?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return &Store{GormStore: persistence.NewGormStore(db)}, nil
}

func migrate(db *gorm.DB) error {
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
			id BIGSERIAL PRIMARY KEY,
			uuid TEXT,
			name TEXT NOT NULL,
			ticker TEXT NOT NULL DEFAULT '',
			asset_class TEXT NOT NULL,
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
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS uuid TEXT`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS ticker TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS target_allocation TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS initial_investment DOUBLE PRECISION NOT NULL DEFAULT 0`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS initial_investment_date TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text`,
		`ALTER TABLE investments ADD COLUMN IF NOT EXISTS updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP::text`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_investments_uuid ON investments(uuid)`,
	}
	for _, statement := range investmentStatements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}

	type investmentIDRow struct {
		ID int64 `gorm:"column:id"`
	}
	var investmentRows []investmentIDRow
	if err := db.Raw(`SELECT id FROM investments WHERE uuid IS NULL OR uuid = ''`).Scan(&investmentRows).Error; err != nil {
		return err
	}
	for _, row := range investmentRows {
		investmentUUID, err := uuid.NewV6()
		if err != nil {
			return err
		}
		if err := db.Exec(`UPDATE investments SET uuid = $1 WHERE id = $2`, investmentUUID.String(), row.ID).Error; err != nil {
			return err
		}
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS investment_categories (
			id BIGSERIAL PRIMARY KEY,
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
			id BIGSERIAL PRIMARY KEY,
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
			id BIGSERIAL PRIMARY KEY,
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

	return db.AutoMigrate(
		&persistence.LLMSettingsRow{},
		&persistence.InvestmentRow{},
		&persistence.InvestmentCategoryRow{},
		&persistence.InvestmentExpenseRow{},
		&persistence.InvestmentSaleAssumptionRow{},
	)
}

var _ persistence.Store = (*Store)(nil)
