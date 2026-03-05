package entities

// CtxKeyString is a helper type for context keys used inside the project.
type CtxKeyString string

const (
	// ActionShort is an audit action name used when a short URL is created.
	ActionShort = "shorten"
	// ActionFollow is an audit action name used when a short URL is followed.
	ActionFollow = "follow"
)

// Item represents a single shortened URL owned by a user.
type Item struct {
	UUID        string `json:"uuid,omitempty"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// BatchItem represents one element of a batch-shortening request/response.
type BatchItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url"`
}

// Event describes a high-level user action that can be sent to auditors.
type Event struct {
	TS     int    `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}
