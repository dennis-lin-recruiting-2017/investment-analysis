package persistence

import (
	"context"
	"errors"
	model2 "investment-analysis/persistence/model"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const expenseOrderClause = "COALESCE(NULLIF(end_date, ''), due_date) ASC, id ASC"

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (s *GormStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *GormStore) GetSettings(provider string) (model2.LLMSettings, error) {
	var row LLMSettingsRow
	if err := s.db.Where("provider = ?", provider).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model2.DefaultSettings(provider), nil
		}
		return model2.LLMSettings{}, err
	}
	return row.ToModel(), nil
}

func (s *GormStore) ListSettings() ([]model2.LLMSettings, error) {
	providers := model2.Providers()
	out := make([]model2.LLMSettings, 0, len(providers))
	for _, provider := range providers {
		settings, err := s.GetSettings(provider)
		if err != nil {
			return nil, err
		}
		out = append(out, settings)
	}
	return out, nil
}

func (s *GormStore) UpsertSettings(settings model2.LLMSettings) error {
	row := NewLLMSettingsRow(settings)
	now := currentTimestamp()
	row.UpdatedAt = now

	return s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "provider"}},
		DoUpdates: clause.Assignments(map[string]any{
			"endpoint":      row.Endpoint,
			"model":         row.Model,
			"api_key":       row.APIKey,
			"temperature":   row.Temperature,
			"system_prompt": row.SystemPrompt,
			"updated_at":    now,
		}),
	}).Create(&row).Error
}

func (s *GormStore) DeleteSettings(provider string) error {
	return s.db.Delete(&LLMSettingsRow{}, "provider = ?", provider).Error
}

func (s *GormStore) GetRetrievalSettings() (model2.RetrievalSettings, error) {
	var row RetrievalSettingsRow
	if err := s.db.First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model2.DefaultRetrievalSettings(), nil
		}
		return model2.RetrievalSettings{}, err
	}
	return row.ToModel(), nil
}

func (s *GormStore) UpsertRetrievalSettings(settings model2.RetrievalSettings) error {
	row := RetrievalSettingsRow{
		ID:                       1,
		PlaywrightTimeoutSeconds: settings.PlaywrightTimeoutSeconds,
		UpdatedAt:                currentTimestamp(),
	}
	return s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"playwright_timeout_seconds": row.PlaywrightTimeoutSeconds,
			"updated_at":                row.UpdatedAt,
		}),
	}).Create(&row).Error
}

func (s *GormStore) ListInvestments() ([]model2.Investment, error) {
	var rows []InvestmentRow
	if err := s.db.Order("updated_at DESC").Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}

	investments := make([]model2.Investment, 0, len(rows))
	for _, row := range rows {
		investments = append(investments, row.ToModel())
	}
	return investments, nil
}

func (s *GormStore) GetInvestment(investmentUUID string) (model2.Investment, error) {
	var row InvestmentRow
	err := s.db.Preload("Categories", func(db *gorm.DB) *gorm.DB {
		return db.Order("name ASC")
	}).Preload("Expenses", func(db *gorm.DB) *gorm.DB {
		return db.Order(expenseOrderClause)
	}).Preload("SaleAssumptions", func(db *gorm.DB) *gorm.DB {
		return db.Order("COALESCE(NULLIF(end_date, ''), start_date) ASC, id ASC")
	}).Where("uuid = ?", investmentUUID).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model2.Investment{}, ErrNotFound
		}
		return model2.Investment{}, err
	}
	return row.ToModel(), nil
}

func normalizeCategoryName(name string) string {
	trimmed := strings.TrimSpace(name)
	if strings.EqualFold(trimmed, "uncategorized") {
		return ""
	}
	return trimmed
}

func (s *GormStore) ListInvestmentCategories(investmentUUID string) ([]string, error) {
	var rows []InvestmentCategoryRow
	if err := s.db.Where("investment_uuid = ?", investmentUUID).Order("name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	categories := make([]string, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, row.Name)
	}
	return categories, nil
}

func (s *GormStore) CreateInvestmentCategory(investmentUUID string, name string) error {
	name = normalizeCategoryName(name)
	if name == "" {
		return nil
	}

	row := InvestmentCategoryRow{
		InvestmentUUID: investmentUUID,
		Name:           name,
	}
	return s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

func (s *GormStore) UpdateInvestmentCategory(investmentUUID string, currentName string, newName string) error {
	currentName = normalizeCategoryName(currentName)
	newName = normalizeCategoryName(newName)
	if currentName == "" || newName == "" {
		return ErrNotFound
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&InvestmentCategoryRow{}).
			Where("investment_uuid = ? AND name = ?", investmentUUID, currentName).
			Update("name", newName)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}

		return tx.Model(&InvestmentExpenseRow{}).
			Where("investment_uuid = ? AND category = ?", investmentUUID, currentName).
			Update("category", newName).Error
	})
}

