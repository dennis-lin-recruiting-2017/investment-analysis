package model

import (
	"context"
	"errors"
	"investment-analysis/util"
	"strings"

	"gorm.io/gorm"
)

// Inference is a single LLM inference call associated with an
// investment.  The struct is both the GORM row for the inferences
// table and the JSON-serializable representation.  It captures the
// prompt and raw response, the call metadata (token usage and timing),
// and an optional processed-artifact slot for downstream derived text
// (e.g. an extracted JSON object or summary).
type Inference struct {
	UUID              string `gorm:"column:uuid;primaryKey"                                                                        json:"uuid"`
	InvestmentUUID    string `gorm:"column:investment_uuid;not null;index;index:idx_inferences_investment_key,priority:1"          json:"investmentUuid"`
	Key               string `gorm:"column:inference_key;not null;default:'';index:idx_inferences_investment_key,priority:2"       json:"key,omitempty"`
	Model             string `gorm:"column:model;not null;default:''"                                                              json:"model"`
	Prompt            string `gorm:"column:prompt;not null;default:''"                                                             json:"prompt"`
	Response          string `gorm:"column:response;not null;default:''"                                                           json:"response"`
	ProcessedArtifact string `gorm:"column:processed_artifact;not null;default:''"                                                 json:"processedArtifact,omitempty"`
	PromptTokens      int    `gorm:"column:prompt_tokens;not null;default:0"                                                       json:"promptTokens"`
	CompletionTokens  int    `gorm:"column:completion_tokens;not null;default:0"                                                   json:"completionTokens"`
	TotalTokens       int    `gorm:"column:total_tokens;not null;default:0"                                                        json:"totalTokens"`
	RequestedAt       string `gorm:"column:requested_at;not null;default:''"                                                       json:"requestedAt,omitempty"`
	RespondedAt       string `gorm:"column:responded_at;not null;default:''"                                                       json:"respondedAt,omitempty"`
	ElapsedMs         int64  `gorm:"column:elapsed_ms;not null;default:0"                                                          json:"elapsedMs"`
	CreatedAt         string `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"                                          json:"createdAt,omitempty"`
}

func (Inference) TableName() string { return "inferences" }

func (i *Inference) BeforeCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() {
		util.LogIfErr(ctx, &err, "Inference.BeforeCreate", "uuid", i.UUID, "investmentUuid", i.InvestmentUUID, "key", i.Key)
	}()
	if i.UUID == "" {
		i.UUID = NewSortableID()
	}
	if i.CreatedAt == "" {
		i.CreatedAt = CurrentTimestamp()
	}
	return nil
}

// InferencesTable provides CRUD on the inferences table.
type InferencesTable struct {
	db *gorm.DB
}

// NewInferencesTable wraps an open *gorm.DB as an InferencesTable.
func NewInferencesTable(db *gorm.DB) *InferencesTable {
	return &InferencesTable{db: db}
}

// Save persists a new inference row.  InvestmentUUID is required; the
// row's UUID is auto-assigned if empty.
func (t *InferencesTable) Save(ctx context.Context, i Inference) (out Inference, err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "InferencesTable.Save", "investmentUuid", i.InvestmentUUID, "key", i.Key, "model", i.Model)
	}()
	if strings.TrimSpace(i.InvestmentUUID) == "" {
		return Inference{}, errors.New("persistence: Save requires InvestmentUUID")
	}
	if i.UUID == "" {
		i.UUID = NewSortableID()
	}
	if err = t.db.WithContext(ctx).Create(&i).Error; err != nil {
		return Inference{}, err
	}
	return i, nil
}

// Insert persists i.  Returns an error if a row with the same UUID
// already exists.  InvestmentUUID is required.
func (t *InferencesTable) Insert(ctx context.Context, i *Inference) (err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "InferencesTable.Insert", "uuid", i.UUID, "investmentUuid", i.InvestmentUUID, "key", i.Key)
	}()
	if strings.TrimSpace(i.InvestmentUUID) == "" {
		return errors.New("persistence: Insert requires InvestmentUUID")
	}
	return t.db.WithContext(ctx).Create(i).Error
}

