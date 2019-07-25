package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/authn"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/app"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/signal"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net"
)

func main() {
	cfg, err := loadConfig("APP")
	exitOnError(err, "Error while loading app config")
	parseFlags(cfg)

	k8sConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	exitOnError(err, "Error while initializing REST client config")

	var authRequest authenticator.Request
	if cfg.OIDC.IssuerURL != "" {
		authRequest, err = authn.NewOIDCAuthenticator(&cfg.OIDC)
	}
	exitOnError(err, "Error while creating OIDC authenticator")

	stopCh := signal.SetupChannel()

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	listener, err := net.Listen("tcp", addr)
	exitOnError(err, "Error while binding listener")
	glog.Infof("Listening on %s", addr)

	err = app.Run(listener, stopCh, cfg, k8sConfig, authRequest)
	if err != nil {
		glog.Fatal(err)
	}
}

func loadConfig(prefix string) (app.Config, error) {
	cfg := app.Config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		glog.Fatal(wrappedError)
	}
}

func parseFlags(cfg app.Config) {
	if cfg.Verbose {
		err := flag.Set("stderrthreshold", "INFO")
		if err != nil {
			glog.Error(errors.Wrap(err, "while parsing flags"))
		}
	}
	flag.Parse()
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
