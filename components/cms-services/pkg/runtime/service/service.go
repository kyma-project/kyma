package service

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Config is used to customize the service configuration.
type Config struct {
	Host string `envconfig:"default=127.0.0.1"`
	Port int    `envconfig:"default=3000"`
}

// Service is the interface implemented by Asset Store services.
type Service interface {
	Register(endpoint HTTPEndpoint)
	Start(ctx context.Context) error
}

// HTTPEndpoint is the interface implemented by Asset Store endpoints.
type HTTPEndpoint interface {
	Name() string
	Handle(writer http.ResponseWriter, request *http.Request)
}

type service struct {
	endpoints []HTTPEndpoint
	host      string
	port      int
}

var _ Service = &service{}

// New is the constructor that creates a new Asset Store service.
func New(config Config) Service {
	return &service{
		host: config.Host,
		port: config.Port,
	}
}

func (s *service) setupHandlers() *http.ServeMux {
	mux := http.NewServeMux()

	for _, endpoint := range s.endpoints {
		log.Infof("Registering %s endpoint", endpoint.Name())
		path := fmt.Sprintf("/%s", endpoint.Name())
		mux.HandleFunc(path, endpoint.Handle)
	}

	return mux
}

// Start runs a service.
func (s *service) Start(ctx context.Context) error {
	mux := s.setupHandlers()

	host := fmt.Sprintf("%s:%d", s.host, s.port)

	srv := &http.Server{Addr: host, Handler: mux}
	log.Infof("Service listen at %s", host)

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Errorf("HTTP server Shutdown: %v", err)
		}
	}()

	return srv.ListenAndServe()
}

// Register adds an endpoint to a service.
func (s *service) Register(endpoint HTTPEndpoint) {
	s.endpoints = append(s.endpoints, endpoint)
}
