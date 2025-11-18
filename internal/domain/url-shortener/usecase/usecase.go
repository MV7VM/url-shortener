package usecase

import (
	"context"
	"strings"

	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/repository/cache"
	"go.uber.org/zap"
)

// -----------------------------------------------------------------------------
// Use-case layer (business-logic façade)
// -----------------------------------------------------------------------------

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ012345678"
)

type Usecase struct {
	log   *zap.Logger
	count uint64
	repo  repo
}

type repo interface {
	Set(ctx context.Context, key, value string) error
	Get(ctx context.Context, s string) (string, error)
}

func NewUsecase(l *zap.Logger, repo *cache.Repository) (*Usecase, error) {
	return &Usecase{log: l.Named("usecase"), repo: repo}, nil
}

func (u *Usecase) GetByID(ctx context.Context, s string) (string, error) {
	url, err := u.repo.Get(ctx, s)
	if err != nil {
		u.log.Error("failed to get url", zap.String("url", s), zap.Error(err))
		return "", err
	}

	return url, nil
}

func (u *Usecase) CreateShortURL(ctx context.Context, url string) (string, error) {
	encodedUrl := u.shortenURL()

	err := u.repo.Set(ctx, encodedUrl, url)
	if err != nil {
		u.log.Error("failed to set url", zap.String("url", url), zap.Error(err))
		return "", err
	}

	return encodedUrl, nil
}

func (u *Usecase) shortenURL() string {
	u.count++
	return base62Encode(u.count)
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
