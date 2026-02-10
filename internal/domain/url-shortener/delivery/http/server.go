// Package http implements the public REST API facade over the business use-case
// layer. All endpoints are grouped under the legacy prefix "/app" for mobile
// backward-compatibility.
package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MV7VM/url-shortener/internal/config"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/metrics/watcher"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/entities"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/usecase"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server exposes the public HTTP API of the URL shortener service.
// It wires together Gin engine, business use-case layer and audit metrics.
type Server struct {
	logger  *zap.Logger
	serv    *gin.Engine
	cfg     *config.Model
	uc      uc
	auditor Auditor
}

type uc interface {
	GetByID(context.Context, string) (string, bool, error)
	CreateShortURL(context.Context, string, string) (string, bool, error)
	Ping(ctx context.Context) error
	BatchURLs(ctx context.Context, urls []entities.BatchItem, userID string) error
	GetUsersUrls(ctx context.Context, userID string) ([]entities.Item, error)
	Delete(ctx context.Context, shortURL []string, userID string) error
}

// Auditor describes a component that receives events about user interaction
// with short URLs (creation, redirects, deletions) for further processing.
type Auditor interface {
	Notify(event *entities.Event)
}

// NewServer wires up Gin, logging and use-case dependencies.
func NewServer(logger *zap.Logger, cfg *config.Model, uc *usecase.Usecase, auditor *watcher.Watcher) (*Server, error) {
	if cfg.HTTP.ReturningURL[len(cfg.HTTP.ReturningURL)-1] != '/' {
		cfg.HTTP.ReturningURL += "/"
	}
	// Gin already installs its own recovery & logging middleware; leave as-is.
	return &Server{
		logger:  logger,
		serv:    gin.Default(),
		uc:      uc,
		cfg:     cfg,
		auditor: auditor,
	}, nil
}

// OnStart registers routes and launches an HTTP listener in a goroutine.
func (s *Server) OnStart(_ context.Context) error {
	go func() {
		s.createController()

		s.logger.Info("HTTP server started", zap.String("addr", s.cfg.HTTP.Host))
		if err := s.serv.Run(s.cfg.HTTP.Host); err != nil {
			s.logger.Error("HTTP server exited", zap.Error(err))
		}
	}()

	return nil
}

// OnStop is a no-op here (Gin has no explicit shutdown hook).
func (s *Server) OnStop(_ context.Context) error {
	s.logger.Info("HTTP server stopped")
	return nil
}

// CreateShortURL handles POST "/" requests with a plain-text URL in the body
// and returns a shortened URL as a text response.
func (s *Server) CreateShortURL(c *gin.Context) {
	// Получаем raw body
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to read request body: " + err.Error(),
		})
		return
	}

	url := strings.TrimSpace(string(body))
	if !validateURL(url) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid url",
		})
		return
	}

	shortURL, conflict, err := s.uc.CreateShortURL(c.Request.Context(), url, c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	s.auditor.Notify(&entities.Event{
		TS:     int(time.Now().Unix()),
		Action: entities.ActionShort,
		UserID: c.GetString("userID"),
		URL:    url,
	})

	if conflict {
		c.String(http.StatusConflict, s.cfg.HTTP.ReturningURL+shortURL)
		return
	}

	c.String(http.StatusCreated, s.cfg.HTTP.ReturningURL+shortURL)
}

// CreateShortURLByBodyReq describes the JSON payload for POST "/api/shorten",
// where the original URL is passed in the "url" field.
type CreateShortURLByBodyReq struct {
	URL string `json:"url"`
}

// CreateShortURLByBodyResp describes the JSON response from POST "/api/shorten"
// containing the resulting short URL in the "result" field.
type CreateShortURLByBodyResp struct {
	ShortURL string `json:"result"`
}

