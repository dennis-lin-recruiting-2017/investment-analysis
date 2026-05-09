package model

// StoredDocument represents a retrieved financial document persisted in the database.
type StoredDocument struct {
	Key          string
	DocumentType DocumentType
	Ticker       string
	FiscalYear   int
	FiscalQtr    int
	Form         string
	SourceURL    string
	OutputLabel  string
	MimeType     string
	Body         []byte
}

// RetrievalAttempt records the outcome of a single document retrieval attempt.
type RetrievalAttempt struct {
	DocKey        string
	DocumentType  DocumentType
	Ticker        string
	FiscalYear    int
	FiscalQuarter int
	Status        string
	Message       string
}
