package model

// RetrievalSettings holds application-wide configuration for the document
// retrieval subsystem.
type RetrievalSettings struct {
	// PlaywrightTimeoutSeconds is the maximum number of seconds to wait for a
	// browser-driven page fetch to complete (navigation + DOM stability).
	// Default: 300 (5 minutes).
	PlaywrightTimeoutSeconds int `json:"playwrightTimeoutSeconds"`
}

// DefaultRetrievalSettings returns the out-of-the-box retrieval configuration.
func DefaultRetrievalSettings() RetrievalSettings {
	return RetrievalSettings{
		PlaywrightTimeoutSeconds: 300,
	}
}
