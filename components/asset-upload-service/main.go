package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/bucket"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/configurer"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/requesthandler"
	"github.com/kyma-project/kyma/components/asset-upload-service/pkg/signal"
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// config contains configuration fields used for upload
type config struct {
	Host           string `envconfig:"default=127.0.0.1"`
	Port           int    `envconfig:"default=3000"`
	KubeconfigPath string `envconfig:"optional"`
	ConfigMap      configurer.Config
	Upload         struct {
		Endpoint         string `envconfig:"default=minio.kyma.local"`
		ExternalEndpoint string `envconfig:"default=https://minio.kyma.local"`
		Port             int    `envconfig:"default=443"`
		Secure           bool   `envconfig:"default=true"`
		AccessKey        string
		SecretKey        string
	}
	Bucket           bucket.Config
	MaxUploadWorkers int           `envconfig:"default=10"`
	UploadTimeout    time.Duration `envconfig:"default=30m"`
	Verbose          bool          `envconfig:"default=false"`
}

func main() {
	cfg, err := loadConfig("APP")
	parseFlags(cfg)
	exitOnError(err, "Error while loading app config")

	stopCh := signal.SetupChannel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelOnInterrupt(stopCh, ctx, cancel)

	uploadEndpoint := fmt.Sprintf("%s:%d", cfg.Upload.Endpoint, cfg.Upload.Port)

	client, err := minio.New(uploadEndpoint, cfg.Upload.AccessKey, cfg.Upload.SecretKey, cfg.Upload.Secure)
	exitOnError(err, "Error during upload client initialization")

	k8sConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	exitOnError(err, "Error while initializing REST client config")
	k8sCoreCli, err := corev1.NewForConfig(k8sConfig)
	exitOnError(err, "Error during K8s Core client initialization")
	c := configurer.New(k8sCoreCli, cfg.ConfigMap)
	exitOnError(err, "Error during configurer creation")

	var buckets bucket.SystemBucketNames
	sharedConfig, err := c.Load()
	exitOnError(err, "Error during loading configuration")

	if sharedConfig == nil {
		handler := bucket.NewHandler(client, cfg.Bucket)
		buckets, err = handler.CreateSystemBuckets()
		exitOnError(err, "Error during creating system buckets")

		err = c.Save(configurer.SharedAppConfig{SystemBuckets: buckets})
		exitOnError(err, "Error during saving system bucket configuration")
	} else {
		buckets = sharedConfig.SystemBuckets
	}

	var uploadExternalEndpoint string
	if cfg.Upload.ExternalEndpoint != "" {
		uploadExternalEndpoint = cfg.Upload.ExternalEndpoint
	} else {
		uploadExternalEndpoint = cfg.Upload.Endpoint
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/upload", requesthandler.New(client, buckets, uploadExternalEndpoint, cfg.UploadTimeout, cfg.MaxUploadWorkers))

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

func newRestClientConfig(kubeconfigPath string) (*restclient.Config, error) {
	var config *restclient.Config
	var err error
	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		config, err = restclient.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}
	return config, nil
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
