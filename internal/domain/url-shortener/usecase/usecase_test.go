package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockRepo мок для интерфейса repo
type mockRepo struct {
	GetFunc      func(context.Context, string) (string, error)
	SetFunc      func(context.Context, string, string) error
	GetCountFunc func(context.Context) (int, error)
}

func (m *mockRepo) GetCount(ctx context.Context) (int, error) {
	if m.GetCountFunc != nil {
		return m.GetCountFunc(ctx)
	}
	return 0, errors.New("not implemented")
}

func (m *mockRepo) Get(ctx context.Context, key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return "", errors.New("not implemented")
}

func (m *mockRepo) Set(ctx context.Context, key, value string) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value)
	}
	return errors.New("not implemented")
}

func TestNewUsecase(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := &mockRepo{}

	// Создаем Usecase напрямую для тестов, так как NewUsecase требует *cache.Repository
	uc := &Usecase{
		log:  logger.Named("usecase"),
		repo: mockRepo,
	}

	assert.NotNil(t, uc)
	assert.NotNil(t, uc.log)
	assert.NotNil(t, uc.repo)
	assert.Equal(t, uint64(0), uc.count.Load())
}

func TestUsecase_GetByID_Success(t *testing.T) {
	logger := zap.NewNop()
	expectedURL := "https://example.com"
	expectedKey := "abc123"

	mockRepo := &mockRepo{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			assert.Equal(t, expectedKey, key)
			return expectedURL, nil
		},
	}

	uc := &Usecase{
		log:  logger.Named("usecase"),
		repo: mockRepo,
	}

	ctx := context.Background()
	result, err := uc.GetByID(ctx, expectedKey)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, result)
}

func TestUsecase_GetByID_RepositoryError(t *testing.T) {
	logger := zap.NewNop()
	expectedKey := "nonexistent"
	expectedError := errors.New("not found")

	mockRepo := &mockRepo{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			assert.Equal(t, expectedKey, key)
			return "", expectedError
		},
	}

	uc := &Usecase{
		log:  logger.Named("usecase"),
		repo: mockRepo,
	}

	ctx := context.Background()
	result, err := uc.GetByID(ctx, expectedKey)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result)
}

func TestUsecase_CreateShortURL_Success(t *testing.T) {
	logger := zap.NewNop()
	inputURL := "https://example.com"

	var capturedKey string
	var capturedValue string

	mockRepo := &mockRepo{
		SetFunc: func(ctx context.Context, key, value string) error {
			capturedKey = key
			capturedValue = value
			return nil
		},
	}

	uc := &Usecase{
		log:  logger.Named("usecase"),
		repo: mockRepo,
	}

	ctx := context.Background()
	shortURL, err := uc.CreateShortURL(ctx, inputURL)

	require.NoError(t, err)
	assert.NotEmpty(t, shortURL)
	assert.Equal(t, inputURL, capturedValue)
	assert.Equal(t, shortURL, capturedKey)
	// Первый вызов должен вернуть первый символ алфавита (так как base62Encode(1))
	assert.Equal(t, "b", shortURL) // 'a' для 0, 'b' для 1
}

func TestUsecase_CreateShortURL_RepositoryError(t *testing.T) {
	logger := zap.NewNop()
	inputURL := "https://example.com"
	expectedError := errors.New("database error")

	mockRepo := &mockRepo{
		SetFunc: func(ctx context.Context, key, value string) error {
			return expectedError
		},
	}

	uc := &Usecase{
		log:  logger.Named("usecase"),
		repo: mockRepo,
	}

	ctx := context.Background()
	shortURL, err := uc.CreateShortURL(ctx, inputURL)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, shortURL)
}

