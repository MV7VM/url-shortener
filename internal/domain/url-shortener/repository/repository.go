package repository

import (
	"context"

	"github.com/MV7VM/url-shortener/internal/config"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/entities"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/repository/cache"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/repository/postgres"
)

type repository interface {
	Set(ctx context.Context, key string, value, userID string) (string, error)
	Get(ctx context.Context, s string) (string, bool, error)
	GetCount(ctx context.Context) (int, error)
	GetUsersUrls(ctx context.Context, userID string) ([]entities.Item, error)
	OnStart(_ context.Context) error
	OnStop(_ context.Context) error
}

// Repo is a façade over a concrete repository implementation (cache or Postgres)
// that also keeps a direct reference to the PostgreSQL repository for features
// that are only available in the DB-backed storage.
type Repo struct {
	repository
	psql *postgres.Repository
}

// NewRepo selects an appropriate repository implementation based on config
// and returns a composite Repo instance that satisfies the use-case needs.
func NewRepo(ctx context.Context, cfg *config.Model) (*Repo, error) {
	psql, err := postgres.NewRepository(ctx, cfg)
	if err != nil {
		return nil, err
	}

	var repo repository
	if cfg.Repo.PsqlConfig.PsqlConnString == "" {
		repo = cache.NewRepository(cfg)
	} else {
		repo = psql
	}

	return &Repo{
		repository: repo,
		psql:       psql,
	}, nil
}

func (r *Repo) OnStart(ctx context.Context) error {
	err := r.repository.OnStart(ctx)
	if err != nil {
		return err
	}

	if r.repository != r.psql {
		r.psql.OnStart(ctx)
	}

	return nil
}

func (r *Repo) OnStop(ctx context.Context) error {
	err := r.repository.OnStop(ctx)
	if err != nil {
		return err
	}

	if r.repository != r.psql {
		err = r.psql.OnStop(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repo) Set(ctx context.Context, key string, value, userID string) (string, error) {
	return r.repository.Set(ctx, key, value, userID)
}

func (r *Repo) Get(ctx context.Context, s string) (string, bool, error) {
	return r.repository.Get(ctx, s)
}

func (r *Repo) GetCount(ctx context.Context) (int, error) {
	return r.repository.GetCount(ctx)
}

func (r *Repo) Ping(ctx context.Context) error {
	return r.psql.Ping(ctx)
}

func (r *Repo) GetUsersUrls(ctx context.Context, userID string) ([]entities.Item, error) {
	return r.repository.GetUsersUrls(ctx, userID)
}

func (r *Repo) Delete(ctx context.Context, shortURL []string, userID string) error {
	return r.psql.Delete(ctx, shortURL, userID)
}
