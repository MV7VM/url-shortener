package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/MV7VM/url-shortener/internal/config"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/metrics/watcher"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/entities"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// exampleUC is a minimal implementation of the internal uc interface
// used only in documentation examples.
type exampleUC struct{}

func (exampleUC) GetByID(ctx context.Context, id string) (string, bool, error) {
	return "https://example.com", false, nil
}

func (exampleUC) CreateShortURL(ctx context.Context, url, userID string) (string, bool, error) {
	return "abc123", false, nil
}

func (exampleUC) Ping(ctx context.Context) error {
	return nil
}

func (exampleUC) BatchURLs(ctx context.Context, urls []entities.BatchItem, userID string) error {
	for i := range urls {
		urls[i].ShortURL = fmt.Sprintf("s%d", i+1)
	}
	return nil
}

func (exampleUC) GetUsersUrls(ctx context.Context, userID string) ([]entities.Item, error) {
	return []entities.Item{
		{ShortURL: "abc123", OriginalURL: "https://example.com"},
	}, nil
}

func (exampleUC) Delete(ctx context.Context, shortURL []string, userID string) error {
	return nil
}

// ExampleServer_CreateShortURL demonstrates how to call the text/plain
// shortening endpoint (POST "/") using an in-memory Gin router.
func ExampleServer_CreateShortURL() {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	cfg := &config.Model{
		HTTP: config.HTTPConfig{
			ReturningURL: "http://localhost:8080/",
		},
	}

	server := &Server{
		logger:  logger,
		serv:    gin.New(),
		cfg:     cfg,
		auditor: watcher.NewWatcher(),
		uc:      exampleUC{},
	}

	router := gin.New()
	router.POST("/", server.CreateShortURL)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://example.com"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	fmt.Println(rec.Code, rec.Body.String())
	// Output:
	// 201 http://localhost:8080/abc123
}

// ExampleServer_CreateShortURLByBody shows how to send a JSON payload to
// "/api/shorten" and decode the resulting short URL.
func ExampleServer_CreateShortURLByBody() {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	cfg := &config.Model{
		HTTP: config.HTTPConfig{
			ReturningURL: "http://localhost:8080/",
		},
	}

	server := &Server{
		logger:  logger,
		serv:    gin.New(),
		cfg:     cfg,
		auditor: watcher.NewWatcher(),
		uc:      exampleUC{},
	}

	router := gin.New()
	router.POST("/api/shorten", server.CreateShortURLByBody)

	body := CreateShortURLByBodyReq{URL: "https://example.com"}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	var resp CreateShortURLByBodyResp
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	fmt.Println(rec.Code, resp.ShortURL)
	// Output:
	// 201 http://localhost:8080/abc123
}

// ExampleServer_BatchURL illustrates the shape of the JSON payload and response for
// "/api/shorten/batch".
func ExampleServer_BatchURL() {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	cfg := &config.Model{
		HTTP: config.HTTPConfig{
			ReturningURL: "http://localhost:8080/",
		},
	}

	server := &Server{
		logger:  logger,
		serv:    gin.New(),
		cfg:     cfg,
		auditor: watcher.NewWatcher(),
		uc:      exampleUC{},
	}

	router := gin.New()
	router.POST("/api/shorten/batch", server.BatchURL)

	items := []entities.BatchItem{
		{CorrelationID: "1", OriginalURL: "https://example.com"},
		{CorrelationID: "2", OriginalURL: "https://example.org"},
	}

	data, _ := json.Marshal(items)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	var resp []entities.BatchItem
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	fmt.Println(rec.Code, len(resp), resp[0].ShortURL != "", resp[1].ShortURL != "")
	// Output:
	// 201 2 true true
}
