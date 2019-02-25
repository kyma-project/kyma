package asset_store

import (
	"flag"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	restclient "k8s.io/client-go/rest"
	"github.com/vrischmann/envconfig"
)

// config contains configuration fields used for upload
type config struct {
	KubeconfigPath string `envconfig:"optional"`
	Verbose        bool   `envconfig:"default=false"`
}

func main() {
	cfg, err := loadConfig("APP")
	parseFlags(cfg)
	exitOnError(err, "Error while loading app config")

	restConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	exitOnError(err, "Error while loading K8s REST config")



	// Upload test data with upload service

	// Create Bucket CR
	// Create asset CR (maybe more? single file and package)

	// Check if assets have been uploaded
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
