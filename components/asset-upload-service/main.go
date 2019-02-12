package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/asset-upload-service/pkg/signal"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"io/ioutil"
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
		Name   string `envconfig:"default=resources"`
		Region string `envconfig:"default=us-east-1"`
	}
	MaxUploadWorkers int           `envconfig:"default=10"`
	UploadTimeout    time.Duration `envconfig:"default=1h"`
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

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	runServer(stopCh, addr)
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

func runServer(stop <-chan struct{}, addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", func(wr http.ResponseWriter, rq *http.Request) {

		err := rq.ParseMultipartForm(32 << 20)
		if err != nil {
			return
		}

		filesMap := rq.MultipartForm.File
		for i := range filesMap {
			for _, file := range filesMap[i] {
				glog.Infof("Opening file %s...", file.Filename)

				openFile, err := file.Open()
				if err != nil {
					glog.Errorf("while opening file %s", file.Filename)
				}


				bytes, err := ioutil.ReadAll(openFile)
				if err != nil {
					glog.Errorf("while opening file %s", file.Filename)
				}

				fmt.Println(string(bytes))

				err = openFile.Close()
				if err != nil {
					glog.Errorf("while closing file %s", file.Filename)
				}
			}

		}

		return
	})

	srv := &http.Server{Addr: addr, Handler: mux}

	glog.Infof("Listening on %s", addr)

	go func() {
		<-stop
		// Interrupt signal received - shut down the server
		if err := srv.Shutdown(context.Background()); err != nil {
			glog.Errorf("HTTP server Shutdown: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		glog.Errorf("HTTP server ListenAndServe: %v", err)
	}
}
