package sqlite

import (
	"fmt"
	"os"
	"path/filepath"

	sqlitegorm "github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"investment-analysis/persistence"
)

type Store struct {
	*persistence.GormStore
}

func Open() (*Store, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(filepath.Dir(exePath), "app-template.db")
	db, err := gorm.Open(sqlitegorm.Open(dbPath), &gorm.Config{})
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
			temperature REAL NOT NULL DEFAULT 0.7,
			system_prompt TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS investments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT,
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
		"uuid":                    "TEXT",
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

	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_investments_uuid ON investments(uuid)`).Error; err != nil {
		return err
	}

	var investmentIDs []int64
	if err := db.Raw(`SELECT id FROM investments WHERE uuid IS NULL OR uuid = ''`).Scan(&investmentIDs).Error; err != nil {
		return err
	}
	for _, id := range investmentIDs {
		investmentUUID, err := uuid.NewV6()
		if err != nil {
			return err
		}
		if err := db.Exec(`UPDATE investments SET uuid = ? WHERE id = ?`, investmentUUID.String(), id).Error; err != nil {
			return err
		}
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS investment_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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

	return db.AutoMigrate(
		&persistence.LLMSettingsRow{},
		&persistence.InvestmentRow{},
		&persistence.InvestmentCategoryRow{},
		&persistence.InvestmentExpenseRow{},
		&persistence.InvestmentSaleAssumptionRow{},
	)
}

func ensureColumn(db *gorm.DB, table string, column string, definition string) error {
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

var _ persistence.Store = (*Store)(nil)
