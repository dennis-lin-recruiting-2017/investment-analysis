package model

import (
	"errors"
	"investment-analysis/util"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// expenseOrderClause is the canonical sort order for investment_expenses
// rows: by end_date (falling back to due_date), then by id for
// determinism among same-day rows.
const expenseOrderClause = "COALESCE(NULLIF(end_date, ''), due_date) ASC, id ASC"

// normalizeCategoryName trims whitespace and treats "uncategorized" (in
// any case) as the empty string.
func normalizeCategoryName(name string) string {
	trimmed := strings.TrimSpace(name)
	if strings.EqualFold(trimmed, "uncategorized") {
		return ""
	}
	return trimmed
}

// ── Investment ────────────────────────────────────────────────────────────────

// Investment is the row for the investments table; the same struct
// serializes as JSON.  CategoryRows is the GORM relation; it is hidden
// from JSON.  Categories is the JSON-only flat list of category names,
// populated by InvestmentsTable.Get from CategoryRows.
type Investment struct {
	UUID                  string  `gorm:"column:uuid;primaryKey"                             json:"uuid"`
	Name                  string  `gorm:"column:name;not null"                               json:"name"`
	Ticker                string  `gorm:"column:ticker;not null;default:''"                  json:"ticker"`
	AssetClass            string  `gorm:"column:asset_class;not null"                        json:"assetClass"`
	PurchasePrice         float64 `gorm:"column:purchase_price;not null;default:0"           json:"purchasePrice"`
	CouponRate            float64 `gorm:"column:coupon_rate;not null;default:0"              json:"couponRate"`
	MaturityDate          string  `gorm:"column:maturity_date;not null;default:''"           json:"maturityDate"`
	CallableDateStart     string  `gorm:"column:callable_date_start;not null;default:''"     json:"callableDateStart"`
	CallPrice             float64 `gorm:"column:call_price;not null;default:0"               json:"callPrice"`
	CallDate              string  `gorm:"column:call_date;not null;default:''"               json:"callDate"`
	Thesis                string  `gorm:"column:thesis;not null"                             json:"thesis"`
	TargetAllocation      string  `gorm:"column:target_allocation;not null;default:''"       json:"targetAllocation"`
	InitialInvestment     float64 `gorm:"column:initial_investment;not null;default:0"       json:"initialInvestment"`
	InitialInvestmentDate string  `gorm:"column:initial_investment_date;not null;default:''" json:"initialInvestmentDate"`
	Notes                 string  `gorm:"column:notes;not null;default:''"                   json:"notes"`

	CategoryRows    []InvestmentCategory       `gorm:"foreignKey:InvestmentUUID;references:UUID" json:"-"`
	Expenses        []InvestmentExpense        `gorm:"foreignKey:InvestmentUUID;references:UUID" json:"expenses,omitempty"`
	SaleAssumptions []InvestmentSaleAssumption `gorm:"foreignKey:InvestmentUUID;references:UUID" json:"saleAssumptions,omitempty"`

	// Categories is the JSON-only flat list of category names.  It is
	// populated by InvestmentsTable.Get from CategoryRows.
	Categories []string `gorm:"-" json:"categories,omitempty"`

	CreatedAt string `gorm:"column:created_at;not null" json:"createdAt,omitempty"`
	UpdatedAt string `gorm:"column:updated_at;not null" json:"updatedAt,omitempty"`
}

func (Investment) TableName() string { return "investments" }

func (i *Investment) BeforeCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() { util.LogIfErr(ctx, &err, "Investment.BeforeCreate", "uuid", i.UUID, "name", i.Name) }()
	if i.UUID == "" {
		i.UUID = NewSortableID()
	}
	now := CurrentTimestamp()
	if i.CreatedAt == "" {
		i.CreatedAt = now
	}
	i.UpdatedAt = now
	return nil
}

func (i *Investment) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() { util.LogIfErr(ctx, &err, "Investment.BeforeUpdate", "uuid", i.UUID) }()
	i.UpdatedAt = CurrentTimestamp()
	return nil
}

