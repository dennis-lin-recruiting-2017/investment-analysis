package persistence

import (
	model2 "investment-analysis/persistence/model"
	"time"

	"gorm.io/gorm"
)

const timestampLayout = "2006-01-02 15:04:05"

func currentTimestamp() string {
	return time.Now().UTC().Format(timestampLayout)
}

type LLMSettingsRow struct {
	Provider     string  `gorm:"column:provider;primaryKey"`
	Endpoint     string  `gorm:"column:endpoint;not null"`
	Model        string  `gorm:"column:model;not null"`
	APIKey       string  `gorm:"column:api_key;not null;default:''"`
	Temperature  float64 `gorm:"column:temperature;not null;default:0.7"`
	SystemPrompt string  `gorm:"column:system_prompt;not null;default:''"`
	UpdatedAt    string  `gorm:"column:updated_at;not null"`
}

func (LLMSettingsRow) TableName() string {
	return "llm_settings"
}

func (r *LLMSettingsRow) BeforeCreate(_ *gorm.DB) error {
	r.UpdatedAt = currentTimestamp()
	return nil
}

func (r *LLMSettingsRow) BeforeUpdate(_ *gorm.DB) error {
	r.UpdatedAt = currentTimestamp()
	return nil
}

func (r LLMSettingsRow) ToModel() model2.LLMSettings {
	return model2.LLMSettings{
		Provider:     r.Provider,
		Endpoint:     r.Endpoint,
		Model:        r.Model,
		APIKey:       r.APIKey,
		Temperature:  r.Temperature,
		SystemPrompt: r.SystemPrompt,
	}
}

func NewLLMSettingsRow(settings model2.LLMSettings) LLMSettingsRow {
	return LLMSettingsRow{
		Provider:     settings.Provider,
		Endpoint:     settings.Endpoint,
		Model:        settings.Model,
		APIKey:       settings.APIKey,
		Temperature:  settings.Temperature,
		SystemPrompt: settings.SystemPrompt,
	}
}

type InvestmentRow struct {
	ID                    int64                         `gorm:"column:id;primaryKey;autoIncrement"`
	UUID                  string                        `gorm:"column:uuid;not null;uniqueIndex"`
	Name                  string                        `gorm:"column:name;not null"`
	Ticker                string                        `gorm:"column:ticker;not null;default:''"`
	AssetClass            string                        `gorm:"column:asset_class;not null"`
	PurchasePrice         float64                       `gorm:"column:purchase_price;not null;default:0"`
	CouponRate            float64                       `gorm:"column:coupon_rate;not null;default:0"`
	MaturityDate          string                        `gorm:"column:maturity_date;not null;default:''"`
	CallableDateStart     string                        `gorm:"column:callable_date_start;not null;default:''"`
	CallPrice             float64                       `gorm:"column:call_price;not null;default:0"`
	CallDate              string                        `gorm:"column:call_date;not null;default:''"`
	Thesis                string                        `gorm:"column:thesis;not null"`
	TargetAllocation      string                        `gorm:"column:target_allocation;not null;default:''"`
	InitialInvestment     float64                       `gorm:"column:initial_investment;not null;default:0"`
	InitialInvestmentDate string                        `gorm:"column:initial_investment_date;not null;default:''"`
	Notes                 string                        `gorm:"column:notes;not null;default:''"`
	Categories            []InvestmentCategoryRow       `gorm:"foreignKey:InvestmentUUID;references:UUID"`
	Expenses              []InvestmentExpenseRow        `gorm:"foreignKey:InvestmentUUID;references:UUID"`
	SaleAssumptions       []InvestmentSaleAssumptionRow `gorm:"foreignKey:InvestmentUUID;references:UUID"`
	CreatedAt             string                        `gorm:"column:created_at;not null"`
	UpdatedAt             string                        `gorm:"column:updated_at;not null"`
}

func (InvestmentRow) TableName() string {
	return "investments"
}

func (r *InvestmentRow) BeforeCreate(_ *gorm.DB) error {
	now := currentTimestamp()
	if r.CreatedAt == "" {
		r.CreatedAt = now
	}
	r.UpdatedAt = now
	return nil
}

func (r *InvestmentRow) BeforeUpdate(_ *gorm.DB) error {
	r.UpdatedAt = currentTimestamp()
	return nil
}

