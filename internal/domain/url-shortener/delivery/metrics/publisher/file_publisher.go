package publisher

import (
	"encoding/json"
	"os"

	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/entities"
	"go.uber.org/zap"
)

type FilePublisher struct {
	logger   *zap.Logger
	filePath string
}

func NewFilePublisher(log *zap.Logger, filePath string) *FilePublisher {
	return &FilePublisher{filePath: filePath, logger: log}
}

func (p *FilePublisher) Update(s *entities.Event) {
	file, err := os.OpenFile(p.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		p.logger.Error("Failed to open file", zap.Error(err))
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(s)
	if err != nil {
		p.logger.Error("Failed to write to file", zap.Error(err))
		return
	}
}
