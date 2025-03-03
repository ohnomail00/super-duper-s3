package storage

import (
	"net"
	"net/http"
	"strconv"

	partHandler "github.com/ohnomail00/super-duper-s3/api/storage/part"
	"github.com/ohnomail00/super-duper-s3/config"
	"github.com/ohnomail00/super-duper-s3/middlewares"
)

type routers struct {
	part *partHandler.Handlers
}

type Service struct {
	cfg     *config.Storage
	Server  *http.Server
	mux     *http.ServeMux
	routers routers
}

func (s *Service) ConfigureAPI() error {
	// Set up CORS before using middleware.
	middlewares.SetupCors(s.cfg.CorsAllowedOrigins)

	// Configure API using provided handlers.
	s.mux.Handle("PUT /{bucket}/{object}/{part}", middlewares.With(s.routers.part.Put))
	s.mux.Handle("GET /{bucket}/{object}/{part}", middlewares.With(s.routers.part.Get))
	return nil
}

func (s *Service) ConfigureHTTP() error {
	addrStr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	s.Server = &http.Server{
		Addr:    addrStr,
		Handler: s.mux,
	}
	return nil
}

func NewServiceDI(cfg *config.Storage, handlers *partHandler.Handlers) *Service {
	mux := http.NewServeMux()
	srv := &Service{
		cfg: cfg,
		mux: mux,
		routers: routers{
			part: handlers,
		},
	}

	return srv
}