// investmentScalarFields returns the column-keyed map of editable
// scalar fields on Investment, used by Update / InsertOrUpdate /
// UpdateByUUID to avoid touching GORM relation fields (CategoryRows,
// Expenses, SaleAssumptions) and the auto-managed timestamps.
func investmentScalarFields(inv *Investment) map[string]any {
	return map[string]any{
		"name":                    inv.Name,
		"ticker":                  inv.Ticker,
		"asset_class":             inv.AssetClass,
		"purchase_price":          inv.PurchasePrice,
		"coupon_rate":             inv.CouponRate,
		"maturity_date":           inv.MaturityDate,
		"callable_date_start":     inv.CallableDateStart,
		"call_price":              inv.CallPrice,
		"call_date":               inv.CallDate,
		"thesis":                  inv.Thesis,
		"target_allocation":       inv.TargetAllocation,
		"initial_investment":      inv.InitialInvestment,
		"initial_investment_date": inv.InitialInvestmentDate,
		"notes":                   inv.Notes,
	}
}

// InvestmentsTable provides CRUD on the investments table.  Methods
// that affect child tables (Delete cascades into expenses, categories,
// and sale_assumptions) are implemented here as transactions; the
// child table structs only handle their own rows.
type InvestmentsTable struct {
	db *gorm.DB
}

// NewInvestmentsTable wraps an open *gorm.DB as an InvestmentsTable.
func NewInvestmentsTable(db *gorm.DB) *InvestmentsTable {
	return &InvestmentsTable{db: db}
}

func (t *InvestmentsTable) List() (out []Investment, err error) {
	defer func() { util.LogIfErr(nil, &err, "InvestmentsTable.List") }()
	var rows []Investment
	if err = t.db.Order("updated_at DESC").Order("uuid DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// Get returns the investment with the given UUID, including its
// categories, expenses, and sale assumptions.  The flat
// Investment.Categories list of names is populated from the preloaded
// CategoryRows so the JSON shape matches the existing API contract.
func (t *InvestmentsTable) Get(investmentUUID string) (out Investment, err error) {
	defer func() { util.LogIfErr(nil, &err, "InvestmentsTable.Get", "uuid", investmentUUID) }()
	var row Investment
	err = t.db.Preload("CategoryRows", func(db *gorm.DB) *gorm.DB {
		return db.Order("name ASC")
	}).Preload("Expenses", func(db *gorm.DB) *gorm.DB {
		return db.Order(expenseOrderClause)
	}).Preload("SaleAssumptions", func(db *gorm.DB) *gorm.DB {
		return db.Order("COALESCE(NULLIF(end_date, ''), start_date) ASC, id ASC")
	}).Where("uuid = ?", investmentUUID).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Investment{}, ErrNotFound
		}
		return Investment{}, err
	}
	row.Categories = make([]string, 0, len(row.CategoryRows))
	for _, c := range row.CategoryRows {
		row.Categories = append(row.Categories, c.Name)
	}
	return row, nil
}

// Insert persists inv.  Returns an error if a row with the same UUID
// already exists.  inv.UUID is auto-assigned if empty.
func (t *InvestmentsTable) Insert(inv *Investment) (err error) {
	defer func() { util.LogIfErr(nil, &err, "InvestmentsTable.Insert", "uuid", inv.UUID, "name", inv.Name) }()
	return t.db.Create(inv).Error
}

// Update modifies the existing row matching inv.UUID, updating only
// the editable scalar fields (relations and auto-managed timestamps
// are left untouched).  Returns ErrNotFound if no such row exists.
func (t *InvestmentsTable) Update(inv *Investment) (err error) {
	defer func() { util.LogIfErr(nil, &err, "InvestmentsTable.Update", "uuid", inv.UUID) }()
	return t.db.Transaction(func(tx *gorm.DB) error {
		var existing Investment
		if err := tx.Where("uuid = ?", inv.UUID).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		return tx.Model(&existing).Updates(investmentScalarFields(inv)).Error
	})
}

