package postgres

// Package postgres implements the pgx-based data-access layer.
// It is Fx-compatible (provides lifecycle hooks) and reflects Forest Fairy
// «UUID edition» schema after the April-2025 migration.

import (
	"context"

	"github.com/MV7VM/url-shortener/internal/config"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/entities"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const txKey entities.CtxKeyString = "tx"

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
func (r *Repository) OnStart(ctx context.Context) (err error) {
	r.db, err = pgxpool.New(r.ctx, r.cfg.PsqlConnString)
	if err != nil {
		return err
	}

	if err = r.withTx(ctx, func(ctxTx context.Context) error {
		tx := ctxTx.Value(txKey).(pgx.Tx)
		return r.migrate(ctx, tx)
	}); err != nil {
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

const qSet = `
INSERT INTO 
    shortener.urls (short_url, url) 
VALUES 
    ($1, $2)`

func (r *Repository) Set(ctx context.Context, key string, value string) error {
	if _, err := r.db.Exec(ctx, qSet, key, value); err != nil {
		return err
	}

	return nil
}

const qGet = `
select 
    url 
from 
    shortener.urls 
where 
    short_url = $1`

func (r *Repository) Get(ctx context.Context, s string) (url string, err error) {
	err = r.db.QueryRow(ctx, qGet, s).Scan(&url)
	if err != nil {
		return "", err
	}

	return url, nil
}

const qGetCount = `
select 
    count(*) 
from 
    shortener.urls`

func (r *Repository) GetCount(ctx context.Context) (count int, err error) {
	err = r.db.QueryRow(ctx, qGetCount).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// migrate создает схему и таблицу для хранения URL, если они не существуют.
// Если tx == nil, операции выполняются напрямую через пул соединений.
func (r *Repository) migrate(ctx context.Context, tx pgx.Tx) error {
	var execFunc func(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	if tx != nil {
		execFunc = tx.Exec
	} else {
		execFunc = r.db.Exec
	}

	// Создаем схему shortener, если её нет
	_, err := execFunc(ctx, `CREATE SCHEMA IF NOT EXISTS shortener`)
	if err != nil {
		return err
	}

	// Создаем таблицу urls, если её нет
	_, err = execFunc(ctx, `
		CREATE TABLE IF NOT EXISTS shortener.urls (
			short_url TEXT PRIMARY KEY,
			url TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) withTx(ctx context.Context, f func(context.Context) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ctxTx := context.WithValue(ctx, txKey, tx)

	err = f(ctxTx)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