func (s *GormStore) DeleteInvestmentCategory(investmentUUID string, name string) error {
	name = normalizeCategoryName(name)
	if name == "" {
		return ErrNotFound
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&InvestmentExpenseRow{}).
			Where("investment_uuid = ? AND category = ?", investmentUUID, name).
			Update("category", "").Error; err != nil {
			return err
		}

		result := tx.Delete(&InvestmentCategoryRow{}, "investment_uuid = ? AND name = ?", investmentUUID, name)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (s *GormStore) CreateInvestment(investment model2.Investment) (model2.Investment, error) {
	if investment.UUID == "" {
		investmentUUID, err := uuid.NewV6()
		if err != nil {
			return model2.Investment{}, err
		}
		investment.UUID = investmentUUID.String()
	}

	row := NewInvestmentRow(investment)
	if err := s.db.Create(&row).Error; err != nil {
		return model2.Investment{}, err
	}

	return s.GetInvestment(row.UUID)
}

func (s *GormStore) UpdateInvestment(investmentUUID string, investment model2.Investment) (model2.Investment, error) {
	var existing InvestmentRow
	if err := s.db.Where("uuid = ?", investmentUUID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model2.Investment{}, ErrNotFound
		}
		return model2.Investment{}, err
	}

	row := NewInvestmentRow(investment)
	row.UUID = investmentUUID
	row.CreatedAt = existing.CreatedAt
	row.UpdatedAt = existing.UpdatedAt

	if err := s.db.Model(&existing).Updates(map[string]any{
		"name":                    row.Name,
		"ticker":                  row.Ticker,
		"asset_class":             row.AssetClass,
		"purchase_price":          row.PurchasePrice,
		"coupon_rate":             row.CouponRate,
		"maturity_date":           row.MaturityDate,
		"callable_date_start":     row.CallableDateStart,
		"call_price":              row.CallPrice,
		"call_date":               row.CallDate,
		"thesis":                  row.Thesis,
		"target_allocation":       row.TargetAllocation,
		"initial_investment":      row.InitialInvestment,
		"initial_investment_date": row.InitialInvestmentDate,
		"notes":                   row.Notes,
	}).Error; err != nil {
		return model2.Investment{}, err
	}

	return s.GetInvestment(investmentUUID)
}

func (s *GormStore) DeleteInvestment(investmentUUID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&InvestmentExpenseRow{}, "investment_uuid = ?", investmentUUID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&InvestmentSaleAssumptionRow{}, "investment_uuid = ?", investmentUUID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&InvestmentCategoryRow{}, "investment_uuid = ?", investmentUUID).Error; err != nil {
			return err
		}

		result := tx.Delete(&InvestmentRow{}, "uuid = ?", investmentUUID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (s *GormStore) ListInvestmentExpenses(investmentUUID string) ([]model2.InvestmentExpense, error) {
	var rows []InvestmentExpenseRow
	if err := s.db.Where("investment_uuid = ?", investmentUUID).Order(expenseOrderClause).Find(&rows).Error; err != nil {
		return nil, err
	}

	expenses := make([]model2.InvestmentExpense, 0, len(rows))
	for _, row := range rows {
		expenses = append(expenses, row.ToModel())
	}
	return expenses, nil
}

func (s *GormStore) CreateInvestmentExpense(investmentUUID string, expense model2.InvestmentExpense) (model2.InvestmentExpense, error) {
	expense.InvestmentUUID = investmentUUID
	row := NewInvestmentExpenseRow(expense)
	if err := s.db.Create(&row).Error; err != nil {
		return model2.InvestmentExpense{}, err
	}
	return row.ToModel(), nil
}

func (s *GormStore) UpdateInvestmentExpense(investmentUUID string, expenseID int64, expense model2.InvestmentExpense) (model2.InvestmentExpense, error) {
	var existing InvestmentExpenseRow
	if err := s.db.Where("investment_uuid = ? AND id = ?", investmentUUID, expenseID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model2.InvestmentExpense{}, ErrNotFound
		}
		return model2.InvestmentExpense{}, err
	}

	row := NewInvestmentExpenseRow(expense)
	row.ID = expenseID
	row.InvestmentUUID = investmentUUID
	row.CreatedAt = existing.CreatedAt
	row.UpdatedAt = existing.UpdatedAt

	if err := s.db.Model(&existing).Updates(map[string]any{
		"event_type":           row.EventType,
		"flow_type":            row.FlowType,
		"recurrence_interval":  row.RecurrenceInterval,
		"label":                row.Label,
		"amount":               row.Amount,
		"description":          row.Description,
		"start_date":           row.StartDate,
		"end_date":             row.EndDate,
		"due_date":             row.DueDate,
		"category":             row.Category,
		"adjustment_frequency": row.AdjustmentFrequency,
		"adjustment_mode":      row.AdjustmentMode,
		"adjustment_value":     row.AdjustmentValue,
		"notes":                row.Notes,
	}).Error; err != nil {
		return model2.InvestmentExpense{}, err
	}

	if err := s.db.Where("investment_uuid = ? AND id = ?", investmentUUID, expenseID).First(&existing).Error; err != nil {
		return model2.InvestmentExpense{}, err
	}

	return existing.ToModel(), nil
}

