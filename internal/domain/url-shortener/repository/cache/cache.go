package cache

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
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

func (r *Repository) recovery() error {
	file, err := os.OpenFile(r.cfg.Repo.SavingFilePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var item entities.Item

		err = json.Unmarshal(scanner.Bytes(), &item)
		if err != nil {
			return err
		}

		r.db.Store(item.ShortUrl, item.OriginalUrl)
	}

	return nil
}

func (r *Repository) save() error {
	file, err := os.OpenFile(r.cfg.Repo.SavingFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	r.db.Range(func(k, v any) bool {
		item, err := json.Marshal(entities.Item{
			ShortUrl:    k.(string),
			OriginalUrl: v.(string),
		})
		if err != nil {
			return false
		}

		item = append(item, []byte{',', '\n'}...)

		_, err = writer.Write(item)
		if err != nil {
			return false
		}

		return true
	}) //todo

	return nil
}
