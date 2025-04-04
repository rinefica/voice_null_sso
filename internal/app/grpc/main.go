package grpcapp

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	authgrpc "github.com/rinefica/voice_null_sso/internal/grpc/auth"
	"github.com/rinefica/voice_null_sso/internal/services/auth"
	"github.com/rinefica/voice_null_sso/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"net"
	"time"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
	storage    storage.Storage
}

func NewApp(
	log *slog.Logger,
	port int,
	storage *storage.Storage,
	tokenTTL time.Duration,
) *App {
	loggingOpts := []logging.Option{logging.WithLogOnEvents(
		logging.PayloadReceived, logging.PayloadSent,
	),
	}
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) error {
			log.Error("recovery from panic", slog.Any("panic", p))
			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(recoveryOpts...),
			logging.UnaryServerInterceptor(interceptorLogger(log), loggingOpts...),
		),
	)

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	authgrpc.Register(gRPCServer, authService)
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const tag = "grpcapp.run"
	log := a.log.With("tag", tag)
	log.Info("starting grpc server")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s failed to listen: %w", tag, err)
	}
	log.Info("grpc server is running", slog.String("addr", lis.Addr().String()))

	if err := a.gRPCServer.Serve(lis); err != nil {
		return fmt.Errorf("%s failed to serve: %w", tag, err)
	}
	return nil
}

func (a *App) Stop() {
	const tag = "grpcapp.stop"
	log := a.log.With("tag", tag)
	log.Info("stopping grpc server")

	a.storage.Close()
	a.gRPCServer.GracefulStop()
}
