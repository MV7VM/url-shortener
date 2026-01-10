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
    shortener.urls (short_url, url, user_id) 
VALUES 
    ($1, $2, $3) 
ON CONFLICT (url) DO UPDATE 
    SET short_url = shortener.urls.short_url 
RETURNING short_url
`

func (r *Repository) Set(ctx context.Context, key, value, userID string) (string, error) {
	var storedKey string
	if err := r.db.QueryRow(ctx, qSet, key, value, userID).Scan(&storedKey); err != nil {
		return "", err
	}

	return storedKey, nil
}

const qGet = `
select 
    url, is_deleted 
from 
    shortener.urls 
where 
    short_url = $1`

func (r *Repository) Get(ctx context.Context, s string) (url string, isDelete bool, err error) {
	err = r.db.QueryRow(ctx, qGet, s).Scan(&url, &isDelete)
	if err != nil {
		return "", false, err
	}

	return url, isDelete, nil
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

const qGetUsersUrls = `
select 
    short_url, url
from 
    shortener.urls 
where 
    user_id = $1`

func (r *Repository) GetUsersUrls(ctx context.Context, userID string) ([]entities.Item, error) {
	rows, err := r.db.Query(ctx, qGetUsersUrls, userID)
	if err != nil {
		return nil, err
	}

	urls := make([]entities.Item, 0, 8)

	for rows.Next() {
		url := entities.Item{}
		err = rows.Scan(&url.ShortURL, &url.OriginalURL)
		if err != nil {
			return nil, err
		}

		urls = append(urls, url)
	}

	return urls, nil
}

const qDelete = `
update 
    shortener.urls 
set 
    is_deleted = case 
        when user_id =$2
            then true 
            else is_deleted 
        end
where short_url = any($1);
`

func (r *Repository) Delete(ctx context.Context, shortURL []string, userID string) error {
	_, err := r.db.Exec(ctx, qDelete, shortURL, userID)
	if err != nil {
		return err
	}

	return nil
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
			url TEXT NOT NULL unique,
			user_id TEXT,
			is_deleted bool default false
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