// CreateShortURLByBody handles POST "/api/shorten" with a JSON payload and
// returns a JSON object with a shortened URL.
func (s *Server) CreateShortURLByBody(c *gin.Context) {
	// Получаем raw body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	// Декодируем JSON в структуру
	var reqBody CreateShortURLByBodyReq

	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
		})
		return
	}

	url := strings.TrimSpace(reqBody.URL)
	if !validateURL(url) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid url",
		})
		return
	}

	shortURL, conflict, err := s.uc.CreateShortURL(c.Request.Context(), url, c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	s.auditor.Notify(&entities.Event{
		TS:     int(time.Now().Unix()),
		Action: entities.ActionShort,
		UserID: c.GetString("userID"),
		URL:    url,
	})

	if conflict {
		c.JSON(http.StatusConflict, CreateShortURLByBodyResp{
			ShortURL: s.cfg.HTTP.ReturningURL + shortURL,
		})
		return
	}

	c.JSON(http.StatusCreated, CreateShortURLByBodyResp{
		ShortURL: s.cfg.HTTP.ReturningURL + shortURL,
	})
}

// GetByID handles GET "/:id" requests and redirects the client to the original
// URL if it exists and is not marked as deleted.
func (s *Server) GetByID(c *gin.Context) {
	id := c.Param("id")

	url, isDeleted, err := s.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		s.logger.Error("failed to get url", zap.String("url", id), zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	s.auditor.Notify(&entities.Event{
		TS:     int(time.Now().Unix()),
		Action: entities.ActionFollow,
		UserID: c.GetString("userID"),
		URL:    url,
	})

	if isDeleted {
		c.AbortWithStatus(http.StatusGone)
		return
	}

	c.Header("Location", url)
	c.Status(http.StatusTemporaryRedirect)
}

// Ping handles GET "/ping" requests and checks the availability of the
// underlying storage via use-case Ping method.
func (s *Server) Ping(c *gin.Context) {
	err := s.uc.Ping(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}

// BatchURL handles POST "/api/shorten/batch" requests and creates multiple
// short URLs in a single call, returning the enriched batch payload.
func (s *Server) BatchURL(c *gin.Context) { //todo 409
	var batchedReq []entities.BatchItem
	if err := c.ShouldBindJSON(&batchedReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to read request body",
		})
		return
	}

	if len(batchedReq) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "batch payload is empty",
		})
		return
	}

	err := s.uc.BatchURLs(c.Request.Context(), batchedReq, c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	for i := range batchedReq {
		batchedReq[i].ShortURL = s.cfg.HTTP.ReturningURL + batchedReq[i].ShortURL
	}

	c.JSON(http.StatusCreated, batchedReq)
}

// GetUsersUrls handles GET "/api/user/urls" requests and returns all URLs
// previously created by the authenticated user.
func (s *Server) GetUsersUrls(c *gin.Context) {
	urls, err := s.uc.GetUsersUrls(c.Request.Context(), c.GetString("userID"))
	if err != nil {
		s.logger.Error("failed to get urls", zap.Error(err))
		return
	}

	if len(urls) == 0 {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	for i := range urls {
		urls[i].ShortURL = s.cfg.HTTP.ReturningURL + urls[i].ShortURL
	}

	c.JSON(http.StatusOK, urls)
}

// DeleteURLs handles DELETE "/api/user/urls" requests and accepts an array of
// short URL identifiers that should be marked as deleted asynchronously.
func (s *Server) DeleteURLs(c *gin.Context) {
	var items []string

	// Привязываем JSON из тела запроса
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON format",
			"details": err.Error(),
		})
		return
	}

	//В случае успешного приёма запроса хендлер должен возвращать HTTP-статус 202 Accepted.
	//Фактический результат удаления может происходить позже — оповещать пользователя об успешности или неуспешности не нужно.
	//context.Background() применен исходя из задания
	go s.uc.Delete(c, items, c.GetString("userID"))

	c.AbortWithStatus(http.StatusAccepted)
}

func validateURL(urlStr string) bool {
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return false
	}

	// Пытаемся распарсить URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Если нет схемы, добавляем http:// и пытаемся снова
	if u.Scheme == "" {
		u, err = url.Parse("http://" + urlStr)
		if err != nil {
			return false
		}
	}

	// Проверяем, что есть host
	if u.Host == "" {
		return false
	}

	// Проверяем, что схема поддерживается
	if u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	return true
}