// InsertOrUpdate persists inv, inserting a new row if inv.UUID does
// not exist or updating the existing row otherwise.  The returned bool
// is true if a new row was inserted.
func (t *InvestmentsTable) InsertOrUpdate(inv *Investment) (inserted bool, err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentsTable.InsertOrUpdate", "uuid", inv.UUID, "name", inv.Name)
	}()
	if inv.UUID == "" {
		inv.UUID = NewSortableID()
	}
	err = t.db.Transaction(func(tx *gorm.DB) error {
		var existing Investment
		err := tx.Where("uuid = ?", inv.UUID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			inserted = true
			return tx.Create(inv).Error
		}
		if err != nil {
			return err
		}
		return tx.Model(&existing).Updates(investmentScalarFields(inv)).Error
	})
	return inserted, err
}

// UpdateByUUID is the legacy update method that looks up an existing
// row by URL-supplied investmentUUID and writes the inv body's scalar
// fields to it; it returns the freshly loaded Investment (with
// preloads).  Prefer Update for new code.
func (t *InvestmentsTable) UpdateByUUID(investmentUUID string, inv Investment) (out Investment, err error) {
	defer func() { util.LogIfErr(nil, &err, "InvestmentsTable.UpdateByUUID", "uuid", investmentUUID) }()
	inv.UUID = investmentUUID
	if err = t.Update(&inv); err != nil {
		return Investment{}, err
	}
	return t.Get(investmentUUID)
}

// Create is a legacy wrapper around Insert that returns the inserted
// Investment (with preloads).
func (t *InvestmentsTable) Create(inv Investment) (out Investment, err error) {
	defer func() { util.LogIfErr(nil, &err, "InvestmentsTable.Create", "uuid", inv.UUID, "name", inv.Name) }()
	if err = t.Insert(&inv); err != nil {
		return Investment{}, err
	}
	return t.Get(inv.UUID)
}

// Delete removes the investment along with its expenses, sale
// assumptions, and categories in a single transaction.
func (t *InvestmentsTable) Delete(investmentUUID string) (err error) {
	defer func() { util.LogIfErr(nil, &err, "InvestmentsTable.Delete", "uuid", investmentUUID) }()
	return t.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&InvestmentExpense{}, "investment_uuid = ?", investmentUUID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&InvestmentSaleAssumption{}, "investment_uuid = ?", investmentUUID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&InvestmentCategory{}, "investment_uuid = ?", investmentUUID).Error; err != nil {
			return err
		}
		result := tx.Delete(&Investment{}, "uuid = ?", investmentUUID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// ── InvestmentCategory ────────────────────────────────────────────────────────

// InvestmentCategory is one category row attached to an investment.
type InvestmentCategory struct {
	ID             string `gorm:"column:id;primaryKey"                                                       json:"id"`
	InvestmentUUID string `gorm:"column:investment_uuid;not null;uniqueIndex:idx_investment_categories_name" json:"investmentUuid,omitempty"`
	Name           string `gorm:"column:name;not null;uniqueIndex:idx_investment_categories_name"            json:"name"`
	CreatedAt      string `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"                       json:"createdAt,omitempty"`
	UpdatedAt      string `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"                       json:"updatedAt,omitempty"`
}

func (InvestmentCategory) TableName() string { return "investment_categories" }

func (c *InvestmentCategory) BeforeCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() {
		util.LogIfErr(ctx, &err, "InvestmentCategory.BeforeCreate", "id", c.ID, "investmentUuid", c.InvestmentUUID, "name", c.Name)
	}()
	if c.ID == "" {
		c.ID = NewSortableID()
	}
	now := CurrentTimestamp()
	if c.CreatedAt == "" {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	return nil
}

func (c *InvestmentCategory) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() { util.LogIfErr(ctx, &err, "InvestmentCategory.BeforeUpdate", "id", c.ID) }()
	c.UpdatedAt = CurrentTimestamp()
	return nil
}

// InvestmentCategoriesTable provides CRUD on the investment_categories
// table.  Rename propagates the rename to investment_expenses; Delete
// blanks out the category column on associated expenses.
type InvestmentCategoriesTable struct {
	db *gorm.DB
}

