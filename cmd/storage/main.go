package main

import (
	"fmt"
	"log/slog"
	"os"

	partHandler "github.com/ohnomail00/super-duper-s3/api/storage/part"
	"github.com/ohnomail00/super-duper-s3/config"
	"github.com/ohnomail00/super-duper-s3/engine/storage/part"
	"github.com/ohnomail00/super-duper-s3/http/storage"
	"github.com/ohnomail00/super-duper-s3/logger"
)

func main() {
	cfg := config.LoadStorageConfig()
	slog.Info(fmt.Sprintf("starting with cfg: %+v\n", cfg))
	logger.Init(cfg.LogLevel)

	putHandler := part.NewPut(cfg.StoragePath)
	getHandler := part.NewGet(cfg.StoragePath)
	handlers := partHandler.NewHandlers(&cfg, getHandler, putHandler)

	svc := storage.NewServiceDI(&cfg, handlers)
	if err := svc.ConfigureAPI(); err != nil {
		slog.Error("configure API error", "error", err)
		os.Exit(1)
	}
	if err := svc.ConfigureHTTP(); err != nil {
		slog.Error("configure HTTP error", "error", err)
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("Server storage run on %s:%d...", cfg.Host, cfg.Port))
	if err := svc.Server.ListenAndServe(); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}
