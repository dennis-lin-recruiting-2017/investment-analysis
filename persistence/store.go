// Package persistence provides the storage client for the application.
//
// The model structs that represent database rows (and that serialize
// directly as JSON) live in persistence/model along with the *Table
// CRUD wrappers that operate on them.  The Store struct in this
// package owns one *Table per concept and is the entry point used by
// the rest of the codebase:
//
//	store, _ := sqlite.NewStore("/path/to.db")
//	defer store.Close()
//	inv, _ := store.Investments.Get(uuid)
//	store.Inferences.Save(ctx, inference)
//
// Construct a Store via the sqlite or postgres sub-package's NewStore
// — those handle connection setup and schema migration.
package persistence

import (
	"investment-analysis/util"

	"gorm.io/gorm"

	"investment-analysis/persistence/model"
)

// Store is the persistence client.  Its exported fields are pointers
// to the *Table CRUD wrappers in persistence/model, one per database
// table.
type Store struct {
	db *gorm.DB

	LLMSettings               *model.LLMSettingsTable
	RetrievalSettings         *model.RetrievalSettingsTable
	Investments               *model.InvestmentsTable
	InvestmentCategories      *model.InvestmentCategoriesTable
	InvestmentExpenses        *model.InvestmentExpensesTable
	InvestmentSaleAssumptions *model.InvestmentSaleAssumptionsTable
	Documents                 *model.DocumentsTable
	RetrievalAttempts         *model.RetrievalAttemptsTable
	Inferences                *model.InferencesTable
}

// NewStore wraps an already-configured *gorm.DB (with migrations
// applied) into a Store with all table structs wired up.
func NewStore(db *gorm.DB) *Store {
	return &Store{
		db: db,

		LLMSettings:               model.NewLLMSettingsTable(db),
		RetrievalSettings:         model.NewRetrievalSettingsTable(db),
		Investments:               model.NewInvestmentsTable(db),
		InvestmentCategories:      model.NewInvestmentCategoriesTable(db),
		InvestmentExpenses:        model.NewInvestmentExpensesTable(db),
		InvestmentSaleAssumptions: model.NewInvestmentSaleAssumptionsTable(db),
		Documents:                 model.NewDocumentsTable(db),
		RetrievalAttempts:         model.NewRetrievalAttemptsTable(db),
		Inferences:                model.NewInferencesTable(db),
	}
}

// Close closes the underlying database connection.
func (s *Store) Close() (err error) {
	defer func() { util.LogIfErr(nil, &err, "persistence.Store.Close") }()
	sqlDB, e := s.db.DB()
	if e != nil {
		return e
	}
	return sqlDB.Close()
}
