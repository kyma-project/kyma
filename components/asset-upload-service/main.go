package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/uploader"
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
		Endpoint  string `envconfig:"default=minio.kyma.local"`
		Port      int    `envconfig:"default=9000"`
		Secure    bool   `envconfig:"default=true"`
		AccessKey string
		SecretKey string
	}
	Bucket struct {
		Private string `envconfig:"default=private"`
		Public  string `envconfig:"default=public"`
	}
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

	client, err := minio.New(cfg.Upload.Endpoint, cfg.Upload.AccessKey, cfg.Upload.SecretKey, cfg.Upload.Secure)
	fatalOnError(err, "Error during upload client initialization")

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", func(wr http.ResponseWriter, rq *http.Request) {
		err := rq.ParseMultipartForm(32 << 20)
		if err != nil {
			return
		}
		defer rq.MultipartForm.RemoveAll()

		privateFiles := rq.MultipartForm.File["private"]
		publicFiles := rq.MultipartForm.File["public"]
		filesCount := len(publicFiles) + len(privateFiles)

		u := uploader.New(client, cfg.UploadTimeout, cfg.MaxUploadWorkers)

		fileToUploadCh := make(chan uploader.FileUpload, filesCount)

		go func() {
			for _, file := range publicFiles {
				fileToUploadCh <- uploader.FileUpload{
					Bucket: cfg.Bucket.Public,
					File:   file,
				}
			}
			for _, file := range privateFiles {
				fileToUploadCh <- uploader.FileUpload{
					Bucket: cfg.Bucket.Private,
					File:   file,
				}
			}
			close(fileToUploadCh)
		}()

		err = u.UploadFiles(context.Background(), fileToUploadCh, filesCount)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while uploading files"))
		}

		//TODO: Return results
		wr.WriteHeader(http.StatusCreated)
	})

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