// NewInvestmentCategoriesTable wraps an open *gorm.DB as an
// InvestmentCategoriesTable.
func NewInvestmentCategoriesTable(db *gorm.DB) *InvestmentCategoriesTable {
	return &InvestmentCategoriesTable{db: db}
}

func (t *InvestmentCategoriesTable) List(investmentUUID string) (out []string, err error) {
	defer func() { util.LogIfErr(nil, &err, "InvestmentCategoriesTable.List", "investmentUuid", investmentUUID) }()
	var rows []InvestmentCategory
	if err = t.db.Where("investment_uuid = ?", investmentUUID).Order("name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out = make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Name)
	}
	return out, nil
}

// Create adds a category by name (skip if it already exists).  This is
// the by-name semantic helper used by the HTTP API.
func (t *InvestmentCategoriesTable) Create(investmentUUID, name string) (err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentCategoriesTable.Create", "investmentUuid", investmentUUID, "name", name)
	}()
	name = normalizeCategoryName(name)
	if name == "" {
		return nil
	}
	row := InvestmentCategory{
		InvestmentUUID: investmentUUID,
		Name:           name,
	}
	return t.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

// Rename changes a category's name and propagates the rename to
// matching investment_expenses rows.  Returns ErrNotFound if no
// category matches (investmentUUID, currentName).
func (t *InvestmentCategoriesTable) Rename(investmentUUID, currentName, newName string) (err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentCategoriesTable.Rename", "investmentUuid", investmentUUID, "currentName", currentName, "newName", newName)
	}()
	currentName = normalizeCategoryName(currentName)
	newName = normalizeCategoryName(newName)
	if currentName == "" || newName == "" {
		return ErrNotFound
	}

	return t.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&InvestmentCategory{}).
			Where("investment_uuid = ? AND name = ?", investmentUUID, currentName).
			Update("name", newName)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return tx.Model(&InvestmentExpense{}).
			Where("investment_uuid = ? AND category = ?", investmentUUID, currentName).
			Update("category", newName).Error
	})
}

// Delete removes a category by name and blanks out the category column
// on associated investment_expenses rows.
func (t *InvestmentCategoriesTable) Delete(investmentUUID, name string) (err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentCategoriesTable.Delete", "investmentUuid", investmentUUID, "name", name)
	}()
	name = normalizeCategoryName(name)
	if name == "" {
		return ErrNotFound
	}
	return t.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&InvestmentExpense{}).
			Where("investment_uuid = ? AND category = ?", investmentUUID, name).
			Update("category", "").Error; err != nil {
			return err
		}
		result := tx.Delete(&InvestmentCategory{}, "investment_uuid = ? AND name = ?", investmentUUID, name)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// Insert persists c.  Returns an error if a row with the same ID (or
// (investment_uuid, name) uniqueness constraint) already exists.
func (t *InvestmentCategoriesTable) Insert(c *InvestmentCategory) (err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentCategoriesTable.Insert", "id", c.ID, "investmentUuid", c.InvestmentUUID, "name", c.Name)
	}()
	return t.db.Create(c).Error
}

// Update modifies the existing row matching c.ID.  Returns ErrNotFound
// if no such row exists.
func (t *InvestmentCategoriesTable) Update(c *InvestmentCategory) (err error) {
	defer func() { util.LogIfErr(nil, &err, "InvestmentCategoriesTable.Update", "id", c.ID, "name", c.Name) }()
	return t.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&InvestmentCategory{}).Where("id = ?", c.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrNotFound
		}
		return tx.Save(c).Error
	})
}

// InsertOrUpdate persists c, inserting a new row if c.ID does not
// exist or updating the existing row otherwise.  The returned bool is
// true if a new row was inserted.
func (t *InvestmentCategoriesTable) InsertOrUpdate(c *InvestmentCategory) (inserted bool, err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentCategoriesTable.InsertOrUpdate", "id", c.ID, "investmentUuid", c.InvestmentUUID, "name", c.Name)
	}()
	if c.ID == "" {
		c.ID = NewSortableID()
	}
	err = t.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&InvestmentCategory{}).Where("id = ?", c.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			inserted = true
			return tx.Create(c).Error
		}
		return tx.Save(c).Error
	})
	return inserted, err
}

