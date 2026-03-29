package mux

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/MV7VM/url-shortener/internal/config"
	deliverygrpc "github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/grpc"
	deliveryhttp "github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/http"
	"github.com/soheilhy/cmux"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type runner struct {
	logger     *zap.Logger
	cfg        *config.Model
	httpServer *deliveryhttp.Server
	grpcServer *deliverygrpc.Server

	listener net.Listener
}

func New() fx.Option {
	return fx.Module("server-mux",
		fx.Provide(
			deliveryhttp.NewServer,
			deliverygrpc.NewServer,
			newRunner,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, r *runner) {
				lc.Append(fx.Hook{
					OnStart: r.OnStart,
					OnStop:  r.OnStop,
				})
			},
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("server mux")
		}),
	)
}

func newRunner(logger *zap.Logger, cfg *config.Model, httpServer *deliveryhttp.Server, grpcServer *deliverygrpc.Server) *runner {
	return &runner{
		logger:     logger,
		cfg:        cfg,
		httpServer: httpServer,
		grpcServer: grpcServer,
	}
}

func (r *runner) OnStart(_ context.Context) error {
	if r.cfg.HTTP.IsSecured {
		return errors.New("cmux mode does not support ENABLE_HTTPS yet")
	}

	lis, err := net.Listen("tcp", r.cfg.HTTP.Host)
	if err != nil {
		return err
	}
	r.listener = lis

	m := cmux.New(lis)
	grpcL := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpL := m.Match(cmux.Any())

	go func() {
		if err := r.grpcServer.Serve(grpcL); err != nil && !errors.Is(err, cmux.ErrListenerClosed) && !isClosedConn(err) {
			r.logger.Error("grpc server exited", zap.Error(err))
		}
	}()

	go func() {
		if err := r.httpServer.Serve(httpL); err != nil && !errors.Is(err, cmux.ErrListenerClosed) && !isClosedConn(err) {
			r.logger.Error("http server exited", zap.Error(err))
		}
	}()

	go func() {
		if err := m.Serve(); err != nil && !errors.Is(err, cmux.ErrListenerClosed) && !isClosedConn(err) {
			r.logger.Error("cmux exited", zap.Error(err))
		}
	}()

	r.logger.Info("HTTP + gRPC mux started", zap.String("addr", r.cfg.HTTP.Host))
	return nil
}

func (r *runner) OnStop(ctx context.Context) error {
	if err := r.httpServer.OnStop(ctx); err != nil {
		r.logger.Error("http stop failed", zap.Error(err))
	}
	if err := r.grpcServer.OnStop(ctx); err != nil {
		r.logger.Error("grpc stop failed", zap.Error(err))
	}
	if r.listener != nil {
		if err := r.listener.Close(); err != nil && !isClosedConn(err) {
			r.logger.Error("listener close failed", zap.Error(err))
			return err
		}
	}

	r.logger.Info("HTTP + gRPC mux stopped")
	return nil
}

func isClosedConn(err error) bool {
	return err != nil && strings.Contains(err.Error(), "use of closed network connection")
}
