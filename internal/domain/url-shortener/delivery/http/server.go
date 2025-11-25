// Package http implements the public REST API facade over the business use-case
// layer.  All endpoints are grouped under the legacy prefix "/app" for mobile
// backward-compatibility.
package http

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MV7VM/url-shortener/internal/config"
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

func (s *Server) withLogger(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		handler(c)

		s.logger.Info("",
			zap.String("uri", c.Request.RequestURI),
			zap.String("method", c.Request.Method),
			zap.Any("duration", time.Since(startTime)),
		)
	}
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
	var body CreateShortURLByBodyReq
	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to read request body: " + err.Error(),
		})
		return
	}

	url := strings.TrimSpace(body.URL)
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
