package metrics

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type Server struct {
	srv    http.Server
	logger *logrus.Logger
}

func NewServer(logger *logrus.Logger) *Server {
	return &Server{logger: logger}
}

func (s *Server) Start(address string) error {
	if len(strings.TrimSpace(address)) > 0 {
		s.srv = http.Server{Handler: promhttp.Handler()}

		listener, err := net.Listen("tcp", address)
		if err != nil {
			return err
		}

		s.logger.Infof("Metrics server started on %v", address)
		go s.srv.Serve(listener)
	}

	return nil
}

func (s *Server) Stop() {
	if err := s.srv.Shutdown(context.Background()); err != nil {
		s.logger.Infof("Metrics server failed to shutdown with error: %v", err)
	}
}
