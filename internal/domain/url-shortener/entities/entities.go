package entities

type CtxKeyString string

type Item struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchItem struct {
	CorrelationId string `json:"correlation_id"`
	OriginalUrl   string `json:"original_url,omitempty"`
	ShortUrl      string `json:"short_url"`
}