// ── InvestmentExpense ─────────────────────────────────────────────────────────

// InvestmentExpense is one cash-flow row attached to an investment.
type InvestmentExpense struct {
	ID                  string  `gorm:"column:id;primaryKey"                            json:"id"`
	InvestmentUUID      string  `gorm:"column:investment_uuid;not null;index"           json:"investmentUuid,omitempty"`
	EventType           string  `gorm:"column:event_type;not null;default:'cash-flow'"  json:"eventType"`
	FlowType            string  `gorm:"column:flow_type;not null;default:'one-time'"    json:"flowType"`
	RecurrenceInterval  string  `gorm:"column:recurrence_interval;not null;default:''"  json:"recurrenceInterval"`
	Label               string  `gorm:"column:label;not null"                           json:"label"`
	Amount              float64 `gorm:"column:amount;not null"                          json:"amount"`
	Description         string  `gorm:"column:description;not null;default:''"          json:"description"`
	StartDate           string  `gorm:"column:start_date;not null;default:''"           json:"startDate"`
	EndDate             string  `gorm:"column:end_date;not null;default:''"             json:"endDate"`
	DueDate             string  `gorm:"column:due_date;not null;default:''"             json:"dueDate"`
	Category            string  `gorm:"column:category;not null;default:''"             json:"category"`
	AdjustmentFrequency string  `gorm:"column:adjustment_frequency;not null;default:''" json:"adjustmentFrequency"`
	AdjustmentMode      string  `gorm:"column:adjustment_mode;not null;default:''"      json:"adjustmentMode"`
	AdjustmentValue     float64 `gorm:"column:adjustment_value;not null;default:0"      json:"adjustmentValue"`
	Notes               string  `gorm:"column:notes;not null;default:''"                json:"notes"`
	CreatedAt           string  `gorm:"column:created_at;not null"                      json:"createdAt,omitempty"`
	UpdatedAt           string  `gorm:"column:updated_at;not null"                      json:"updatedAt,omitempty"`
}

func (InvestmentExpense) TableName() string { return "investment_expenses" }

func (e *InvestmentExpense) BeforeCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() {
		util.LogIfErr(ctx, &err, "InvestmentExpense.BeforeCreate", "id", e.ID, "investmentUuid", e.InvestmentUUID, "label", e.Label)
	}()
	if e.ID == "" {
		e.ID = NewSortableID()
	}
	now := CurrentTimestamp()
	if e.CreatedAt == "" {
		e.CreatedAt = now
	}
	e.UpdatedAt = now
	return nil
}

func (e *InvestmentExpense) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() { util.LogIfErr(ctx, &err, "InvestmentExpense.BeforeUpdate", "id", e.ID) }()
	e.UpdatedAt = CurrentTimestamp()
	return nil
}

// expenseScalarFields returns the column-keyed map of editable scalar
// fields on InvestmentExpense.
func expenseScalarFields(e *InvestmentExpense) map[string]any {
	return map[string]any{
		"event_type":           e.EventType,
		"flow_type":            e.FlowType,
		"recurrence_interval":  e.RecurrenceInterval,
		"label":                e.Label,
		"amount":               e.Amount,
		"description":          e.Description,
		"start_date":           e.StartDate,
		"end_date":             e.EndDate,
		"due_date":             e.DueDate,
		"category":             e.Category,
		"adjustment_frequency": e.AdjustmentFrequency,
		"adjustment_mode":      e.AdjustmentMode,
		"adjustment_value":     e.AdjustmentValue,
		"notes":                e.Notes,
	}
}

// InvestmentExpensesTable provides CRUD on the investment_expenses
// table.
type InvestmentExpensesTable struct {
	db *gorm.DB
}

// NewInvestmentExpensesTable wraps an open *gorm.DB as an
// InvestmentExpensesTable.
func NewInvestmentExpensesTable(db *gorm.DB) *InvestmentExpensesTable {
	return &InvestmentExpensesTable{db: db}
}

