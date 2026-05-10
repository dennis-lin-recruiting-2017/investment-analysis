package model

import (
	"context"
	"errors"
	"investment-analysis/util"

	"gorm.io/gorm"
)

// Document is a single retrieved financial document, stored in the
// documents table and serializable directly as JSON.  InvestmentUUID
// scopes the document to an investment (empty for unscoped/global
// documents); DocKey is the caller-supplied identifier and is globally
// unique.
type Document struct {
	ID             string       `gorm:"column:id;primaryKey"                                                                          json:"id"`
	InvestmentUUID string       `gorm:"column:investment_uuid;not null;default:'';index;index:idx_documents_investment_key,priority:1" json:"investmentUuid,omitempty"`
	DocKey         string       `gorm:"column:doc_key;not null;uniqueIndex;default:'';index:idx_documents_investment_key,priority:2"   json:"docKey"`
	DocumentType   DocumentType `gorm:"column:document_type;not null;default:''"                                                      json:"documentType"`
	Ticker         string       `gorm:"column:ticker;not null;default:''"                                                             json:"ticker,omitempty"`
	FiscalYear     int          `gorm:"column:fiscal_year;not null;default:0"                                                         json:"fiscalYear,omitempty"`
	FiscalQuarter  int          `gorm:"column:fiscal_quarter;not null;default:0"                                                      json:"fiscalQuarter,omitempty"`
	Form           string       `gorm:"column:form;not null;default:''"                                                               json:"form,omitempty"`
	SourceURL      string       `gorm:"column:source_url;not null;default:''"                                                         json:"sourceUrl,omitempty"`
	OutputLabel    string       `gorm:"column:output_label;not null;default:''"                                                       json:"outputLabel,omitempty"`
	MimeType       string       `gorm:"column:mime_type;not null;default:''"                                                          json:"mimeType,omitempty"`
	Body           []byte       `gorm:"column:body;not null"                                                                          json:"body,omitempty"`
	CreatedAt      string       `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"                                          json:"createdAt,omitempty"`
}

func (Document) TableName() string { return "documents" }

func (d *Document) BeforeCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	defer func() {
		util.LogIfErr(ctx, &err, "Document.BeforeCreate", "id", d.ID, "docKey", d.DocKey, "investmentUuid", d.InvestmentUUID)
	}()
	if d.ID == "" {
		d.ID = NewSortableID()
	}
	if d.CreatedAt == "" {
		d.CreatedAt = CurrentTimestamp()
	}
	return nil
}

// DocumentsTable provides CRUD on the documents table.
type DocumentsTable struct {
	db *gorm.DB
}

// NewDocumentsTable wraps an open *gorm.DB as a DocumentsTable.
func NewDocumentsTable(db *gorm.DB) *DocumentsTable {
	return &DocumentsTable{db: db}
}

// Exists reports whether a document with the given doc_key has already
// been stored.
func (t *DocumentsTable) Exists(ctx context.Context, key string) (exists bool, err error) {
	defer func() { util.LogIfErr(ctx, &err, "DocumentsTable.Exists", "docKey", key) }()
	var count int64
	err = t.db.WithContext(ctx).Model(&Document{}).Where("doc_key = ?", key).Count(&count).Error
	return count > 0, err
}

// Insert stores a new document row.  doc.DocKey must be set.  Returns
// an error if a row with the same primary key (or doc_key uniqueness
// constraint) already exists.
func (t *DocumentsTable) Insert(ctx context.Context, doc *Document) (err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "DocumentsTable.Insert", "id", doc.ID, "docKey", doc.DocKey, "investmentUuid", doc.InvestmentUUID)
	}()
	return t.db.WithContext(ctx).Create(doc).Error
}

// Update modifies the existing row matching doc.ID.  Returns
// ErrNotFound if no such row exists.
func (t *DocumentsTable) Update(ctx context.Context, doc *Document) (err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "DocumentsTable.Update", "id", doc.ID, "docKey", doc.DocKey, "investmentUuid", doc.InvestmentUUID)
	}()
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&Document{}).Where("id = ?", doc.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrNotFound
		}
		return tx.Save(doc).Error
	})
}

// InsertOrUpdate persists doc, inserting a new row if doc.ID does not
// exist or updating the existing row otherwise.  The returned bool is
// true if a new row was inserted.
func (t *DocumentsTable) InsertOrUpdate(ctx context.Context, doc *Document) (inserted bool, err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "DocumentsTable.InsertOrUpdate", "id", doc.ID, "docKey", doc.DocKey, "investmentUuid", doc.InvestmentUUID)
	}()
	if doc.ID == "" {
		doc.ID = NewSortableID()
	}
	err = t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&Document{}).Where("id = ?", doc.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			inserted = true
			return tx.Create(doc).Error
		}
		return tx.Save(doc).Error
	})
	return inserted, err
}

// GetBody returns the raw bytes of the document with the given doc_key,
// or ErrNotFound.
func (t *DocumentsTable) GetBody(ctx context.Context, key string) (body []byte, err error) {
	defer func() { util.LogIfErr(ctx, &err, "DocumentsTable.GetBody", "docKey", key) }()
	var row Document
	err = t.db.WithContext(ctx).Where("doc_key = ?", key).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return row.Body, nil
}

// ListByInvestment returns all documents tied to investmentUUID in
// newest-first order (UUIDv7 IDs are sortable by creation time).
func (t *DocumentsTable) ListByInvestment(ctx context.Context, investmentUUID string) (out []Document, err error) {
	defer func() { util.LogIfErr(ctx, &err, "DocumentsTable.ListByInvestment", "investmentUuid", investmentUUID) }()
	var rows []Document
	if err = t.db.WithContext(ctx).
		Where("investment_uuid = ?", investmentUUID).
		Order("id DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// GetByKey returns the document matching (investmentUUID, key) where
// key is the caller-supplied DocKey.  Returns ErrNotFound if no row
// matches.  doc_key is globally unique, so this lookup also rejects
// documents whose investment_uuid doesn't match.
func (t *DocumentsTable) GetByKey(ctx context.Context, investmentUUID, key string) (out Document, err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "DocumentsTable.GetByKey", "investmentUuid", investmentUUID, "docKey", key)
	}()
	var row Document
	err = t.db.WithContext(ctx).
		Where("investment_uuid = ? AND doc_key = ?", investmentUUID, key).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Document{}, ErrNotFound
	}
	if err != nil {
		return Document{}, err
	}
	return row, nil
}
