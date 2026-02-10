package publisher

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/entities"
	"go.uber.org/zap"
)

type URLPublisher struct {
	url    string
	logger *zap.Logger
}

func NewURLPublisher(log *zap.Logger, url string) *URLPublisher {
	return &URLPublisher{url: url, logger: log}
}

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
