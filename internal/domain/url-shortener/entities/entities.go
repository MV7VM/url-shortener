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
	Ts     int    `json:"ts"`
	Action string `json:"action"`
	UserId string `json:"user_id"`
	Url    string `json:"url"`
}
