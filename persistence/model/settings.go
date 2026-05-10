package model

import (
	"errors"
	"investment-analysis/util"

	"gorm.io/gorm"
)

// LLMSettings is the row for the llm_settings table; the same struct
// serializes directly as JSON via the json tags on each field.
type LLMSettings struct {
	Provider     string  `gorm:"column:provider;primaryKey"               json:"provider"`
	Endpoint     string  `gorm:"column:endpoint;not null"                 json:"endpoint"`
	Model        string  `gorm:"column:model;not null"                    json:"model"`
	APIKey       string  `gorm:"column:api_key;not null;default:''"       json:"apiKey,omitempty"`
	Temperature  float64 `gorm:"column:temperature;not null;default:0.7"  json:"temperature"`
	SystemPrompt string  `gorm:"column:system_prompt;not null;default:''" json:"systemPrompt"`
	UpdatedAt    string  `gorm:"column:updated_at;not null"               json:"updatedAt,omitempty"`
}

func (LLMSettings) TableName() string { return "llm_settings" }

func (s *LLMSettings) BeforeCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() { util.LogIfErr(ctx, &err, "LLMSettings.BeforeCreate", "provider", s.Provider) }()
	s.UpdatedAt = CurrentTimestamp()
	return nil
}

func (s *LLMSettings) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() { util.LogIfErr(ctx, &err, "LLMSettings.BeforeUpdate", "provider", s.Provider) }()
	s.UpdatedAt = CurrentTimestamp()
	return nil
}

// DefaultSettings returns the out-of-the-box configuration for a given
// provider.  Provider may be "lm-studio", "ollama", or
// "external-endpoint"; an unknown provider falls back to "lm-studio".
func DefaultSettings(provider string) LLMSettings {
	switch provider {
	case "ollama":
		return LLMSettings{
			Provider:     "ollama",
			Endpoint:     "http://localhost:11434/api/generate",
			Model:        "llama3.2",
			Temperature:  0.7,
			SystemPrompt: "You are a helpful assistant.",
		}
	case "external-endpoint":
		return LLMSettings{
			Provider:     "external-endpoint",
			Endpoint:     "https://api.example.com/v1/chat/completions",
			Model:        "gpt-4.1-mini",
			Temperature:  0.7,
			SystemPrompt: "You are a helpful assistant.",
		}
	default:
		return LLMSettings{
			Provider:     "lm-studio",
			Endpoint:     "http://localhost:1234/v1/chat/completions",
			Model:        "local-model",
			Temperature:  0.7,
			SystemPrompt: "You are a helpful assistant.",
		}
	}
}

// Providers returns the list of LLM providers known to this build.
func Providers() []string {
	return []string{"lm-studio", "ollama", "external-endpoint"}
}

// LLMSettingsTable provides CRUD on the llm_settings table.
type LLMSettingsTable struct {
	db *gorm.DB
}

// NewLLMSettingsTable wraps an open *gorm.DB as an LLMSettingsTable.
func NewLLMSettingsTable(db *gorm.DB) *LLMSettingsTable {
	return &LLMSettingsTable{db: db}
}

// Get returns the settings for provider, or DefaultSettings(provider)
// if no row exists yet.
func (t *LLMSettingsTable) Get(provider string) (out LLMSettings, err error) {
	defer func() { util.LogIfErr(nil, &err, "LLMSettingsTable.Get", "provider", provider) }()
	var row LLMSettings
	if err = t.db.Where("provider = ?", provider).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DefaultSettings(provider), nil
		}
		return LLMSettings{}, err
	}
	return row, nil
}

// List returns settings for every provider known to Providers().
// Missing rows are returned as their default settings.
func (t *LLMSettingsTable) List() (out []LLMSettings, err error) {
	defer func() { util.LogIfErr(nil, &err, "LLMSettingsTable.List") }()
	providers := Providers()
	out = make([]LLMSettings, 0, len(providers))
	for _, provider := range providers {
		s, e := t.Get(provider)
		if e != nil {
			return nil, e
		}
		out = append(out, s)
	}
	return out, nil
}

// Insert persists s.  Returns an error if a row with the same Provider
// already exists.
func (t *LLMSettingsTable) Insert(s *LLMSettings) (err error) {
	defer func() { util.LogIfErr(nil, &err, "LLMSettingsTable.Insert", "provider", s.Provider) }()
	return t.db.Create(s).Error
}

// Update modifies the existing row matching s.Provider.  Returns
// ErrNotFound if no such row exists.
func (t *LLMSettingsTable) Update(s *LLMSettings) (err error) {
	defer func() { util.LogIfErr(nil, &err, "LLMSettingsTable.Update", "provider", s.Provider) }()
	return t.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&LLMSettings{}).Where("provider = ?", s.Provider).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrNotFound
		}
		return tx.Save(s).Error
	})
}

// InsertOrUpdate persists s, inserting a new row if Provider does not
// exist or updating the existing row otherwise.  The returned bool is
// true if a new row was inserted, false if an existing row was updated.
func (t *LLMSettingsTable) InsertOrUpdate(s *LLMSettings) (inserted bool, err error) {
	defer func() { util.LogIfErr(nil, &err, "LLMSettingsTable.InsertOrUpdate", "provider", s.Provider) }()
	err = t.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&LLMSettings{}).Where("provider = ?", s.Provider).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			inserted = true
			return tx.Create(s).Error
		}
		return tx.Save(s).Error
	})
	return inserted, err
}

func (t *LLMSettingsTable) Delete(provider string) (err error) {
	defer func() { util.LogIfErr(nil, &err, "LLMSettingsTable.Delete", "provider", provider) }()
	return t.db.Delete(&LLMSettings{}, "provider = ?", provider).Error
}
