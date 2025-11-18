package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockUsecase struct {
	GetByIDFunc        func(context.Context, string) (string, error)
	CreateShortURLFunc func(context.Context, string) (string, error)
}

func (m *mockUsecase) GetByID(ctx context.Context, id string) (string, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return "", errors.New("not implemented")
}

func (m *mockUsecase) CreateShortURL(ctx context.Context, url string) (string, error) {
	if m.CreateShortURLFunc != nil {
		return m.CreateShortURLFunc(ctx, url)
	}
	return "", errors.New("not implemented")
}

func setupTestRouter(s *Server) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/", s.CreateShortURL)
	router.GET("/:id", s.GetByID)
	return router
}

func TestServer_CreateShortURL_Success(t *testing.T) {
	logger := zap.NewNop()
	mockUC := &mockUsecase{
		CreateShortURLFunc: func(ctx context.Context, url string) (string, error) {
			assert.Equal(t, "https://example.com", url)
			return "abc123", nil
		},
	}

	server := &Server{
		logger: logger,
		uc:     mockUC,
	}

	router := setupTestRouter(server)

	reqBody := "https://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "abc123", rec.Body.String())
}

func TestServer_CreateShortURL_EmptyBody(t *testing.T) {
	logger := zap.NewNop()
	mockUC := &mockUsecase{}

	server := &Server{
		logger: logger,
		uc:     mockUC,
	}

	router := setupTestRouter(server)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestServer_CreateShortURL_UsecaseError(t *testing.T) {
	logger := zap.NewNop()
	mockUC := &mockUsecase{
		CreateShortURLFunc: func(ctx context.Context, url string) (string, error) {
			return "", errors.New("database error")
		},
	}

	server := &Server{
		logger: logger,
		uc:     mockUC,
	}

	router := setupTestRouter(server)

	reqBody := "https://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "database error", resp["error"])
}

func TestServer_CreateShortURL_WithWhitespace(t *testing.T) {
	logger := zap.NewNop()
	mockUC := &mockUsecase{
		CreateShortURLFunc: func(ctx context.Context, url string) (string, error) {
			assert.Equal(t, "https://example.com", url)
			return "xyz789", nil
		},
	}

	server := &Server{
		logger: logger,
		uc:     mockUC,
	}

	router := setupTestRouter(server)

	reqBody := "  https://example.com  "
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "xyz789", rec.Body.String())
}

func TestServer_GetByID_Success(t *testing.T) {
	logger := zap.NewNop()
	mockUC := &mockUsecase{
		GetByIDFunc: func(ctx context.Context, id string) (string, error) {
			assert.Equal(t, "abc123", id)
			return "https://example.com", nil
		},
	}

	server := &Server{
		logger: logger,
		uc:     mockUC,
	}

	router := setupTestRouter(server)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)

	var resp struct {
		Location string `json:"Location"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", resp.Location)
}

func TestServer_GetByID_NotFound(t *testing.T) {
	logger := zap.NewNop()
	mockUC := &mockUsecase{
		GetByIDFunc: func(ctx context.Context, id string) (string, error) {
			return "", errors.New("not found")
		},
	}

	server := &Server{
		logger: logger,
		uc:     mockUC,
	}

	router := setupTestRouter(server)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid http url",
			url:      "http://example.com",
			expected: true,
		},
		{
			name:     "valid https url",
			url:      "https://example.com",
			expected: true,
		},
		{
			name:     "valid url with www",
			url:      "https://www.example.com",
			expected: true,
		},
		{
			name:     "valid url with path",
			url:      "https://example.com/path/to/page",
			expected: true,
		},
		{
			name:     "valid url without protocol",
			url:      "example.com",
			expected: true,
		},
		{
			name:     "invalid url - empty string",
			url:      "",
			expected: false,
		},
		{
			name:     "invalid url - missing domain",
			url:      "http://",
			expected: false,
		},
		{
			name:     "valid url with subdomain",
			url:      "https://subdomain.example.com",
			expected: true,
		},
		{
			name:     "valid url with query params",
			url:      "https://example.com?param=value",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateURL(tt.url)
			assert.Equal(t, tt.expected, result, "URL: %s", tt.url)
		})
	}
}
