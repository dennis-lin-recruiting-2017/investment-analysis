package model

type Investment struct {
	UUID                  string                     `json:"uuid"`
	Name                  string                     `json:"name"`
	Ticker                string                     `json:"ticker"`
	AssetClass            string                     `json:"assetClass"`
	PurchasePrice         float64                    `json:"purchasePrice"`
	CouponRate            float64                    `json:"couponRate"`
	MaturityDate          string                     `json:"maturityDate"`
	CallableDateStart     string                     `json:"callableDateStart"`
	CallPrice             float64                    `json:"callPrice"`
	CallDate              string                     `json:"callDate"`
	Thesis                string                     `json:"thesis"`
	TargetAllocation      string                     `json:"targetAllocation"`
	InitialInvestment     float64                    `json:"initialInvestment"`
	InitialInvestmentDate string                     `json:"initialInvestmentDate"`
	Notes                 string                     `json:"notes"`
	Categories            []string                   `json:"categories,omitempty"`
	Expenses              []InvestmentExpense        `json:"expenses,omitempty"`
	SaleAssumptions       []InvestmentSaleAssumption `json:"saleAssumptions,omitempty"`
	CreatedAt             string                     `json:"createdAt,omitempty"`
	UpdatedAt             string                     `json:"updatedAt,omitempty"`
}

type InvestmentSaleAssumption struct {
	ID             int64   `json:"id"`
	InvestmentUUID string  `json:"investmentUuid,omitempty"`
	Label          string  `json:"label"`
	Amount         float64 `json:"amount"`
	GrowthType     string  `json:"growthType"`
	GrowthPeriod   string  `json:"growthPeriod"`
	GrowthMode     string  `json:"growthMode"`
	GrowthValue    float64 `json:"growthValue"`
	Category       string  `json:"category"`
	Description    string  `json:"description"`
	StartDate      string  `json:"startDate"`
	EndDate        string  `json:"endDate"`
	Notes          string  `json:"notes"`
	CreatedAt      string  `json:"createdAt,omitempty"`
	UpdatedAt      string  `json:"updatedAt,omitempty"`
}

type InvestmentExpense struct {
	ID                  int64   `json:"id"`
	InvestmentUUID      string  `json:"investmentUuid,omitempty"`
	EventType           string  `json:"eventType"`
	FlowType            string  `json:"flowType"`
	RecurrenceInterval  string  `json:"recurrenceInterval"`
	Label               string  `json:"label"`
	Amount              float64 `json:"amount"`
	Description         string  `json:"description"`
	StartDate           string  `json:"startDate"`
	EndDate             string  `json:"endDate"`
	DueDate             string  `json:"dueDate"`
	Category            string  `json:"category"`
	AdjustmentFrequency string  `json:"adjustmentFrequency"`
	AdjustmentMode      string  `json:"adjustmentMode"`
	AdjustmentValue     float64 `json:"adjustmentValue"`
	Notes               string  `json:"notes"`
	CreatedAt           string  `json:"createdAt,omitempty"`
	UpdatedAt           string  `json:"updatedAt,omitempty"`
}
