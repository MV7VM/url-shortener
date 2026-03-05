package publisher

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/entities"
	"go.uber.org/zap"
)

// URLPublisher sends audit events as JSON over HTTP to a configured endpoint.
type URLPublisher struct {
	url    string
	logger *zap.Logger
}

// NewURLPublisher constructs a URLPublisher that posts events to the given URL.
func NewURLPublisher(log *zap.Logger, url string) *URLPublisher {
	return &URLPublisher{url: url, logger: log}
}

// Update marshals the event to JSON and performs an HTTP POST to the endpoint.
func (p *URLPublisher) Update(s *entities.Event) {
	eventJSON, err := json.Marshal(s)
	if err != nil {
		p.logger.Error("failed to marshal event", zap.Error(err))
		return
	}

	req, err := http.NewRequest(http.MethodPost, p.url, bytes.NewBuffer(eventJSON))
	if err != nil {
		p.logger.Error("failed to create request", zap.Error(err))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		p.logger.Error("failed to send request", zap.Error(err))
		return
	}
	defer resp.Body.Close()
}
