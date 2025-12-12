package repository

import (
	"go.uber.org/fx"
)

func New() fx.Option {
	return fx.Module("repository",
		fx.Provide(
			NewRepo,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, s *Repo) {
				lc.Append(fx.Hook{
					OnStart: s.OnStart,
					OnStop:  s.OnStop,
				})
			},
		),
	)
}
