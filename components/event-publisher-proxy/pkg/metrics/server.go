package metrics

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
)

const (
	metricsServerLoggerName = "metrics-server"
	readHeaderTimeout       = 5 * time.Second
)

type Server struct {
	srv    http.Server
	logger *logger.Logger
}

func NewServer(logger *logger.Logger) *Server {
	return &Server{logger: logger}
}

func (s *Server) Start(address string) error {
	if len(strings.TrimSpace(address)) > 0 {
		s.srv = http.Server{
			Handler:           promhttp.Handler(),
			ReadHeaderTimeout: readHeaderTimeout,
		}

		listener, err := net.Listen("tcp", address)
		if err != nil {
			return err
		}

		s.namedLogger().Infof("Metrics server started on %v", address)
		go s.srv.Serve(listener) //nolint:errcheck
	}

	return nil
}

func (s *Server) Stop() {
	if err := s.srv.Shutdown(context.Background()); err != nil {
		s.namedLogger().Warnw("Failed to shutdown metrics server", "error", err)
	}
}
func (s *Server) namedLogger() *zap.SugaredLogger {
	return s.logger.WithContext().Named(metricsServerLoggerName)
}