func (s *GormStore) DeleteInvestmentExpense(investmentUUID string, expenseID int64) error {
	result := s.db.Delete(&InvestmentExpenseRow{}, "investment_uuid = ? AND id = ?", investmentUUID, expenseID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *GormStore) ListInvestmentSaleAssumptions(investmentUUID string) ([]model2.InvestmentSaleAssumption, error) {
	var rows []InvestmentSaleAssumptionRow
	if err := s.db.Where("investment_uuid = ?", investmentUUID).Order("COALESCE(NULLIF(end_date, ''), start_date) ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model2.InvestmentSaleAssumption, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToModel())
	}
	return out, nil
}

func (s *GormStore) CreateInvestmentSaleAssumption(investmentUUID string, assumption model2.InvestmentSaleAssumption) (model2.InvestmentSaleAssumption, error) {
	assumption.InvestmentUUID = investmentUUID
	row := NewInvestmentSaleAssumptionRow(assumption)
	if err := s.db.Create(&row).Error; err != nil {
		return model2.InvestmentSaleAssumption{}, err
	}
	return row.ToModel(), nil
}

func (s *GormStore) UpdateInvestmentSaleAssumption(investmentUUID string, assumptionID int64, assumption model2.InvestmentSaleAssumption) (model2.InvestmentSaleAssumption, error) {
	var existing InvestmentSaleAssumptionRow
	if err := s.db.Where("investment_uuid = ? AND id = ?", investmentUUID, assumptionID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model2.InvestmentSaleAssumption{}, ErrNotFound
		}
		return model2.InvestmentSaleAssumption{}, err
	}

	row := NewInvestmentSaleAssumptionRow(assumption)
	row.ID = assumptionID
	row.InvestmentUUID = investmentUUID
	row.CreatedAt = existing.CreatedAt
	row.UpdatedAt = existing.UpdatedAt

	if err := s.db.Model(&existing).Updates(map[string]any{
		"label":         row.Label,
		"amount":        row.Amount,
		"growth_type":   row.GrowthType,
		"growth_period": row.GrowthPeriod,
		"growth_mode":   row.GrowthMode,
		"growth_value":  row.GrowthValue,
		"category":      row.Category,
		"description":   row.Description,
		"start_date":    row.StartDate,
		"end_date":      row.EndDate,
		"notes":         row.Notes,
	}).Error; err != nil {
		return model2.InvestmentSaleAssumption{}, err
	}
	if err := s.db.Where("investment_uuid = ? AND id = ?", investmentUUID, assumptionID).First(&existing).Error; err != nil {
		return model2.InvestmentSaleAssumption{}, err
	}
	return existing.ToModel(), nil
}

func (s *GormStore) DeleteInvestmentSaleAssumption(investmentUUID string, assumptionID int64) error {
	result := s.db.Delete(&InvestmentSaleAssumptionRow{}, "investment_uuid = ? AND id = ?", investmentUUID, assumptionID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *GormStore) DocumentExists(ctx context.Context, key string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&DocumentRow{}).Where("doc_key = ?", key).Count(&count).Error
	return count > 0, err
}

func (s *GormStore) InsertDocument(ctx context.Context, doc model2.StoredDocument) error {
	row := DocumentRow{
		DocKey:        doc.Key,
		DocumentType:  doc.DocumentType,
		Ticker:        doc.Ticker,
		FiscalYear:    doc.FiscalYear,
		FiscalQuarter: doc.FiscalQtr,
		Form:          doc.Form,
		SourceURL:     doc.SourceURL,
		OutputLabel:   doc.OutputLabel,
		MimeType:      doc.MimeType,
		Body:          doc.Body,
	}
	return s.db.WithContext(ctx).Create(&row).Error
}

func (s *GormStore) LogRetrievalAttempt(ctx context.Context, attempt model2.RetrievalAttempt) error {
	row := RetrievalAttemptRow{
		DocKey:        attempt.DocKey,
		DocumentType:  attempt.DocumentType,
		Ticker:        attempt.Ticker,
		FiscalYear:    attempt.FiscalYear,
		FiscalQuarter: attempt.FiscalQuarter,
		Status:        attempt.Status,
		Message:       attempt.Message,
	}
	return s.db.WithContext(ctx).Create(&row).Error
}

func (s *GormStore) GetDocumentBody(ctx context.Context, key string) ([]byte, error) {
	var row DocumentRow
	err := s.db.WithContext(ctx).Where("doc_key = ?", key).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return row.Body, nil
}

var _ Store = (*GormStore)(nil)
