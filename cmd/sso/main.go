package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/rinefica/voice_null_sso/internal/app"
	"github.com/rinefica/voice_null_sso/internal/config"
	"github.com/rinefica/voice_null_sso/internal/lib/sl"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	log := sl.SetupLogger(cfg.Env)
	log.Debug("config env:: ", slog.Any("config", cfg))

	applctn := app.NewApp(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	go applctn.GRPCServer.MustRun()

	db, err := sql.Open("postgres", cfg.StoragePath)
	defer db.Close()
	if err != nil {
		log.Debug("DB connect Failed")
		log.Error(err.Error())
	}
	if err = db.Ping(); err != nil {
		log.Debug("DB Ping Failed")
		log.Error(err.Error())
	}
	log.Debug("DB Connection started successfully")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	applctn.GRPCServer.Stop()

	log.Debug("app stopped")
}
