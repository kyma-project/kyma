package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/bucket"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/requesthandler"
	"github.com/kyma-project/kyma/components/asset-upload-service/pkg/signal"
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"net/http"
	"time"
)

// config contains configuration fields used for upload
type config struct {
	Host   string `envconfig:"default=127.0.0.1"`
	Port   int    `envconfig:"default=3003"`
	Upload struct {
		Endpoint  string `envconfig:"default=play.minio.io"`
		Port      int    `envconfig:"default=9000"`
		Secure    bool   `envconfig:"default=true"`
		AccessKey string `envconfig:"default=Q3AM3UQ867SPQQA43P2F"`
		SecretKey string `envconfig:"default=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"`
	}
	Bucket           bucket.Config
	MaxUploadWorkers int           `envconfig:"default=10"`
	UploadTimeout    time.Duration `envconfig:"default=30m"`
	Verbose          bool          `envconfig:"default=false"`
}

func main() {
	cfg, err := loadConfig("APP")
	parseFlags(cfg)
	fatalOnError(err, "Error while loading app config")

	stopCh := signal.SetupChannel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelOnInterrupt(stopCh, ctx, cancel)

	endpoint := fmt.Sprintf("%s:%d", cfg.Upload.Endpoint, cfg.Upload.Port)
	client, err := minio.New(endpoint, cfg.Upload.AccessKey, cfg.Upload.SecretKey, cfg.Upload.Secure)
	fatalOnError(err, "Error during upload client initialization")

	handler := bucket.NewHandler(client, cfg.Bucket)
	buckets, err := handler.CreateSystemBuckets()
	fatalOnError(err, "Error during creating buckets")

	//TODO: Read and Save bucket names from configmap

	mux := http.NewServeMux()
	mux.Handle("/upload", requesthandler.New(client, buckets, cfg.UploadTimeout, cfg.MaxUploadWorkers))

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

func fatalOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		glog.Fatal(wrappedError)
	}
}
