package repository

import (
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/repository/cache"
	"go.uber.org/fx"
)

func New() fx.Option {
	return fx.Module("repository",
		fx.Provide(
			cache.NewRepository,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, s *cache.Repository) {
				lc.Append(fx.Hook{
					OnStart: s.OnStart,
					OnStop:  s.OnStop,
				})
			},
		),
	)
}