func (t *InvestmentExpensesTable) List(investmentUUID string) (out []InvestmentExpense, err error) {
	defer func() { util.LogIfErr(nil, &err, "InvestmentExpensesTable.List", "investmentUuid", investmentUUID) }()
	var rows []InvestmentExpense
	if err = t.db.Where("investment_uuid = ?", investmentUUID).Order(expenseOrderClause).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// Create is a legacy wrapper that takes an investmentUUID separately
// (overriding any value on the expense) and returns the inserted
// InvestmentExpense.
func (t *InvestmentExpensesTable) Create(investmentUUID string, expense InvestmentExpense) (out InvestmentExpense, err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentExpensesTable.Create", "investmentUuid", investmentUUID, "label", expense.Label)
	}()
	expense.InvestmentUUID = investmentUUID
	if err = t.Insert(&expense); err != nil {
		return InvestmentExpense{}, err
	}
	return expense, nil
}

// UpdateByID is the legacy update that scopes by both
// (investmentUUID, expenseID) and returns the freshly loaded row.
// Prefer Update for new code.
func (t *InvestmentExpensesTable) UpdateByID(investmentUUID, expenseID string, expense InvestmentExpense) (out InvestmentExpense, err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentExpensesTable.UpdateByID", "investmentUuid", investmentUUID, "id", expenseID)
	}()
	var existing InvestmentExpense
	if err = t.db.Where("investment_uuid = ? AND id = ?", investmentUUID, expenseID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return InvestmentExpense{}, ErrNotFound
		}
		return InvestmentExpense{}, err
	}
	if err = t.db.Model(&existing).Updates(expenseScalarFields(&expense)).Error; err != nil {
		return InvestmentExpense{}, err
	}
	if err = t.db.Where("investment_uuid = ? AND id = ?", investmentUUID, expenseID).First(&existing).Error; err != nil {
		return InvestmentExpense{}, err
	}
	return existing, nil
}

// Insert persists expense.  Returns an error if a row with the same ID
// already exists.  expense.ID is auto-assigned if empty.
func (t *InvestmentExpensesTable) Insert(expense *InvestmentExpense) (err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentExpensesTable.Insert", "id", expense.ID, "investmentUuid", expense.InvestmentUUID, "label", expense.Label)
	}()
	return t.db.Create(expense).Error
}

// Update modifies the existing row matching expense.ID, updating only
// editable scalar fields.  Returns ErrNotFound if no such row exists.
func (t *InvestmentExpensesTable) Update(expense *InvestmentExpense) (err error) {
	defer func() { util.LogIfErr(nil, &err, "InvestmentExpensesTable.Update", "id", expense.ID) }()
	return t.db.Transaction(func(tx *gorm.DB) error {
		var existing InvestmentExpense
		if err := tx.Where("id = ?", expense.ID).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		return tx.Model(&existing).Updates(expenseScalarFields(expense)).Error
	})
}

// InsertOrUpdate persists expense, inserting a new row if expense.ID
// does not exist or updating the existing row otherwise.  The returned
// bool is true if a new row was inserted.
func (t *InvestmentExpensesTable) InsertOrUpdate(expense *InvestmentExpense) (inserted bool, err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentExpensesTable.InsertOrUpdate", "id", expense.ID, "investmentUuid", expense.InvestmentUUID, "label", expense.Label)
	}()
	if expense.ID == "" {
		expense.ID = NewSortableID()
	}
	err = t.db.Transaction(func(tx *gorm.DB) error {
		var existing InvestmentExpense
		err := tx.Where("id = ?", expense.ID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			inserted = true
			return tx.Create(expense).Error
		}
		if err != nil {
			return err
		}
		return tx.Model(&existing).Updates(expenseScalarFields(expense)).Error
	})
	return inserted, err
}

