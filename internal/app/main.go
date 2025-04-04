package app

import (
	grpcapp "github.com/rinefica/voice_null_sso/internal/app/grpc"
	"github.com/rinefica/voice_null_sso/internal/storage"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcapp.App
}

func NewApp(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {

	strg, err := storage.NewStorage(log, storagePath)
	if err != nil {
		panic(err)
	}

	grpcApp := grpcapp.NewApp(log, grpcPort, strg, tokenTTL)
	return &App{
		GRPCServer: grpcApp,
	}
}
