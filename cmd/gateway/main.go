package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ohnomail00/super-duper-s3/config"
	"github.com/ohnomail00/super-duper-s3/database"
	"github.com/ohnomail00/super-duper-s3/engine"
	gatewayEngine "github.com/ohnomail00/super-duper-s3/engine/gateway"
	"github.com/ohnomail00/super-duper-s3/engine/hash"
	"github.com/ohnomail00/super-duper-s3/http/clients"
	"github.com/ohnomail00/super-duper-s3/http/gateway"
	"github.com/ohnomail00/super-duper-s3/logger"
)

func main() {
	cfg := config.LoadGatewayConfig()

	logger.Init(cfg.LogLevel)

	hr := hash.NewRing()
	for _, addr := range cfg.StorageAddrs {
		hr.AddNode(engine.Server{Address: addr}, cfg.VirtualReplicas)
	}
	db := database.New()
	uploader, err := gatewayEngine.NewUploader(hr, cfg.PartCount, clients.NewDummyFactory())
	if err != nil {
		slog.Error("failed to create uploader", "error", err)
		os.Exit(1)
	}
	downloader := gatewayEngine.NewDownloader(clients.NewDummyFactory())

	svc := gateway.NewServiceDI(&cfg, db, hr, uploader, downloader)

	if err := svc.ConfigureAPI(); err != nil {
		slog.Error("configure API error", "error", err)
		os.Exit(1)
	}
	if err := svc.ConfigureHTTP(); err != nil {
		slog.Error("configure HTTP error", "error", err)
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("Server gateway run on %s:%d...", cfg.Host, cfg.Port))
	if err := svc.Server.ListenAndServe(); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}