func (t *InvestmentExpensesTable) Delete(investmentUUID, expenseID string) (err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentExpensesTable.Delete", "investmentUuid", investmentUUID, "id", expenseID)
	}()
	result := t.db.Delete(&InvestmentExpense{}, "investment_uuid = ? AND id = ?", investmentUUID, expenseID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ── InvestmentSaleAssumption ──────────────────────────────────────────────────

// InvestmentSaleAssumption is one terminal-value sale-assumption row
// attached to an investment.
type InvestmentSaleAssumption struct {
	ID             string  `gorm:"column:id;primaryKey"                        json:"id"`
	InvestmentUUID string  `gorm:"column:investment_uuid;not null;index"       json:"investmentUuid,omitempty"`
	Label          string  `gorm:"column:label;not null"                       json:"label"`
	Amount         float64 `gorm:"column:amount;not null"                      json:"amount"`
	GrowthType     string  `gorm:"column:growth_type;not null;default:'fixed'" json:"growthType"`
	GrowthPeriod   string  `gorm:"column:growth_period;not null;default:''"    json:"growthPeriod"`
	GrowthMode     string  `gorm:"column:growth_mode;not null;default:''"      json:"growthMode"`
	GrowthValue    float64 `gorm:"column:growth_value;not null;default:0"      json:"growthValue"`
	Category       string  `gorm:"column:category;not null;default:''"         json:"category"`
	Description    string  `gorm:"column:description;not null;default:''"      json:"description"`
	StartDate      string  `gorm:"column:start_date;not null;default:''"       json:"startDate"`
	EndDate        string  `gorm:"column:end_date;not null;default:''"         json:"endDate"`
	Notes          string  `gorm:"column:notes;not null;default:''"            json:"notes"`
	CreatedAt      string  `gorm:"column:created_at;not null"                  json:"createdAt,omitempty"`
	UpdatedAt      string  `gorm:"column:updated_at;not null"                  json:"updatedAt,omitempty"`
}

func (InvestmentSaleAssumption) TableName() string { return "investment_sale_assumptions" }

func (a *InvestmentSaleAssumption) BeforeCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() {
		util.LogIfErr(ctx, &err, "InvestmentSaleAssumption.BeforeCreate", "id", a.ID, "investmentUuid", a.InvestmentUUID, "label", a.Label)
	}()
	if a.ID == "" {
		a.ID = NewSortableID()
	}
	now := CurrentTimestamp()
	if a.CreatedAt == "" {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	return nil
}

func (a *InvestmentSaleAssumption) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() { util.LogIfErr(ctx, &err, "InvestmentSaleAssumption.BeforeUpdate", "id", a.ID) }()
	a.UpdatedAt = CurrentTimestamp()
	return nil
}

// saleAssumptionScalarFields returns the column-keyed map of editable
// scalar fields on InvestmentSaleAssumption.
func saleAssumptionScalarFields(a *InvestmentSaleAssumption) map[string]any {
	return map[string]any{
		"label":         a.Label,
		"amount":        a.Amount,
		"growth_type":   a.GrowthType,
		"growth_period": a.GrowthPeriod,
		"growth_mode":   a.GrowthMode,
		"growth_value":  a.GrowthValue,
		"category":      a.Category,
		"description":   a.Description,
		"start_date":    a.StartDate,
		"end_date":      a.EndDate,
		"notes":         a.Notes,
	}
}

// InvestmentSaleAssumptionsTable provides CRUD on the
// investment_sale_assumptions table.
type InvestmentSaleAssumptionsTable struct {
	db *gorm.DB
}

// NewInvestmentSaleAssumptionsTable wraps an open *gorm.DB as an
// InvestmentSaleAssumptionsTable.
func NewInvestmentSaleAssumptionsTable(db *gorm.DB) *InvestmentSaleAssumptionsTable {
	return &InvestmentSaleAssumptionsTable{db: db}
}

