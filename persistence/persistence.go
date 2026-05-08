package persistence

import (
	"context"
	"errors"
	model2 "investment-analysis/persistence/model"
)

var ErrNotFound = errors.New("persistence: not found")

type SettingsStore interface {
	GetSettings(provider string) (model2.LLMSettings, error)
	ListSettings() ([]model2.LLMSettings, error)
	UpsertSettings(settings model2.LLMSettings) error
	DeleteSettings(provider string) error
}

type InvestmentStore interface {
	ListInvestments() ([]model2.Investment, error)
	GetInvestment(investmentUUID string) (model2.Investment, error)
	CreateInvestment(investment model2.Investment) (model2.Investment, error)
	UpdateInvestment(investmentUUID string, investment model2.Investment) (model2.Investment, error)
	DeleteInvestment(investmentUUID string) error
	ListInvestmentCategories(investmentUUID string) ([]string, error)
	CreateInvestmentCategory(investmentUUID string, name string) error
	UpdateInvestmentCategory(investmentUUID string, currentName string, newName string) error
	DeleteInvestmentCategory(investmentUUID string, name string) error
	ListInvestmentExpenses(investmentUUID string) ([]model2.InvestmentExpense, error)
	CreateInvestmentExpense(investmentUUID string, expense model2.InvestmentExpense) (model2.InvestmentExpense, error)
	UpdateInvestmentExpense(investmentUUID string, expenseID int64, expense model2.InvestmentExpense) (model2.InvestmentExpense, error)
	DeleteInvestmentExpense(investmentUUID string, expenseID int64) error
	ListInvestmentSaleAssumptions(investmentUUID string) ([]model2.InvestmentSaleAssumption, error)
	CreateInvestmentSaleAssumption(investmentUUID string, assumption model2.InvestmentSaleAssumption) (model2.InvestmentSaleAssumption, error)
	UpdateInvestmentSaleAssumption(investmentUUID string, assumptionID int64, assumption model2.InvestmentSaleAssumption) (model2.InvestmentSaleAssumption, error)
	DeleteInvestmentSaleAssumption(investmentUUID string, assumptionID int64) error
}

type RetrievalStore interface {
	DocumentExists(ctx context.Context, key string) (bool, error)
	InsertDocument(ctx context.Context, doc model2.StoredDocument) error
	LogRetrievalAttempt(ctx context.Context, attempt model2.RetrievalAttempt) error
	GetDocumentBody(ctx context.Context, key string) ([]byte, error)
}

type RetrievalSettingsStore interface {
	GetRetrievalSettings() (model2.RetrievalSettings, error)
	UpsertRetrievalSettings(settings model2.RetrievalSettings) error
}

type Store interface {
	SettingsStore
	InvestmentStore
	RetrievalStore
	RetrievalSettingsStore
	Close() error
}
