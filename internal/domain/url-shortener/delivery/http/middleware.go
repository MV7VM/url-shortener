package http

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

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

func (s *Server) auth(c *gin.Context) {
	userID := ""
	_, err := c.Cookie("auth")
	if err != nil {
		token := ""
		token, userID, err = s.createAuthToken()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		http.SetCookie(c.Writer, &http.Cookie{
			Name:  "auth",
			Value: token,
			Path:  "/",
			//HttpOnly: true,
			//Secure:   true,
			//SameSite: http.SameSiteNoneMode,
			MaxAge: 0,
		})
	}

	if userID != "" {
		c.Set("userID", userID)
		c.Next()
		return
	}

	err = s.parseToken(c)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Next()
}

func (s *Server) createAuthToken() (string, string, error) {
	userID, _ := uuid.NewV7()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID.String(),
	})

	signedString, err := token.SignedString([]byte(s.cfg.HTTP.SecretToken))
	if err != nil {
		return "", "", err
	}

	return signedString, userID.String(), nil
}

func (s *Server) parseToken(c *gin.Context) error {
	cookie, err := c.Cookie("auth")
	if err != nil {
		return err
	}

	token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.cfg.HTTP.SecretToken), nil
	})
	if err != nil {
		return err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		for key, value := range claims {
			c.Set(key, value)
		}
	}

	return nil
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func newGzipWriter(w gin.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		ResponseWriter: w,
		writer:         gzip.NewWriter(w),
	}
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

func (g *gzipWriter) Close() error {
	return g.writer.Close()
}

type gzipReader struct {
	*gzip.Reader
	closer io.Closer
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	reader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &gzipReader{
		Reader: reader,
		closer: r,
	}, nil
}

func (g *gzipReader) Close() error {
	err1 := g.Reader.Close()
	err2 := g.closer.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

func (s *Server) gzipMiddleware(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := c.GetHeader("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newGzipWriter(c.Writer)
			c.Writer = cw
			c.Header("Content-Encoding", "gzip")
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := c.GetHeader("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newGzipReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			// меняем тело запроса на новое
			c.Request.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		handler(c)
	}
}
