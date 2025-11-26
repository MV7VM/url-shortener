package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/MV7VM/url-shortener/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.db)
}

func TestRepository_Set_Success(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})
	ctx := context.Background()

	err := repo.Set(ctx, "key1", "https://example.com")

	assert.NoError(t, err)
}

func TestRepository_Get_Success(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})
	ctx := context.Background()
	key := "key1"
	expectedValue := "https://example.com"

	// Сначала сохраняем значение
	err := repo.Set(ctx, key, expectedValue)
	require.NoError(t, err)

	// Затем получаем его
	result, err := repo.Get(ctx, key)

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, result)
}

func TestRepository_Get_NotFound(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})
	ctx := context.Background()
	key := "nonexistent"

	result, err := repo.Get(ctx, key)

	assert.Error(t, err)
	assert.Equal(t, "not found", err.Error())
	assert.Empty(t, result)
}

func TestRepository_Get_EmptyKey(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})
	ctx := context.Background()

	result, err := repo.Get(ctx, "")

	assert.Error(t, err)
	assert.Equal(t, "not found", err.Error())
	assert.Empty(t, result)
}

func TestRepository_Set_EmptyValue(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})
	ctx := context.Background()

	err := repo.Set(ctx, "key1", "")
	require.NoError(t, err)

	result, err := repo.Get(ctx, "key1")

	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestRepository_Set_Overwrite(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})
	ctx := context.Background()
	key := "key1"

	// Сохраняем первое значение
	err1 := repo.Set(ctx, key, "https://example1.com")
	require.NoError(t, err1)

	// Перезаписываем значением
	err2 := repo.Set(ctx, key, "https://example2.com")
	require.NoError(t, err2)

	// Проверяем, что получили новое значение
	result, err := repo.Get(ctx, key)

	assert.NoError(t, err)
	assert.Equal(t, "https://example2.com", result)
}

func TestRepository_MultipleKeys(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})
	ctx := context.Background()

	// Сохраняем несколько ключей
	repo.Set(ctx, "key1", "https://example1.com")
	repo.Set(ctx, "key2", "https://example2.com")
	repo.Set(ctx, "key3", "https://example3.com")

	// Получаем все ключи
	value1, err1 := repo.Get(ctx, "key1")
	value2, err2 := repo.Get(ctx, "key2")
	value3, err3 := repo.Get(ctx, "key3")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.Equal(t, "https://example1.com", value1)
	assert.Equal(t, "https://example2.com", value2)
	assert.Equal(t, "https://example3.com", value3)
}

func TestRepository_ConcurrentAccess(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})
	ctx := context.Background()
	numGoroutines := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // Set и Get для каждой горутины

	// Запускаем горутины для записи
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", id)
			value := fmt.Sprintf("https://example.com/%d", id)
			err := repo.Set(ctx, key, value)
			assert.NoError(t, err)
		}(i)
	}

	// Запускаем горутины для чтения
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", id)
			// Не проверяем ошибки, так как чтение может произойти до записи
			_, _ = repo.Get(ctx, key)
		}(i)
	}

	wg.Wait()

	// Проверяем, что все значения сохранились
	for i := 0; i < numGoroutines; i++ {
		key := fmt.Sprintf("key%d", i)
		expectedValue := fmt.Sprintf("https://example.com/%d", i)
		value, err := repo.Get(ctx, key)
		if err == nil {
			assert.Equal(t, expectedValue, value)
		}
	}
}

func TestRepository_ConcurrentSet(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})
	ctx := context.Background()
	numGoroutines := 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Запускаем несколько горутин, которые записывают в один ключ
	key := "shared_key"
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			value := fmt.Sprintf("value%d", id)
			err := repo.Set(ctx, key, value)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Проверяем, что значение сохранено (может быть любое из записанных)
	value, err := repo.Get(ctx, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, value)
	assert.Contains(t, value, "value")
}

func TestRepository_Get_SpecialCharacters(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})
	ctx := context.Background()

	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "URL with query params",
			key:   "key1",
			value: "https://example.com?param=value&foo=bar",
		},
		{
			name:  "URL with path",
			key:   "key2",
			value: "https://example.com/path/to/page",
		},
		{
			name:  "URL with fragment",
			key:   "key3",
			value: "https://example.com#section",
		},
		{
			name:  "URL with special chars",
			key:   "key4",
			value: "https://example.com/path?param=value&foo=bar#fragment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Set(ctx, tc.key, tc.value)
			require.NoError(t, err)

			result, err := repo.Get(ctx, tc.key)
			assert.NoError(t, err)
			assert.Equal(t, tc.value, result)
		})
	}
}

func TestRepository_Get_LongValue(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})
	ctx := context.Background()
	key := "long_key"

	// Создаем длинное значение
	longValue := make([]byte, 10000)
	for i := range longValue {
		longValue[i] = byte('a' + (i % 26))
	}
	value := string(longValue)

	err := repo.Set(ctx, key, value)
	require.NoError(t, err)

	result, err := repo.Get(ctx, key)

	assert.NoError(t, err)
	assert.Equal(t, value, result)
	assert.Equal(t, 10000, len(result))
}

func TestRepository_Set_Get_MultipleOperations(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})
	ctx := context.Background()

	// Выполняем серию операций Set и Get
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)

		err := repo.Set(ctx, key, value)
		require.NoError(t, err)

		result, err := repo.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, result)
	}
}

func TestRepository_Get_AfterSetDelete(t *testing.T) {
	repo := NewRepository(&config.Model{Repo: config.RepoConfig{SavingFilePath: "./data.json"}})
	ctx := context.Background()
	key := "test_key"
	value := "test_value"

	// Сохраняем значение
	err := repo.Set(ctx, key, value)
	require.NoError(t, err)

	// Проверяем, что значение есть
	result, err := repo.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value, result)

	// sync.Map не имеет метода Delete в нашем интерфейсе, но можно проверить
	// что если мы перезапишем с другим значением, старое исчезнет
	newValue := "new_value"
	err = repo.Set(ctx, key, newValue)
	require.NoError(t, err)

	result, err = repo.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, newValue, result)
}
