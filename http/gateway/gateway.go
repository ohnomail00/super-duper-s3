package gateway

import (
	"net"
	"net/http"
	"strconv"

	"github.com/ohnomail00/super-duper-s3/api/gateway/object"
	"github.com/ohnomail00/super-duper-s3/api/gateway/servers"
	"github.com/ohnomail00/super-duper-s3/config"
	"github.com/ohnomail00/super-duper-s3/database"
	"github.com/ohnomail00/super-duper-s3/engine/gateway"
	"github.com/ohnomail00/super-duper-s3/engine/hash"
	"github.com/ohnomail00/super-duper-s3/middlewares"
)

type routers struct {
	object  *object.Handlers
	servers *servers.Handlers
}

type Service struct {
	Cfg    *config.Gateway
	Server *http.Server

	hr        *hash.Ring
	partCount int

	db database.Store

	mux *http.ServeMux

	routers routers
}

func (s *Service) ConfigureAPI() error {
	middlewares.SetupCors(s.Cfg.CorsAllowedOrigins)

	s.mux.Handle("PUT /{bucket}/{object}", middlewares.With(s.routers.object.Put))
	s.mux.Handle("GET /{bucket}/{object}", middlewares.With(s.routers.object.Get))
	s.mux.Handle("POST /server", middlewares.With(s.routers.servers.Add))

	return nil
}

func (s *Service) ConfigureHTTP() error {
	addrStr := net.JoinHostPort(s.Cfg.Host, strconv.Itoa(s.Cfg.Port))
	s.Server = &http.Server{
		Addr:    addrStr,
		Handler: s.mux,
	}

	return nil
}

func NewServiceDI(cfg *config.Gateway, db database.Store, hr *hash.Ring, uploader *gateway.Uploader, downloader *gateway.Downloader) *Service {
	mux := http.NewServeMux()
	srv := &Service{
		Cfg:       cfg,
		mux:       mux,
		db:        db,
		hr:        hr,
		partCount: cfg.PartCount,
	}
	srv.routers = routers{
		object:  object.New(uploader, downloader, db),
		servers: servers.New(cfg, hr),
	}
	return srv
}
