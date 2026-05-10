package model

import (
	"errors"
	"investment-analysis/util"

	"gorm.io/gorm"
)

// RetrievalSettings is the row for the retrieval_settings table.  The
// table always holds exactly one row (id = 1).
type RetrievalSettings struct {
	ID                       int    `gorm:"column:id;primaryKey;default:1"                         json:"id,omitempty"`
	PlaywrightTimeoutSeconds int    `gorm:"column:playwright_timeout_seconds;not null;default:300" json:"playwrightTimeoutSeconds"`
	UpdatedAt                string `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"   json:"updatedAt,omitempty"`
}

func (RetrievalSettings) TableName() string { return "retrieval_settings" }

func (r *RetrievalSettings) BeforeCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() { util.LogIfErr(ctx, &err, "RetrievalSettings.BeforeCreate", "id", r.ID) }()
	if r.UpdatedAt == "" {
		r.UpdatedAt = CurrentTimestamp()
	}
	return nil
}

func (r *RetrievalSettings) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() { util.LogIfErr(ctx, &err, "RetrievalSettings.BeforeUpdate", "id", r.ID) }()
	r.UpdatedAt = CurrentTimestamp()
	return nil
}

// DefaultRetrievalSettings returns the out-of-the-box retrieval
// configuration.
func DefaultRetrievalSettings() RetrievalSettings {
	return RetrievalSettings{
		PlaywrightTimeoutSeconds: 300,
	}
}

// RetrievalSettingsTable provides CRUD on the singleton
// retrieval_settings row.
type RetrievalSettingsTable struct {
	db *gorm.DB
}

// NewRetrievalSettingsTable wraps an open *gorm.DB as a
// RetrievalSettingsTable.
func NewRetrievalSettingsTable(db *gorm.DB) *RetrievalSettingsTable {
	return &RetrievalSettingsTable{db: db}
}

func (t *RetrievalSettingsTable) Get() (out RetrievalSettings, err error) {
	defer func() { util.LogIfErr(nil, &err, "RetrievalSettingsTable.Get") }()
	var row RetrievalSettings
	if err = t.db.First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DefaultRetrievalSettings(), nil
		}
		return RetrievalSettings{}, err
	}
	return row, nil
}

// Insert persists settings.  The singleton row's ID defaults to 1 if
// unset.  Returns an error if the singleton row already exists.
func (t *RetrievalSettingsTable) Insert(settings *RetrievalSettings) (err error) {
	defer func() {
		util.LogIfErr(nil, &err, "RetrievalSettingsTable.Insert",
			"id", settings.ID,
			"playwrightTimeoutSeconds", settings.PlaywrightTimeoutSeconds)
	}()
	if settings.ID == 0 {
		settings.ID = 1
	}
	return t.db.Create(settings).Error
}

// Update modifies the existing singleton row.  Returns ErrNotFound if
// no row exists yet.
func (t *RetrievalSettingsTable) Update(settings *RetrievalSettings) (err error) {
	defer func() {
		util.LogIfErr(nil, &err, "RetrievalSettingsTable.Update",
			"id", settings.ID,
			"playwrightTimeoutSeconds", settings.PlaywrightTimeoutSeconds)
	}()
	if settings.ID == 0 {
		settings.ID = 1
	}
	return t.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&RetrievalSettings{}).Where("id = ?", settings.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrNotFound
		}
		return tx.Save(settings).Error
	})
}

// InsertOrUpdate persists settings, inserting a new singleton row if
// none exists or updating the existing one otherwise.  The returned
// bool is true if a new row was inserted.
func (t *RetrievalSettingsTable) InsertOrUpdate(settings *RetrievalSettings) (inserted bool, err error) {
	defer func() {
		util.LogIfErr(nil, &err, "RetrievalSettingsTable.InsertOrUpdate",
			"id", settings.ID,
			"playwrightTimeoutSeconds", settings.PlaywrightTimeoutSeconds)
	}()
	if settings.ID == 0 {
		settings.ID = 1
	}
	err = t.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&RetrievalSettings{}).Where("id = ?", settings.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			inserted = true
			return tx.Create(settings).Error
		}
		return tx.Save(settings).Error
	})
	return inserted, err
}
