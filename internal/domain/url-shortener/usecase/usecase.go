package usecase

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/entities"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/repository"
	"go.uber.org/zap"
)

// -----------------------------------------------------------------------------
// Use-case layer (business-logic façade)
// -----------------------------------------------------------------------------

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ012345678"
)

// Usecase implements the core business logic for shortening and resolving URLs.
// It hides details of the underlying repositories behind a narrow interface.
type Usecase struct {
	log   *zap.Logger
	count atomic.Uint64
	repo  repo
}

type repo interface {
	Set(ctx context.Context, key, value, userID string) (string, error)
	Get(ctx context.Context, s string) (string, bool, error)
	GetCount(ctx context.Context) (int, error)
	Ping(ctx context.Context) error
	GetUsersUrls(ctx context.Context, userID string) ([]entities.Item, error)
	Delete(ctx context.Context, shortURL []string, userID string) error
}

// NewUsecase constructs a new Usecase instance backed by the given repository.
func NewUsecase(l *zap.Logger, repo *repository.Repo) (*Usecase, error) {
	return &Usecase{log: l.Named("usecase"), repo: repo}, nil
}

// OnStart initialises internal counters from the repository and is intended
// to be used as an Fx lifecycle hook.
func (u *Usecase) OnStart(ctx context.Context) error {
	count, err := u.repo.GetCount(ctx)
	if err != nil {
		return err
	}

	u.count.Store(uint64(count))

	u.log.Info("started from", zap.Uint64("count", u.count.Load()))

	return nil
}

// GetByID resolves a short URL identifier into the original URL and a flag
// indicating whether the URL has been logically deleted.
func (u *Usecase) GetByID(ctx context.Context, s string) (string, bool, error) {
	url, isDeleted, err := u.repo.Get(ctx, s)
	if err != nil {
		u.log.Error("failed to get url", zap.String("url", s), zap.Error(err))
		return "", false, err
	}

	return url, isDeleted, nil
}

// CreateShortURL generates a short URL for the provided original URL and
// persists it for the specified user.
// It returns the stored key and a flag indicating whether it already existed.
func (u *Usecase) CreateShortURL(ctx context.Context, url, userID string) (string, bool, error) {
	encodedURL := u.shortenURL()

	shortURL, err := u.repo.Set(ctx, encodedURL, url, userID)
	if err != nil {
		u.log.Error("failed to set url", zap.String("url", url), zap.Error(err))
		return "", false, err
	}

	if encodedURL != shortURL {
		return shortURL, true, nil
	}

	return shortURL, false, nil
}

// Ping checks availability of the underlying repository.
func (u *Usecase) Ping(ctx context.Context) error {
	err := u.repo.Ping(ctx)
	if err != nil {
		u.log.Error("failed to ping repository", zap.Error(err))
		return err
	}

	return nil
}

// BatchURLs shortens multiple URLs at once and mutates the slice in-place
// so that each item contains its resulting short identifier.
func (u *Usecase) BatchURLs(ctx context.Context, urls []entities.BatchItem, userID string) error {
	for i := range urls {
		urls[i].ShortURL = u.shortenURL()

		shortURL, err := u.repo.Set(ctx, urls[i].ShortURL, urls[i].OriginalURL, userID)
		if err != nil {
			u.log.Error("failed to set url", zap.String("url", urls[i].OriginalURL), zap.Error(err))
			return err
		}

		urls[i].OriginalURL = ""
		urls[i].ShortURL = shortURL
	}

	return nil
}

// GetUsersUrls returns all URLs that were created by the given user.
func (u *Usecase) GetUsersUrls(ctx context.Context, userID string) ([]entities.Item, error) {
	urls, err := u.repo.GetUsersUrls(ctx, userID)
	if err != nil {
		u.log.Error("failed to get users urls", zap.Error(err))
		return nil, err
	}

	return urls, nil
}

// Delete marks a set of short URLs as deleted for the given user.
func (u *Usecase) Delete(ctx context.Context, shortURL []string, userID string) error {
	err := u.repo.Delete(ctx, shortURL, userID)
	if err != nil {
		u.log.Error("failed to delete urls", zap.Error(err))
		return err
	}

	return nil
}

func (u *Usecase) shortenURL() string {
	u.count.Add(1)
	return base62Encode(u.count.Load())
}

func base62Encode(number uint64) string {
	length := len(alphabet)
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(10)
	for ; number > 0; number = number / uint64(length) {
		encodedBuilder.WriteByte(alphabet[(number % uint64(length))])
	}

	return encodedBuilder.String()
}
