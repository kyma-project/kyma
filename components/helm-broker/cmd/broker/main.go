package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
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
	"k8s.io/client-go/tools/clientcmd"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	verbose := flag.Bool("verbose", false, "specify if log verbosely loading configuration")
	flag.Parse()
	cfg, err := config.Load(*verbose)
	fatalOnError(err)

	// creates the in-cluster k8sConfig
	k8sConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	fatalOnError(err)

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	fatalOnError(err)

	log := logger.New(&cfg.Logger)

	storageConfig := storage.ConfigList(cfg.Storage)
	sFact, err := storage.NewFactory(&storageConfig)
	fatalOnError(err)

	sch, err := v1alpha1.SchemeBuilder.Build()
	fatalOnError(err)

	dynamicClient, err := client.New(k8sConfig, client.Options{Scheme: sch})
	fatalOnError(err)

	docsProvider := bundle.NewDocsProvider(dynamicClient)
	bundleSyncer := bundle.NewSyncer(sFact.Bundle(), sFact.Chart(), docsProvider, log)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	cancelOnInterrupt(ctx, cancelFunc)

	helmClient := helm.NewClient(cfg.Helm, log)

	srv := broker.New(sFact.Bundle(), sFact.Chart(), sFact.InstanceOperation(), sFact.Instance(), sFact.InstanceBindData(),
		bind.NewRenderer(), bind.NewResolver(clientset.CoreV1()), helmClient, bundleSyncer, log)
	cancelOnInterrupt(ctx, cancelFunc)

	startedCh := make(chan struct{})

	err = srv.Run(ctx, fmt.Sprintf(":%d", cfg.Port), startedCh)
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

func newRestClientConfig(kubeConfigPath string) (*rest.Config, error) {
	if kubeConfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}

	return rest.InClusterConfig()
}
