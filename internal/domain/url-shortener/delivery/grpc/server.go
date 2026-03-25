// Package http implements the public REST API facade over the business use-case
// layer. All endpoints are grouped under the legacy prefix "/app" for mobile
// backward-compatibility.
package grpc

import (
	"context"
	"net"
	"net/url"
	"strings"
	"sync"

	"github.com/MV7VM/url-shortener/internal/config"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/entities"
	"github.com/MV7VM/url-shortener/internal/domain/url-shortener/usecase"
	shortener_v1 "github.com/MV7VM/url-shortener/pkg/proto/shortener/gen/go"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server exposes the public HTTP API of the URL shortener service.
// It wires together Gin engine, business use-case layer and audit metrics.
type Server struct {
	logger *zap.Logger
	serv   *grpc.Server
	cfg    *config.Model
	uc     uc
	once   sync.Once

	shortener_v1.UnimplementedShortenerServiceServer
}

type uc interface {
	GetByID(context.Context, string) (string, bool, error)                    //
	CreateShortURL(context.Context, string, string) (string, bool, error)     //
	GetUsersUrls(ctx context.Context, userID string) ([]entities.Item, error) //
}

// NewServer wires up Gin, logging and use-case dependencies.
func NewServer(logger *zap.Logger, cfg *config.Model, uc *usecase.Usecase) (*Server, error) {
	if cfg.HTTP.ReturningURL[len(cfg.HTTP.ReturningURL)-1] != '/' {
		cfg.HTTP.ReturningURL += "/"
	}

	// Gin already installs its own recovery & logging middleware; leave as-is.
	s := &Server{
		logger: logger,
		//serv:   grpc.NewServer(),
		uc:  uc,
		cfg: cfg,
	}

	s.serv = grpc.NewServer(grpc.UnaryInterceptor(s.authInterceptor))

	return s, nil
}

// OnStart registers routes and launches an HTTP listener in a goroutine.
func (s *Server) OnStart(_ context.Context) error {
	s.logger.Info("grpc server starting", zap.String("addr", s.cfg.HTTP.Host))

	lis, err := net.Listen("tcp", s.cfg.HTTP.Host)
	if err != nil {
		s.logger.Error("failed to listen", zap.Error(err))
		return err
	}

	s.register()

	go func() {
		if err := s.serv.Serve(lis); err != nil {
			s.logger.Error("HTTP server exited", zap.Error(err))
		}
	}()

	return nil
}

func (s *Server) register() {
	s.once.Do(func() {
		shortener_v1.RegisterShortenerServiceServer(s.serv, s)
		reflection.Register(s.serv)
	})
}

// Serve starts gRPC handling on the provided listener.
// This mode is used by a shared TCP multiplexer.
func (s *Server) Serve(lis net.Listener) error {
	s.register()
	s.logger.Info("grpc server starting", zap.String("addr", lis.Addr().String()))
	return s.serv.Serve(lis)
}

// OnStop is a no-op here (Gin has no explicit shutdown hook).
func (s *Server) OnStop(_ context.Context) error {
	s.logger.Info("grpc server stopped")

	s.serv.GracefulStop()

	return nil
}

func (s *Server) ShortenURL(ctx context.Context, req *shortener_v1.URLShortenRequest) (*shortener_v1.URLShortenResponse, error) {
	userID, err := s.getUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if !validateURL(req.GetUrl()) {
		return nil, status.Error(codes.InvalidArgument, "invalid url")
	}

	shortURL, conflict, err := s.uc.CreateShortURL(ctx, req.GetUrl(), userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create short url")
	}

	if conflict {
		return nil, status.Error(codes.AlreadyExists, "conflicting short url")
	}

	return shortener_v1.URLShortenResponse_builder{Result: shortURL}.Build(), nil
}

func (s *Server) ExpandURL(ctx context.Context, request *shortener_v1.URLExpandRequest) (*shortener_v1.URLExpandResponse, error) {
	url, isDeleted, err := s.uc.GetByID(ctx, request.GetId())
	if err != nil {
		s.logger.Error("failed to get url", zap.String("url", request.GetId()), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get url")
	}

	if isDeleted {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return shortener_v1.URLExpandResponse_builder{Result: url}.Build(), nil
}

func (s *Server) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*shortener_v1.UserURLsResponse, error) {
	userID, err := s.getUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	urls, err := s.uc.GetUsersUrls(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get urls", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get urls")
	}

	if len(urls) == 0 {
		return nil, status.Error(codes.NotFound, "not found")
	}

	for i := range urls {
		urls[i].ShortURL = s.cfg.HTTP.ReturningURL + urls[i].ShortURL
	}

	return shortener_v1.UserURLsResponse_builder{
		Url: s.convertURLs(urls),
	}.Build(), nil
}

func (s *Server) getUserIDFromCtx(ctx context.Context) (string, error) {
	var userID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get("userID")
		if len(values) > 0 {
			userID = values[0]
		} else {
			return "", status.Error(codes.Unauthenticated, "missing userID")
		}
	}

	return userID, nil
}

func (s *Server) authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var (
		token string
		err   error
	)

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get("auth")
		if len(values) > 0 {
			token = values[0]
		}
	}

	if len(token) == 0 {
		token, _, err = s.createAuthToken()
		if err != nil {
			return nil, err
		}

		md.Append("auth", token)
	}

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.cfg.HTTP.SecretToken), nil
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid auth token")
	}

	if claims, ok := tokenParsed.Claims.(jwt.MapClaims); ok && tokenParsed.Valid {
		for key, value := range claims {
			md.Append(key, value.(string))
		}
	}

	resp, errResp := handler(metadata.NewIncomingContext(ctx, md), req)

	header := metadata.Pairs("auth", token)

	if err := grpc.SetHeader(ctx, header); err != nil {
		s.logger.Error("failed to set header", zap.Error(err))
	}

	return resp, errResp
}

func (s *Server) createAuthToken() (string, string, error) {
	userID, err := uuid.NewV7()
	if err != nil {
		return "", "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID.String(),
	})

	signedString, err := token.SignedString([]byte(s.cfg.HTTP.SecretToken))
	if err != nil {
		return "", "", err
	}

	return signedString, userID.String(), nil
}

func (s *Server) convertURLs(urls []entities.Item) []*shortener_v1.URLData {
	res := make([]*shortener_v1.URLData, len(urls))
	for i := range urls {
		res[i] = shortener_v1.URLData_builder{
			ShortUrl:    urls[i].ShortURL,
			OriginalUrl: urls[i].OriginalURL,
		}.Build()
	}

	return res
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
