package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	scCs "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyma-project/kyma/components/helm-broker/internal/bind"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/config"
	"github.com/kyma-project/kyma/components/helm-broker/internal/helm"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	mapLabelKey   = "helm-broker-repo"
	mapLabelValue = "true"
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

	// ServiceCatalog
	scClientSet, err := scCs.NewForConfig(k8sConfig)
	fatalOnError(err)
	csbInterface := scClientSet.ServicecatalogV1beta1().ClusterServiceBrokers()

	// broker sync
	stopCh := make(chan struct{})
	bLoader := bundle.NewLoader(cfg.TmpDir, log)
	bundleSyncer := bundle.NewSyncer(sFact.Bundle(), sFact.Chart(), log)
	brokerSyncer := broker.NewClusterServiceBrokerSyncer(csbInterface, log)
	cfgMapInformer := v1.NewFilteredConfigMapInformer(clientset, cfg.Namespace, 10*time.Minute, cache.Indexers{}, func(options *metav1.ListOptions) {
		options.LabelSelector = fmt.Sprintf("%s=%s", mapLabelKey, mapLabelValue)
	})

	repositoryWatcher := bundle.NewRepositoryController(bundleSyncer, bLoader, brokerSyncer, cfg.ClusterServiceBrokerName, cfgMapInformer, log, cfg.DevelopMode)
	go repositoryWatcher.Run(stopCh)
	go cfgMapInformer.Run(stopCh)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	cancelOnChanInterrupt(ctx, stopCh, cancelFunc)

	helmClient := helm.NewClient(cfg.Helm, log)

	srv := broker.New(sFact.Bundle(), sFact.Chart(), sFact.InstanceOperation(), sFact.Instance(), sFact.InstanceBindData(),
		bind.NewRenderer(), bind.NewResolver(clientset.CoreV1()), helmClient, bundleSyncer, log)
	cancelOnInterrupt(ctx, cancelFunc)

	startedCh := make(chan struct{})
	go func() {
		// wait for server HTTP to be ready
		<-startedCh
		log.Infof("Waiting for service %s to be ready", cfg.HelmBrokerURL)

		// Running Helm Broker does not mean it is visible to the service catalog
		// This is the reason of the check cfg.HelmBrokerURL
		waitForHelmBrokerIsReady(cfg.HelmBrokerURL, 15*time.Second, log)
		log.Infof("%s service ready", cfg.HelmBrokerURL)

		err := brokerSyncer.Sync(cfg.ClusterServiceBrokerName, 5)
		if err != nil {
			log.Errorf("Could not synchronize Service Catalog with the broker: %s", err)
		}
	}()

	err = srv.Run(ctx, fmt.Sprintf(":%d", cfg.Port), startedCh)
	fatalOnError(err)
}

func waitForHelmBrokerIsReady(url string, timeout time.Duration, log logrus.FieldLogger) {
	timeoutCh := time.After(timeout)
	for {
		r, err := http.Get(fmt.Sprintf("%s/statusz", url))
		if err == nil {
			// no need to read the response
			ioutil.ReadAll(r.Body)
			r.Body.Close()
		}
		if err == nil && r.StatusCode == http.StatusOK {
			break
		}

		select {
		case <-timeoutCh:
			log.Errorf("Waiting for service %s to be ready timeout %s exceeded.", url, timeout.String())
			if err != nil {
				log.Errorf("Last call error: %s", err.Error())
			} else {
				log.Errorf("Last call response status: %s", r.StatusCode)
			}
			return
		default:
			time.Sleep(time.Second)
		}
	}
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

// cancelOnInterrupt closes given channel and also calls cancel func when os.Interrupt or SIGTERM is received
func cancelOnChanInterrupt(ctx context.Context, ch chan<- struct{}, cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
			close(ch)
		case <-c:
			close(ch)
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
