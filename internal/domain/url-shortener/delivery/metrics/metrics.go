package metrics

import (
	"errors"

	"github.com/MV7VM/url-shortener/internal/config"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/metrics/publisher"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/metrics/watcher"
	"go.uber.org/zap"
)

func NewMetricWrapper(log *zap.Logger, cfg *config.Model) (*watcher.Watcher, error) {
	if cfg == nil {
		log.Warn("config is nil")
		return nil, errors.New("config is nil")
	}

	w := watcher.NewWatcher()

	if cfg.Audit.AuditURL != "" {
		w.Register(publisher.NewURLPublisher(log.Named("url-publisher"), cfg.Audit.AuditURL))
	}
	if cfg.Audit.AuditFilePath != "" {
		w.Register(publisher.NewFilePublisher(log.Named("file-publisher"), cfg.Audit.AuditFilePath))
	}

	return w, nil
}
