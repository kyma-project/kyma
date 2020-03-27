package main

import (
	"fmt"
	"os"

	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/migrator"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const configPrefix = "MIGRATOR"

type config struct {
	KubeconfigPath string `envconfig:"optional"`
	DevLog         bool
	Cfg            migrator.Config
}

var setupLog = log.Log.WithName("setup")

func main() {
	cfg, err := loadConfig(configPrefix)
	if err != nil {
		panic(errors.Wrap(err, "while parsing env variables"))
	}

	log.SetLogger(zap.New(zap.UseDevMode(cfg.DevLog)))

	setupLog.Info(fmt.Sprintf("Configuration: %#v", cfg))

	restConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	failIfErr(err, "Unable to generate Kubernetes rest config")

	mgrt, err := migrator.New(restConfig, cfg.Cfg)
	failIfErr(err, "Unable to create migrator instance")

	err = mgrt.Run()
	failIfErr(err, "while running migration")
}

func failIfErr(err error, msg string) {
	if err != nil {
		setupLog.Error(err, msg)
		os.Exit(1)
	}
}

func newRestClientConfig(kubeconfigPath string) (*restclient.Config, error) {
	if kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}
	return restclient.InClusterConfig()
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}
