package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kyma-project/kyma/components/helm-broker/internal/bind"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/config"
	"github.com/kyma-project/kyma/components/helm-broker/internal/helm"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const maxPopulatorRetries = 10

func main() {
	verbose := flag.Bool("verbose", false, "specify if log verbosely loading configuration")
	flag.Parse()
	cfg, err := config.Load(*verbose)
	fatalOnError(err)

	// creates the in-cluster k8sConfig
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		panic(err.Error())
	}

	log := logger.New(&cfg.Logger)

	ybLoader := bundle.NewLoader(cfg.TmpDir, log)

	storageConfig := storage.ConfigList(cfg.Storage)
	sFact, err := storage.NewFactory(&storageConfig)
	fatalOnError(err)

	cachePop := bundle.NewPopulator(
		bundle.NewHTTPRepository(cfg.Repository),
		ybLoader, sFact.Bundle(), sFact.Chart(), log,
	)
	for i := 0; i < maxPopulatorRetries; i++ {
		err = cachePop.Init()
		if err == nil {
			break
		}
		log.Errorf("(%d/%d) Could not load bundles from the repository %s: %s", i, maxPopulatorRetries, cfg.Repository.BaseURL(), err.Error())
		time.Sleep(4 * time.Second)
	}
	fatalOnError(err)

	helmClient := helm.NewClient(cfg.Helm, log)

	srv := broker.New(sFact.Bundle(), sFact.Chart(), sFact.InstanceOperation(), sFact.Instance(), sFact.InstanceBindData(),
		bind.NewRenderer(), bind.NewResolver(clientset.CoreV1()), helmClient, log)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	cancelOnInterrupt(ctx, cancelFunc)
	err = srv.Run(ctx, fmt.Sprintf(":%d", cfg.Port))
	fatalOnError(err)
}

func fatalOnError(err error) {
	if err != nil {
		logrus.Fatal(err.Error())
	}
}

// cancelOnInterrupt calls cancel func when os.Interrupt or SIGTERM is received
func cancelOnInterrupt(ctx context.Context, cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
		case <-c:
			cancel()
		}
	}()
}
