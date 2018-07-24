package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kyma-project/kyma/tools/stability-checker/internal/notifier"
	"github.com/kyma-project/kyma/tools/stability-checker/internal/runner"
	"github.com/kyma-project/kyma/tools/stability-checker/platform/logger"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config holds application configuration
type Config struct {
	Logger               logger.Config
	Port                 int           `envconfig:"default=8080"`
	KubeConfigPath       string        `envconfig:"optional"`
	TestThrottle         time.Duration `envconfig:"default=5m"`
	WorkingNamespace     string        `envconfig:"default=kyma-system"`
	PodName              string        `envconfig:"HOSTNAME"`
	TestConfigMapName    string        `envconfig:"default=stability-checker"`
	PathToTestingScript  string
	TestResultWindowTime time.Duration `envconfig:"default=6h"`
	SlackClient          notifier.SlackClientConfig
}

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(errors.Wrap(err, "while reading configuration from environment variables"))

	log := logger.New(&cfg.Logger)

	k8sConfig, err := newRestClientConfig(cfg.KubeConfigPath)
	fatalOnError(err)

	ctx := contextCanceledOnInterrupt()

	// k8s client
	k8sCli, err := kubernetes.NewForConfig(k8sConfig)
	fatalOnError(err)

	cfgMapClient := k8sCli.CoreV1().ConfigMaps(cfg.WorkingNamespace)

	// statusz server
	go func() {
		http.HandleFunc("/statusz", func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)
		})

		fatalOnError(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil))
	}()

	// Slack notifier
	slackClient := notifier.NewSlackClient(cfg.SlackClient)
	testRenderer, err := notifier.NewTestRenderer()
	fatalOnError(err)

	sNotifier := notifier.New(slackClient, testRenderer, cfgMapClient, cfg.TestConfigMapName, cfg.TestResultWindowTime, cfg.PodName, cfg.WorkingNamespace, log)
	go sNotifier.Run(ctx)

	// Test runner
	tRunner := runner.NewTestRunner(cfg.TestThrottle, cfg.TestConfigMapName, cfg.PathToTestingScript, cfgMapClient, log)
	tRunner.Run(ctx)
}

func fatalOnError(err error) {
	if err != nil {
		logrus.Fatal(err.Error())
	}
}

// copied from https://github.com/kyma-project/binding-usage-controller/tree/v0.1.3/cmd/controller/main.go:143
func newRestClientConfig(kubeConfigPath string) (*restclient.Config, error) {
	if kubeConfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}

	return restclient.InClusterConfig()
}

// contextCanceledOnInterrupt returns context which is canceled when os.Interrupt or SIGTERM is received
func contextCanceledOnInterrupt() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		cancel()
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return ctx
}
