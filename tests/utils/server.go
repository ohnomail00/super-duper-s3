package utils

import (
	"net"
	"net/http/httptest"
	"strconv"
	"testing"

	partHandler "github.com/ohnomail00/super-duper-s3/api/storage/part"
	"github.com/ohnomail00/super-duper-s3/config"
	"github.com/ohnomail00/super-duper-s3/database"
	"github.com/ohnomail00/super-duper-s3/engine"
	gatewayEngine "github.com/ohnomail00/super-duper-s3/engine/gateway"
	"github.com/ohnomail00/super-duper-s3/engine/hash"
	"github.com/ohnomail00/super-duper-s3/engine/storage/part"
	"github.com/ohnomail00/super-duper-s3/http/clients"
	"github.com/ohnomail00/super-duper-s3/http/gateway"
	"github.com/ohnomail00/super-duper-s3/http/storage"
	"github.com/ohnomail00/super-duper-s3/logger"
)

func StartTestStorageServer(t *testing.T) (*httptest.Server, *storage.Service) {
	cfg := config.Storage{
		Host:               "localhost",
		Port:               GetFreePort(),
		StoragePath:        t.TempDir(),
		CorsAllowedOrigins: []string{"*"},
		LogLevel:           "debug",
	}
	logger.Init(cfg.LogLevel)

	putHandler := part.NewPut(cfg.StoragePath)
	getHandler := part.NewGet(cfg.StoragePath)

	handlers := partHandler.NewHandlers(&cfg, getHandler, putHandler)

	svc := storage.NewServiceDI(&cfg, handlers)

	if err := svc.ConfigureAPI(); err != nil {
		t.Fatal("configure API error", "error", err)
	}
	if err := svc.ConfigureHTTP(); err != nil {
		t.Fatal("configure HTTP error", "error", err)
	}

	ts := httptest.NewServer(svc.Server.Handler)
	t.Cleanup(func() {
		ts.Close()
	})
	return ts, svc
}

func StartStorageServerWithStorageDir(t *testing.T, dir string) (*httptest.Server, *storage.Service) {
	cfg := config.Storage{
		Host:               "localhost",
		Port:               GetFreePort(),
		StoragePath:        dir,
		CorsAllowedOrigins: []string{"*"},
		LogLevel:           "debug",
	}
	logger.Init(cfg.LogLevel)

	putHandler := part.NewPut(cfg.StoragePath)
	getHandler := part.NewGet(cfg.StoragePath)

	handlers := partHandler.NewHandlers(&cfg, getHandler, putHandler)

	svc := storage.NewServiceDI(&cfg, handlers)

	if err := svc.ConfigureAPI(); err != nil {
		t.Fatal("configure API error", "error", err)
	}
	if err := svc.ConfigureHTTP(); err != nil {
		t.Fatal("configure HTTP error", "error", err)
	}

	ts := httptest.NewServer(svc.Server.Handler)
	t.Cleanup(func() {
		ts.Close()
	})
	return ts, svc
}

func StartTestGatewayServer(t *testing.T, storageAddrs []string) (*httptest.Server, *gateway.Service) {
	cfg := config.Gateway{
		Host:               "localhost",
		Port:               GetFreePort(),
		StorageAddrs:       storageAddrs,
		VirtualReplicas:    100,
		PartCount:          6,
		CorsAllowedOrigins: []string{"*"},
		LogLevel:           "debug",
	}

	logger.Init(cfg.LogLevel)

	hr := hash.NewRing()
	for _, addr := range cfg.StorageAddrs {
		hr.AddNode(engine.Server{Address: addr}, cfg.VirtualReplicas)
	}
	db := database.New()
	uploader, err := gatewayEngine.NewUploader(hr, cfg.PartCount, clients.NewDummyFactory())
	if err != nil {
		t.Fatal("failed to create uploader", "error", err)
	}
	downloader := gatewayEngine.NewDownloader(clients.NewDummyFactory())

	svc := gateway.NewServiceDI(&cfg, db, hr, uploader, downloader)

	if err := svc.ConfigureAPI(); err != nil {
		t.Fatal("configure API error", "error", err)
	}
	if err := svc.ConfigureHTTP(); err != nil {
		t.Fatal("configure HTTP error", "error", err)
	}

	ts := httptest.NewServer(svc.Server.Handler)
	t.Cleanup(func() {
		ts.Close()
	})
	return ts, svc
}

func GetFreePort() int {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	_, portStr, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		panic(err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}
	return port
}
