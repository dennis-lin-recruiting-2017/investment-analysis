package model

import (
	"context"
	"errors"
	"investment-analysis/util"

	"gorm.io/gorm"
)

// RetrievalAttempt records the outcome of one document retrieval
// attempt; the same struct is the GORM row and the JSON-serializable
// representation.  InvestmentUUID scopes the attempt to an investment
// (empty for unscoped/global attempts); DocKey is the caller-supplied
// document identifier.
type RetrievalAttempt struct {
	ID             string       `gorm:"column:id;primaryKey"                                                                                  json:"id"`
	InvestmentUUID string       `gorm:"column:investment_uuid;not null;default:'';index;index:idx_retrieval_attempts_investment_key,priority:1" json:"investmentUuid,omitempty"`
	DocKey         string       `gorm:"column:doc_key;not null;index;default:'';index:idx_retrieval_attempts_investment_key,priority:2"        json:"docKey"`
	DocumentType   DocumentType `gorm:"column:document_type;not null;default:''"                                                              json:"documentType"`
	Ticker         string       `gorm:"column:ticker;not null;default:''"                                                                     json:"ticker,omitempty"`
	FiscalYear     int          `gorm:"column:fiscal_year;not null;default:0"                                                                 json:"fiscalYear,omitempty"`
	FiscalQuarter  int          `gorm:"column:fiscal_quarter;not null;default:0"                                                              json:"fiscalQuarter,omitempty"`
	Status         string       `gorm:"column:status;not null;default:''"                                                                     json:"status"`
	Message        string       `gorm:"column:message;not null;default:''"                                                                    json:"message,omitempty"`
	CreatedAt      string       `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"                                                  json:"createdAt,omitempty"`
}

func (RetrievalAttempt) TableName() string { return "retrieval_attempts" }

func (a *RetrievalAttempt) BeforeCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() {
		util.LogIfErr(ctx, &err, "RetrievalAttempt.BeforeCreate", "id", a.ID, "docKey", a.DocKey, "investmentUuid", a.InvestmentUUID, "status", a.Status)
	}()
	if a.ID == "" {
		a.ID = NewSortableID()
	}
	if a.CreatedAt == "" {
		a.CreatedAt = CurrentTimestamp()
	}
	return nil
}

// RetrievalAttemptsTable provides append-only logging of document
// retrieval attempts.
type RetrievalAttemptsTable struct {
	db *gorm.DB
}

// NewRetrievalAttemptsTable wraps an open *gorm.DB as a
// RetrievalAttemptsTable.
func NewRetrievalAttemptsTable(db *gorm.DB) *RetrievalAttemptsTable {
	return &RetrievalAttemptsTable{db: db}
}

// Log appends an attempt to the retrieval_attempts table.  Convenience
// wrapper around Insert that takes the attempt by value.
func (t *RetrievalAttemptsTable) Log(ctx context.Context, attempt RetrievalAttempt) (err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "RetrievalAttemptsTable.Log", "docKey", attempt.DocKey, "investmentUuid", attempt.InvestmentUUID, "status", attempt.Status)
	}()
	return t.db.WithContext(ctx).Create(&attempt).Error
}

// Insert persists attempt.  Returns an error if a row with the same ID
// already exists.
func (t *RetrievalAttemptsTable) Insert(ctx context.Context, attempt *RetrievalAttempt) (err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "RetrievalAttemptsTable.Insert", "id", attempt.ID, "docKey", attempt.DocKey, "investmentUuid", attempt.InvestmentUUID, "status", attempt.Status)
	}()
	return t.db.WithContext(ctx).Create(attempt).Error
}

// Update modifies the existing row matching attempt.ID.  Returns
// ErrNotFound if no such row exists.
func (t *RetrievalAttemptsTable) Update(ctx context.Context, attempt *RetrievalAttempt) (err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "RetrievalAttemptsTable.Update", "id", attempt.ID, "docKey", attempt.DocKey, "investmentUuid", attempt.InvestmentUUID, "status", attempt.Status)
	}()
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&RetrievalAttempt{}).Where("id = ?", attempt.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrNotFound
		}
		return tx.Save(attempt).Error
	})
}

// InsertOrUpdate persists attempt, inserting a new row if attempt.ID
// does not exist or updating the existing row otherwise.  The returned
// bool is true if a new row was inserted.
func (t *RetrievalAttemptsTable) InsertOrUpdate(ctx context.Context, attempt *RetrievalAttempt) (inserted bool, err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "RetrievalAttemptsTable.InsertOrUpdate", "id", attempt.ID, "docKey", attempt.DocKey, "investmentUuid", attempt.InvestmentUUID, "status", attempt.Status)
	}()
	if attempt.ID == "" {
		attempt.ID = NewSortableID()
	}
	err = t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&RetrievalAttempt{}).Where("id = ?", attempt.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			inserted = true
			return tx.Create(attempt).Error
		}
		return tx.Save(attempt).Error
	})
	return inserted, err
}

// ListByInvestment returns all retrieval attempts tied to
// investmentUUID in newest-first order.
func (t *RetrievalAttemptsTable) ListByInvestment(ctx context.Context, investmentUUID string) (out []RetrievalAttempt, err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "RetrievalAttemptsTable.ListByInvestment", "investmentUuid", investmentUUID)
	}()
	var rows []RetrievalAttempt
	if err = t.db.WithContext(ctx).
		Where("investment_uuid = ?", investmentUUID).
		Order("id DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// GetByKey returns the most recent retrieval attempt matching
// (investmentUUID, key) where key is the caller-supplied DocKey.
// Returns ErrNotFound if no row matches.  Multiple attempts can share
// the same (investment, doc_key) — the newest is returned.
func (t *RetrievalAttemptsTable) GetByKey(ctx context.Context, investmentUUID, key string) (out RetrievalAttempt, err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "RetrievalAttemptsTable.GetByKey", "investmentUuid", investmentUUID, "docKey", key)
	}()
	var row RetrievalAttempt
	err = t.db.WithContext(ctx).
		Where("investment_uuid = ? AND doc_key = ?", investmentUUID, key).
		Order("id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return RetrievalAttempt{}, ErrNotFound
	}
	if err != nil {
		return RetrievalAttempt{}, err
	}
	return row, nil
}
