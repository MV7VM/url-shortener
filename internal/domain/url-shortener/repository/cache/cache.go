package cache

import (
	"context"
	"errors"
	"sync"
)

type Repository struct {
	db *sync.Map
}

func NewRepository() *Repository {
	return &Repository{
		db: new(sync.Map),
	}
}

func (r *Repository) Set(ctx context.Context, key, value string) error {
	r.db.Store(key, value)
	return nil
}

func (r *Repository) Get(ctx context.Context, s string) (string, error) {
	url, ok := r.db.Load(s)
	if _, okString := url.(string); !okString || !ok || url == nil {
		return "", errors.New("not found")
	}

	return url.(string), nil
}
