// Package http implements the public REST API facade over the business use-case
// layer.  All endpoints are grouped under the legacy prefix "/app" for mobile
// backward-compatibility.
package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/MV7VM/url-shortener/internal/config"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/entities"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/usecase"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger
	serv   *gin.Engine
	cfg    *config.Model
	uc     uc
}

type uc interface {
	GetByID(context.Context, string) (string, error)
	CreateShortURL(context.Context, string) (string, error)
	Ping(ctx context.Context) error
	BatchURLs(ctx context.Context, urls []entities.BatchItem) error
}

// NewServer wires up Gin, logging and use-case dependencies.
func NewServer(logger *zap.Logger, cfg *config.Model, uc *usecase.Usecase) (*Server, error) {
	if cfg.HTTP.ReturningURL[len(cfg.HTTP.ReturningURL)-1] != '/' {
		cfg.HTTP.ReturningURL += "/"
	}
	// Gin already installs its own recovery & logging middleware; leave as-is.
	return &Server{
		logger: logger,
		serv:   gin.Default(),
		uc:     uc,
		cfg:    cfg,
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

	shortURL, err := s.uc.CreateShortURL(c.Request.Context(), url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.String(http.StatusCreated, s.cfg.HTTP.ReturningURL+shortURL)
}

type CreateShortURLByBodyReq struct {
	URL string `json:"url"`
}

type CreateShortURLByBodyResp struct {
	ShortURL string `json:"result"`
}

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

	shortURL, err := s.uc.CreateShortURL(c.Request.Context(), url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, CreateShortURLByBodyResp{
		ShortURL: s.cfg.HTTP.ReturningURL + shortURL,
	})
}

func (s *Server) GetByID(c *gin.Context) {
	id := c.Param("id")

	url, err := s.uc.GetByID(c, id)
	if err != nil {
		s.logger.Error("failed to get url", zap.String("url", id), zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Header("Location", url)

	c.Status(http.StatusTemporaryRedirect)
}

func (s *Server) Ping(c *gin.Context) {
	err := s.uc.Ping(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) BatchURL(c *gin.Context) {
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

	err := s.uc.BatchURLs(c, batchedReq)
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
