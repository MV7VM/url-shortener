package repository

import (
	"context"

	"github.com/MV7VM/url-shortener/internal/config"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/repository/cache"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/repository/postgres"
)

type Repo struct {
	cache *cache.Repository
	psql  *postgres.Repository
}

func NewRepo(ctx context.Context, cfg *config.Model) (*Repo, error) {
	psql, err := postgres.NewRepository(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Repo{
		cache: cache.NewRepository(cfg),
		psql:  psql,
	}, nil
}

func (r *Repo) OnStart(ctx context.Context) error {
	err := r.cache.OnStart(ctx)
	if err != nil {
		return err
	}

	err = r.psql.OnStart(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) OnStop(ctx context.Context) error {
	err := r.cache.OnStop(ctx)
	if err != nil {
		return err
	}

	err = r.psql.OnStop(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) Set(ctx context.Context, key string, value string) error {
	return r.cache.Set(ctx, key, value)
}

func (r *Repo) Get(ctx context.Context, s string) (string, error) {
	return r.cache.Get(ctx, s)
}

func (r *Repo) GetCount(ctx context.Context) (int, error) {
	return r.cache.GetCount(ctx)
}

func (r *Repo) Ping(ctx context.Context) error {
	return r.psql.Ping(ctx)
}