func (r InvestmentRow) ToModel() model2.Investment {
	categories := make([]string, 0, len(r.Categories))
	for _, category := range r.Categories {
		categories = append(categories, category.Name)
	}

	expenses := make([]model2.InvestmentExpense, 0, len(r.Expenses))
	for _, expense := range r.Expenses {
		expenses = append(expenses, expense.ToModel())
	}

	saleAssumptions := make([]model2.InvestmentSaleAssumption, 0, len(r.SaleAssumptions))
	for _, assumption := range r.SaleAssumptions {
		saleAssumptions = append(saleAssumptions, assumption.ToModel())
	}

	return model2.Investment{
		UUID:                  r.UUID,
		Name:                  r.Name,
		Ticker:                r.Ticker,
		AssetClass:            r.AssetClass,
		PurchasePrice:         r.PurchasePrice,
		CouponRate:            r.CouponRate,
		MaturityDate:          r.MaturityDate,
		CallableDateStart:     r.CallableDateStart,
		CallPrice:             r.CallPrice,
		CallDate:              r.CallDate,
		Thesis:                r.Thesis,
		TargetAllocation:      r.TargetAllocation,
		InitialInvestment:     r.InitialInvestment,
		InitialInvestmentDate: r.InitialInvestmentDate,
		Notes:                 r.Notes,
		Categories:            categories,
		Expenses:              expenses,
		SaleAssumptions:       saleAssumptions,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
	}
}

func NewInvestmentRow(investment model2.Investment) InvestmentRow {
	return InvestmentRow{
		UUID:                  investment.UUID,
		Name:                  investment.Name,
		Ticker:                investment.Ticker,
		AssetClass:            investment.AssetClass,
		PurchasePrice:         investment.PurchasePrice,
		CouponRate:            investment.CouponRate,
		MaturityDate:          investment.MaturityDate,
		CallableDateStart:     investment.CallableDateStart,
		CallPrice:             investment.CallPrice,
		CallDate:              investment.CallDate,
		Thesis:                investment.Thesis,
		TargetAllocation:      investment.TargetAllocation,
		InitialInvestment:     investment.InitialInvestment,
		InitialInvestmentDate: investment.InitialInvestmentDate,
		Notes:                 investment.Notes,
		CreatedAt:             investment.CreatedAt,
		UpdatedAt:             investment.UpdatedAt,
	}
}

