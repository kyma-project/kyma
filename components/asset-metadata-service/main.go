package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/asset-metadata-service/internal/route"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/signal"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

// config contains configuration fields used for upload
type config struct {
	Host           string        `envconfig:"default=127.0.0.1"`
	Port           int           `envconfig:"default=3000"`
	MaxWorkers     int           `envconfig:"default=10"`
	ProcessTimeout time.Duration `envconfig:"default=10m"`
	Verbose        bool          `envconfig:"default=false"`
}

func main() {
	cfg, err := loadConfig("APP")
	exitOnError(err, "Error while loading app config")
	parseFlags(cfg)

	stopCh := signal.SetupChannel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelOnInterrupt(stopCh, ctx, cancel)

	mux := route.SetupHandlers(cfg.MaxWorkers, cfg.ProcessTimeout)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{Addr: addr, Handler: mux}
	glog.Infof("Listening on %s", addr)

	go func() {
		<-stopCh
		if err := srv.Shutdown(context.Background()); err != nil {
			glog.Errorf("HTTP server Shutdown: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		glog.Errorf("HTTP server ListenAndServe: %v", err)
	}
}

// cancelOnInterrupt calls cancel function when os.Interrupt or SIGTERM is received
func cancelOnInterrupt(stopCh <-chan struct{}, ctx context.Context, cancel context.CancelFunc) {
	go func() {
		select {
		case <-ctx.Done():
		case <-stopCh:
			cancel()
		}
	}()
}

func parseFlags(cfg config) {
	if cfg.Verbose {
		err := flag.Set("stderrthreshold", "INFO")
		if err != nil {
			glog.Error(errors.Wrap(err, "while parsing flags"))
		}
	}
	flag.Parse()
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		glog.Fatal(wrappedError)
	}
}