func (t *InvestmentSaleAssumptionsTable) List(investmentUUID string) (out []InvestmentSaleAssumption, err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentSaleAssumptionsTable.List", "investmentUuid", investmentUUID)
	}()
	var rows []InvestmentSaleAssumption
	if err = t.db.Where("investment_uuid = ?", investmentUUID).Order("COALESCE(NULLIF(end_date, ''), start_date) ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// Create is a legacy wrapper that takes an investmentUUID separately
// (overriding any value on the assumption) and returns the inserted
// row.
func (t *InvestmentSaleAssumptionsTable) Create(investmentUUID string, a InvestmentSaleAssumption) (out InvestmentSaleAssumption, err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentSaleAssumptionsTable.Create", "investmentUuid", investmentUUID, "label", a.Label)
	}()
	a.InvestmentUUID = investmentUUID
	if err = t.Insert(&a); err != nil {
		return InvestmentSaleAssumption{}, err
	}
	return a, nil
}

// UpdateByID is the legacy update that scopes by both
// (investmentUUID, assumptionID) and returns the freshly loaded row.
// Prefer Update for new code.
func (t *InvestmentSaleAssumptionsTable) UpdateByID(investmentUUID, assumptionID string, a InvestmentSaleAssumption) (out InvestmentSaleAssumption, err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentSaleAssumptionsTable.UpdateByID", "investmentUuid", investmentUUID, "id", assumptionID)
	}()
	var existing InvestmentSaleAssumption
	if err = t.db.Where("investment_uuid = ? AND id = ?", investmentUUID, assumptionID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return InvestmentSaleAssumption{}, ErrNotFound
		}
		return InvestmentSaleAssumption{}, err
	}
	if err = t.db.Model(&existing).Updates(saleAssumptionScalarFields(&a)).Error; err != nil {
		return InvestmentSaleAssumption{}, err
	}
	if err = t.db.Where("investment_uuid = ? AND id = ?", investmentUUID, assumptionID).First(&existing).Error; err != nil {
		return InvestmentSaleAssumption{}, err
	}
	return existing, nil
}

// Insert persists a.  Returns an error if a row with the same ID
// already exists.  a.ID is auto-assigned if empty.
func (t *InvestmentSaleAssumptionsTable) Insert(a *InvestmentSaleAssumption) (err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentSaleAssumptionsTable.Insert", "id", a.ID, "investmentUuid", a.InvestmentUUID, "label", a.Label)
	}()
	return t.db.Create(a).Error
}

// Update modifies the existing row matching a.ID, updating only
// editable scalar fields.  Returns ErrNotFound if no such row exists.
func (t *InvestmentSaleAssumptionsTable) Update(a *InvestmentSaleAssumption) (err error) {
	defer func() { util.LogIfErr(nil, &err, "InvestmentSaleAssumptionsTable.Update", "id", a.ID) }()
	return t.db.Transaction(func(tx *gorm.DB) error {
		var existing InvestmentSaleAssumption
		if err := tx.Where("id = ?", a.ID).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		return tx.Model(&existing).Updates(saleAssumptionScalarFields(a)).Error
	})
}

// InsertOrUpdate persists a, inserting a new row if a.ID does not
// exist or updating the existing row otherwise.  The returned bool is
// true if a new row was inserted.
func (t *InvestmentSaleAssumptionsTable) InsertOrUpdate(a *InvestmentSaleAssumption) (inserted bool, err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentSaleAssumptionsTable.InsertOrUpdate", "id", a.ID, "investmentUuid", a.InvestmentUUID, "label", a.Label)
	}()
	if a.ID == "" {
		a.ID = NewSortableID()
	}
	err = t.db.Transaction(func(tx *gorm.DB) error {
		var existing InvestmentSaleAssumption
		err := tx.Where("id = ?", a.ID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			inserted = true
			return tx.Create(a).Error
		}
		if err != nil {
			return err
		}
		return tx.Model(&existing).Updates(saleAssumptionScalarFields(a)).Error
	})
	return inserted, err
}

func (t *InvestmentSaleAssumptionsTable) Delete(investmentUUID, assumptionID string) (err error) {
	defer func() {
		util.LogIfErr(nil, &err, "InvestmentSaleAssumptionsTable.Delete", "investmentUuid", investmentUUID, "id", assumptionID)
	}()
	result := t.db.Delete(&InvestmentSaleAssumption{}, "investment_uuid = ? AND id = ?", investmentUUID, assumptionID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
