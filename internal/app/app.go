package app

import (
	"context"

	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/http"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/repository"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/usecase"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func New() *fx.App {
	return fx.New(
		fx.Options(
			repository.New(), //
			usecase.New(),
			http.New(),
		),
		fx.Provide(
			context.Background,
			zap.NewDevelopment,
		),
		fx.WithLogger(
			func(log *zap.Logger) fxevent.Logger {
				return &fxevent.ZapLogger{Logger: log}
			},
		),
	)
}
