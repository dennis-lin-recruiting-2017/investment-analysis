package persistence

import (
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

type Store interface {
	SettingsStore
	InvestmentStore
	Close() error
}
