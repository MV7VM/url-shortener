package postgres

// Package postgres implements the pgx-based data-access layer.
// It is Fx-compatible (provides lifecycle hooks) and reflects Forest Fairy
// «UUID edition» schema after the April-2025 migration.

import (
	"context"

	"github.com/MV7VM/url-shortener/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// -----------------------------------------------------------------------------
// Pg-repository (Fx-ready)
// -----------------------------------------------------------------------------

type Repository struct {
	ctx context.Context
	cfg *config.PsqlConfig
	db  *pgxpool.Pool
}

// NewRepository returns a Repo instance ready to be plugged into an Fx graph.
func NewRepository(ctx context.Context, cfg *config.Model) (*Repository, error) {
	return &Repository{ctx: ctx, cfg: &cfg.Repo.PsqlConfig}, nil
}

// OnStart — Fx Lifecycle hook: opens a pgx connection-pool (with retries).
func (r *Repository) OnStart(_ context.Context) (err error) {
	r.db, err = pgxpool.New(r.ctx, r.cfg.PsqlConnString)
	if err != nil {
		return err
	}

	return nil
}

// OnStop — Fx hook: closes pool.
func (r *Repository) OnStop(_ context.Context) error {
	if r.db != nil {
		r.db.Close()
	}
	return nil
}

func (r *Repository) Ping(ctx context.Context) error {
	err := r.db.Ping(ctx)
	if err != nil {
		return err
	}

	return nil
}
