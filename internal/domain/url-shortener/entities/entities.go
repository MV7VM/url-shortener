package entities

type CtxKeyString string

const (
	ActionShort  = "shorten"
	ActionFollow = "follow"
)

type Item struct {
	UUID        string `json:"uuid,omitempty"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url"`
}

type Event struct {
	TS     int    `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}