type InvestmentCategoryRow struct {
	ID             int64  `gorm:"column:id;primaryKey;autoIncrement"`
	InvestmentUUID string `gorm:"column:investment_uuid;not null;uniqueIndex:idx_investment_categories_name"`
	Name           string `gorm:"column:name;not null;uniqueIndex:idx_investment_categories_name"`
	CreatedAt      string `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      string `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

func (InvestmentCategoryRow) TableName() string {
	return "investment_categories"
}

func (r *InvestmentCategoryRow) BeforeCreate(_ *gorm.DB) error {
	now := currentTimestamp()
	if r.CreatedAt == "" {
		r.CreatedAt = now
	}
	r.UpdatedAt = now
	return nil
}

func (r *InvestmentCategoryRow) BeforeUpdate(_ *gorm.DB) error {
	r.UpdatedAt = currentTimestamp()
	return nil
}

type InvestmentExpenseRow struct {
	ID                  int64   `gorm:"column:id;primaryKey;autoIncrement"`
	InvestmentUUID      string  `gorm:"column:investment_uuid;not null;index"`
	EventType           string  `gorm:"column:event_type;not null;default:'cash-flow'"`
	FlowType            string  `gorm:"column:flow_type;not null;default:'one-time'"`
	RecurrenceInterval  string  `gorm:"column:recurrence_interval;not null;default:''"`
	Label               string  `gorm:"column:label;not null"`
	Amount              float64 `gorm:"column:amount;not null"`
	Description         string  `gorm:"column:description;not null;default:''"`
	StartDate           string  `gorm:"column:start_date;not null;default:''"`
	EndDate             string  `gorm:"column:end_date;not null;default:''"`
	DueDate             string  `gorm:"column:due_date;not null;default:''"`
	Category            string  `gorm:"column:category;not null;default:''"`
	AdjustmentFrequency string  `gorm:"column:adjustment_frequency;not null;default:''"`
	AdjustmentMode      string  `gorm:"column:adjustment_mode;not null;default:''"`
	AdjustmentValue     float64 `gorm:"column:adjustment_value;not null;default:0"`
	Notes               string  `gorm:"column:notes;not null;default:''"`
	CreatedAt           string  `gorm:"column:created_at;not null"`
	UpdatedAt           string  `gorm:"column:updated_at;not null"`
}

type InvestmentSaleAssumptionRow struct {
	ID             int64   `gorm:"column:id;primaryKey;autoIncrement"`
	InvestmentUUID string  `gorm:"column:investment_uuid;not null;index"`
	Label          string  `gorm:"column:label;not null"`
	Amount         float64 `gorm:"column:amount;not null"`
	GrowthType     string  `gorm:"column:growth_type;not null;default:'fixed'"`
	GrowthPeriod   string  `gorm:"column:growth_period;not null;default:''"`
	GrowthMode     string  `gorm:"column:growth_mode;not null;default:''"`
	GrowthValue    float64 `gorm:"column:growth_value;not null;default:0"`
	Category       string  `gorm:"column:category;not null;default:''"`
	Description    string  `gorm:"column:description;not null;default:''"`
	StartDate      string  `gorm:"column:start_date;not null;default:''"`
	EndDate        string  `gorm:"column:end_date;not null;default:''"`
	Notes          string  `gorm:"column:notes;not null;default:''"`
	CreatedAt      string  `gorm:"column:created_at;not null"`
	UpdatedAt      string  `gorm:"column:updated_at;not null"`
}

func (InvestmentSaleAssumptionRow) TableName() string {
	return "investment_sale_assumptions"
}

func (r *InvestmentSaleAssumptionRow) BeforeCreate(_ *gorm.DB) error {
	now := currentTimestamp()
	if r.CreatedAt == "" {
		r.CreatedAt = now
	}
	r.UpdatedAt = now
	return nil
}

func (r *InvestmentSaleAssumptionRow) BeforeUpdate(_ *gorm.DB) error {
	r.UpdatedAt = currentTimestamp()
	return nil
}

func (r InvestmentSaleAssumptionRow) ToModel() model2.InvestmentSaleAssumption {
	return model2.InvestmentSaleAssumption{
		ID:             r.ID,
		InvestmentUUID: r.InvestmentUUID,
		Label:          r.Label,
		Amount:         r.Amount,
		GrowthType:     r.GrowthType,
		GrowthPeriod:   r.GrowthPeriod,
		GrowthMode:     r.GrowthMode,
		GrowthValue:    r.GrowthValue,
		Category:       r.Category,
		Description:    r.Description,
		StartDate:      r.StartDate,
		EndDate:        r.EndDate,
		Notes:          r.Notes,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

func NewInvestmentSaleAssumptionRow(assumption model2.InvestmentSaleAssumption) InvestmentSaleAssumptionRow {
	return InvestmentSaleAssumptionRow{
		ID:             assumption.ID,
		InvestmentUUID: assumption.InvestmentUUID,
		Label:          assumption.Label,
		Amount:         assumption.Amount,
		GrowthType:     assumption.GrowthType,
		GrowthPeriod:   assumption.GrowthPeriod,
		GrowthMode:     assumption.GrowthMode,
		GrowthValue:    assumption.GrowthValue,
		Category:       assumption.Category,
		Description:    assumption.Description,
		StartDate:      assumption.StartDate,
		EndDate:        assumption.EndDate,
		Notes:          assumption.Notes,
		CreatedAt:      assumption.CreatedAt,
		UpdatedAt:      assumption.UpdatedAt,
	}
}

func (InvestmentExpenseRow) TableName() string {
	return "investment_expenses"
}

func (r *InvestmentExpenseRow) BeforeCreate(_ *gorm.DB) error {
	now := currentTimestamp()
	if r.CreatedAt == "" {
		r.CreatedAt = now
	}
	r.UpdatedAt = now
	return nil
}

func (r *InvestmentExpenseRow) BeforeUpdate(_ *gorm.DB) error {
	r.UpdatedAt = currentTimestamp()
	return nil
}

func (r InvestmentExpenseRow) ToModel() model2.InvestmentExpense {
	return model2.InvestmentExpense{
		ID:                  r.ID,
		InvestmentUUID:      r.InvestmentUUID,
		EventType:           r.EventType,
		FlowType:            r.FlowType,
		RecurrenceInterval:  r.RecurrenceInterval,
		Label:               r.Label,
		Amount:              r.Amount,
		Description:         r.Description,
		StartDate:           r.StartDate,
		EndDate:             r.EndDate,
		DueDate:             r.DueDate,
		Category:            r.Category,
		AdjustmentFrequency: r.AdjustmentFrequency,
		AdjustmentMode:      r.AdjustmentMode,
		AdjustmentValue:     r.AdjustmentValue,
		Notes:               r.Notes,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
	}
}

func NewInvestmentExpenseRow(expense model2.InvestmentExpense) InvestmentExpenseRow {
	return InvestmentExpenseRow{
		ID:                  expense.ID,
		InvestmentUUID:      expense.InvestmentUUID,
		EventType:           expense.EventType,
		FlowType:            expense.FlowType,
		RecurrenceInterval:  expense.RecurrenceInterval,
		Label:               expense.Label,
		Amount:              expense.Amount,
		Description:         expense.Description,
		StartDate:           expense.StartDate,
		EndDate:             expense.EndDate,
		DueDate:             expense.DueDate,
		Category:            expense.Category,
		AdjustmentFrequency: expense.AdjustmentFrequency,
		AdjustmentMode:      expense.AdjustmentMode,
		AdjustmentValue:     expense.AdjustmentValue,
		Notes:               expense.Notes,
		CreatedAt:           expense.CreatedAt,
		UpdatedAt:           expense.UpdatedAt,
	}
}

// DocumentRow is the GORM model for the documents table.
type DocumentRow struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement"`
	DocKey        string `gorm:"column:doc_key;not null;uniqueIndex;default:''"`
	DocumentType  string `gorm:"column:document_type;not null;default:''"`
	Ticker        string `gorm:"column:ticker;not null;default:''"`
	FiscalYear    int    `gorm:"column:fiscal_year;not null;default:0"`
	FiscalQuarter int    `gorm:"column:fiscal_quarter;not null;default:0"`
	Form          string `gorm:"column:form;not null;default:''"`
	SourceURL     string `gorm:"column:source_url;not null;default:''"`
	OutputLabel   string `gorm:"column:output_label;not null;default:''"`
	MimeType      string `gorm:"column:mime_type;not null;default:''"`
	Body          []byte `gorm:"column:body;not null"`
	CreatedAt     string `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
}

func (DocumentRow) TableName() string { return "documents" }

func (r *DocumentRow) BeforeCreate(_ *gorm.DB) error {
	if r.CreatedAt == "" {
		r.CreatedAt = currentTimestamp()
	}
	return nil
}

func (r DocumentRow) ToModel() model2.StoredDocument {
	return model2.StoredDocument{
		Key:          r.DocKey,
		DocumentType: model2.DocumentType(r.DocumentType),
		Ticker:       r.Ticker,
		FiscalYear:  r.FiscalYear,
		FiscalQtr:   r.FiscalQuarter,
		Form:        r.Form,
		SourceURL:   r.SourceURL,
		OutputLabel: r.OutputLabel,
		MimeType:    r.MimeType,
		Body:        r.Body,
	}
}

// RetrievalAttemptRow is the GORM model for the retrieval_attempts table.
type RetrievalAttemptRow struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement"`
	DocKey        string `gorm:"column:doc_key;not null;index;default:''"`
	DocumentType  string `gorm:"column:document_type;not null;default:''"`
	Ticker        string `gorm:"column:ticker;not null;default:''"`
	FiscalYear    int    `gorm:"column:fiscal_year;not null;default:0"`
	FiscalQuarter int    `gorm:"column:fiscal_quarter;not null;default:0"`
	Status        string `gorm:"column:status;not null;default:''"`
	Message       string `gorm:"column:message;not null;default:''"`
	CreatedAt     string `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
}

func (RetrievalAttemptRow) TableName() string { return "retrieval_attempts" }

func (r *RetrievalAttemptRow) BeforeCreate(_ *gorm.DB) error {
	if r.CreatedAt == "" {
		r.CreatedAt = currentTimestamp()
	}
	return nil
}

// RetrievalSettingsRow is the GORM model for the retrieval_settings table.
// The table always holds exactly one row (id = 1).
type RetrievalSettingsRow struct {
	ID                       int    `gorm:"column:id;primaryKey;default:1"`
	PlaywrightTimeoutSeconds int    `gorm:"column:playwright_timeout_seconds;not null;default:300"`
	UpdatedAt                string `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

func (RetrievalSettingsRow) TableName() string { return "retrieval_settings" }

func (r *RetrievalSettingsRow) BeforeCreate(_ *gorm.DB) error {
	if r.UpdatedAt == "" {
		r.UpdatedAt = currentTimestamp()
	}
	return nil
}

func (r *RetrievalSettingsRow) BeforeUpdate(_ *gorm.DB) error {
	r.UpdatedAt = currentTimestamp()
	return nil
}

func (r RetrievalSettingsRow) ToModel() model2.RetrievalSettings {
	return model2.RetrievalSettings{
		PlaywrightTimeoutSeconds: r.PlaywrightTimeoutSeconds,
	}
}