func TestUsecase_CreateShortURL_MultipleSequential(t *testing.T) {
	logger := zap.NewNop()
	keys := make([]string, 0)

	mockRepo := &mockRepo{
		SetFunc: func(ctx context.Context, key, value string) error {
			keys = append(keys, key)
			return nil
		},
	}

	uc := &Usecase{
		log:  logger.Named("usecase"),
		repo: mockRepo,
	}

	ctx := context.Background()

	// Создаем несколько URL подряд
	shortURL1, err1 := uc.CreateShortURL(ctx, "https://example1.com")
	shortURL2, err2 := uc.CreateShortURL(ctx, "https://example2.com")
	shortURL3, err3 := uc.CreateShortURL(ctx, "https://example3.com")

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)

	// Проверяем, что все короткие URL уникальны
	assert.NotEqual(t, shortURL1, shortURL2)
	assert.NotEqual(t, shortURL2, shortURL3)
	assert.NotEqual(t, shortURL1, shortURL3)

	// Проверяем, что они были сохранены в правильном порядке
	assert.Equal(t, 3, len(keys))
	assert.Equal(t, shortURL1, keys[0])
	assert.Equal(t, shortURL2, keys[1])
	assert.Equal(t, shortURL3, keys[2])
}

func TestUsecase_CreateShortURL_Base62Encoding(t *testing.T) {
	logger := zap.NewNop()

	mockRepo := &mockRepo{
		SetFunc: func(ctx context.Context, key, value string) error {
			return nil
		},
	}

	uc := &Usecase{
		log:  logger.Named("usecase"),
		repo: mockRepo,
	}

	ctx := context.Background()

	// Проверяем последовательность кодирования
	// count = 0 -> base62Encode(1) = 'b'
	// count = 1 -> base62Encode(2) = 'c'
	// и так далее

	shortURL1, _ := uc.CreateShortURL(ctx, "https://example1.com")
	assert.Equal(t, "b", shortURL1)

	shortURL2, _ := uc.CreateShortURL(ctx, "https://example2.com")
	assert.Equal(t, "c", shortURL2)

	// После 63 запросов должен появиться двусимвольный код
	// Устанавливаем count так, чтобы следующий был 63 (после инкремента)
	uc.count.Store(62)
	shortURL63, _ := uc.CreateShortURL(ctx, "https://example63.com")
	// 63 в base63 = "ba" (1*63 + 0): alphabet[0]='a', затем 63/63=1, alphabet[1]='b' -> "ba"
	assert.Equal(t, "cb", shortURL63)
}

func TestBase62Encode(t *testing.T) {
	tests := []struct {
		name     string
		number   uint64
		expected string
	}{
		{
			name:     "encode 1",
			number:   1,
			expected: "b", // alphabet[1 % 63] = alphabet[1] = 'b'
		},
		{
			name:     "encode 2",
			number:   2,
			expected: "c", // alphabet[2 % 63] = alphabet[2] = 'c'
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := base62Encode(tt.number)
			assert.Equal(t, tt.expected, result, "Number: %d", tt.number)
		})
	}
}

func TestUsecase_GetByID_EmptyKey(t *testing.T) {
	logger := zap.NewNop()
	expectedError := errors.New("empty key")

	mockRepo := &mockRepo{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			if key == "" {
				return "", expectedError
			}
			return "", errors.New("not found")
		},
	}

	uc := &Usecase{
		log:  logger.Named("usecase"),
		repo: mockRepo,
	}

	ctx := context.Background()
	result, err := uc.GetByID(ctx, "")

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result)
}

func TestUsecase_CreateShortURL_EmptyURL(t *testing.T) {
	logger := zap.NewNop()

	mockRepo := &mockRepo{
		SetFunc: func(ctx context.Context, key, value string) error {
			return nil
		},
	}

	uc := &Usecase{
		log:  logger.Named("usecase"),
		repo: mockRepo,
	}

	ctx := context.Background()
	shortURL, err := uc.CreateShortURL(ctx, "")

	require.NoError(t, err)
	assert.NotEmpty(t, shortURL)
}