// Update modifies the existing row matching i.UUID.  Returns
// ErrNotFound if no such row exists.
func (t *InferencesTable) Update(ctx context.Context, i *Inference) (err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "InferencesTable.Update", "uuid", i.UUID, "investmentUuid", i.InvestmentUUID, "key", i.Key)
	}()
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&Inference{}).Where("uuid = ?", i.UUID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrNotFound
		}
		return tx.Save(i).Error
	})
}

// InsertOrUpdate persists i, inserting a new row if i.UUID does not
// exist or updating the existing row otherwise.  The returned bool is
// true if a new row was inserted.  InvestmentUUID is required.
func (t *InferencesTable) InsertOrUpdate(ctx context.Context, i *Inference) (inserted bool, err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "InferencesTable.InsertOrUpdate", "uuid", i.UUID, "investmentUuid", i.InvestmentUUID, "key", i.Key)
	}()
	if strings.TrimSpace(i.InvestmentUUID) == "" {
		return false, errors.New("persistence: InsertOrUpdate requires InvestmentUUID")
	}
	if i.UUID == "" {
		i.UUID = NewSortableID()
	}
	err = t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&Inference{}).Where("uuid = ?", i.UUID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			inserted = true
			return tx.Create(i).Error
		}
		return tx.Save(i).Error
	})
	return inserted, err
}

func (t *InferencesTable) Get(ctx context.Context, inferenceUUID string) (out Inference, err error) {
	defer func() { util.LogIfErr(ctx, &err, "InferencesTable.Get", "uuid", inferenceUUID) }()
	var row Inference
	err = t.db.WithContext(ctx).Where("uuid = ?", inferenceUUID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Inference{}, ErrNotFound
	}
	if err != nil {
		return Inference{}, err
	}
	return row, nil
}

// ListByInvestment returns all inferences for the given investment in
// newest-first order (UUIDv7 is sortable by creation time).
func (t *InferencesTable) ListByInvestment(ctx context.Context, investmentUUID string) (out []Inference, err error) {
	defer func() { util.LogIfErr(ctx, &err, "InferencesTable.ListByInvestment", "investmentUuid", investmentUUID) }()
	var rows []Inference
	if err = t.db.WithContext(ctx).
		Where("investment_uuid = ?", investmentUUID).
		Order("uuid DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// GetByKey returns the most recent inference for (investmentUUID, key).
// Use this when callers identify inferences by a caller-supplied key
// (e.g. a prompt template name or a request fingerprint) rather than
// by the auto-assigned UUID.  Returns ErrNotFound if no row matches.
//
// Keys are not enforced unique at the schema level; if multiple
// inferences share a (investment, key) pair, the newest is returned
// (UUIDv7 is sortable by creation time).
func (t *InferencesTable) GetByKey(ctx context.Context, investmentUUID, key string) (out Inference, err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "InferencesTable.GetByKey", "investmentUuid", investmentUUID, "key", key)
	}()
	var row Inference
	err = t.db.WithContext(ctx).
		Where("investment_uuid = ? AND inference_key = ?", investmentUUID, key).
		Order("uuid DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Inference{}, ErrNotFound
	}
	if err != nil {
		return Inference{}, err
	}
	return row, nil
}

func (t *InferencesTable) UpdateArtifact(ctx context.Context, inferenceUUID, artifact string) (err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "InferencesTable.UpdateArtifact", "uuid", inferenceUUID, "artifactLen", len(artifact))
	}()
	result := t.db.WithContext(ctx).
		Model(&Inference{}).
		Where("uuid = ?", inferenceUUID).
		Update("processed_artifact", artifact)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
