package cache

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/MV7VM/url-shortener/internal/config"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/entities"
)

type Repository struct {
	db  *sync.Map
	cfg *config.Model
}

func NewRepository(cfg *config.Model) *Repository {
	return &Repository{
		db:  new(sync.Map),
		cfg: cfg,
	}
}

func (r *Repository) OnStart(_ context.Context) error {
	return r.recovery()
}

func (r *Repository) OnStop(_ context.Context) error {
	return r.save()
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

func (r *Repository) GetCount(ctx context.Context) (int, error) {
	count := 0

	r.db.Range(func(k, v interface{}) bool {
		count++
		return true
	})

	return count, nil
}

func (r *Repository) recovery() error {
	file, err := os.OpenFile(r.cfg.Repo.SavingFilePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.Size() == 0 {
		return nil
	}

	var items []entities.Item
	if err := json.NewDecoder(file).Decode(&items); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	for _, item := range items {
		if item.ShortURL == "" || item.OriginalURL == "" {
			continue
		}
		r.db.Store(item.ShortURL, item.OriginalURL)
	}

	return nil
}

func (r *Repository) save() error {
	file, err := os.OpenFile(r.cfg.Repo.SavingFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	items := make([]entities.Item, 0)
	r.db.Range(func(k, v any) bool {
		shortURL, ok1 := k.(string)
		originalURL, ok2 := v.(string)
		if !ok1 || !ok2 {
			return true
		}

		items = append(items, entities.Item{
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		})
		return true
	})

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(items)
}
