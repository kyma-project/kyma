package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	scCs "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	catalogInformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/internal/config"
	"github.com/kyma-project/kyma/components/application-broker/internal/mapping"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage/populator"
	"github.com/kyma-project/kyma/components/application-broker/internal/syncer"
	mappingCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	mappingInformer "github.com/kyma-project/kyma/components/application-broker/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger"
	appCli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	appInformer "github.com/kyma-project/kyma/components/application-operator/pkg/client/informers/externalversions"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// informerResyncPeriod defines how often informer will execute relist action. Setting to zero disable resync.
// BEWARE: too short period time will increase the CPU load.
const informerResyncPeriod = 30 * time.Minute

func main() {
	verbose := flag.Bool("verbose", false, "specify if log verbosely loading configuration")
	flag.Parse()
	cfg, err := config.Load(*verbose)
	fatalOnError(err)

	log := logger.New(&cfg.Logger)

	// create storage factory
	storageConfig := storage.ConfigList(cfg.Storage)
	sFact, err := storage.NewFactory(&storageConfig)
	fatalOnError(err)

	k8sConfig, err := restclient.InClusterConfig()
	fatalOnError(err)

	// k8s
	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	fatalOnError(err)
	nsInformer := v1.NewNamespaceInformer(k8sClient, informerResyncPeriod, cache.Indexers{})

	// ServiceCatalog
	scClientSet, err := scCs.NewForConfig(k8sConfig)
	fatalOnError(err)

	scInformerFactory := catalogInformers.NewSharedInformerFactory(scClientSet, informerResyncPeriod)
	scInformersGroup := scInformerFactory.Servicecatalog().V1beta1()

	// instance populator
	instancePopulator := populator.NewInstances(scClientSet, sFact.Instance(), &populator.Converter{})
	popCtx, popCancelFunc := context.WithTimeout(context.Background(), time.Minute)
	defer popCancelFunc()
	log.Info("Instance storage population...")
	err = instancePopulator.Do(popCtx)
	fatalOnError(err)
	log.Info("Instance storage populated")

	// Applications
	appClient, err := appCli.NewForConfig(k8sConfig)
	fatalOnError(err)
	appInformerFactory := appInformer.NewSharedInformerFactory(appClient, informerResyncPeriod)
	appInformersGroup := appInformerFactory.Applicationconnector().V1alpha1()

	// Mapping
	mClient, err := mappingCli.NewForConfig(k8sConfig)
	fatalOnError(err)
	mInformerFactory := mappingInformer.NewSharedInformerFactory(mClient, informerResyncPeriod)
	mInformersGroup := mInformerFactory.Applicationconnector().V1alpha1()

	// internal services
	nsBrokerSyncer := syncer.NewServiceBrokerSyncer(scClientSet.ServicecatalogV1beta1())
	relistRequester := syncer.NewRelistRequester(nsBrokerSyncer, cfg.BrokerRelistDurationWindow, cfg.UniqueSelectorLabelKey, cfg.UniqueSelectorLabelValue, log)
	siFacade := broker.NewServiceInstanceFacade(scInformersGroup.ServiceInstances().Informer())

	accessChecker := access.New(sFact.Application(), mClient.ApplicationconnectorV1alpha1(), sFact.Instance())

	appSyncCtrl := syncer.New(appInformersGroup.Applications(), sFact.Application(), sFact.Application(), relistRequester, log)

	brokerService, err := broker.NewNsBrokerService()
	fatalOnError(err)

	nsBrokerFacade := nsbroker.NewFacade(scClientSet.ServicecatalogV1beta1(), k8sClient.CoreV1(), brokerService, nsBrokerSyncer, cfg.Namespace, cfg.UniqueSelectorLabelKey, cfg.UniqueSelectorLabelValue, int32(cfg.Port), log)

	mappingCtrl := mapping.New(mInformersGroup.ApplicationMappings().Informer(), nsInformer, k8sClient.CoreV1().Namespaces(), sFact.Application(), nsBrokerFacade, nsBrokerSyncer, log)

	// create broker
	srv := broker.New(sFact.Application(), sFact.Instance(), sFact.InstanceOperation(), accessChecker,
		mClient.ApplicationconnectorV1alpha1(), siFacade, mInformersGroup.ApplicationMappings().Lister(), brokerService, log)

	// setup graceful shutdown signals
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	stopCh := make(chan struct{})
	cancelOnInterrupt(ctx, stopCh, cancelFunc)

	// start informers
	scInformerFactory.Start(stopCh)
	appInformerFactory.Start(stopCh)
	mInformerFactory.Start(stopCh)
	go nsInformer.Run(stopCh)

	// wait for cache sync
	scInformerFactory.WaitForCacheSync(stopCh)
	appInformerFactory.WaitForCacheSync(stopCh)
	mInformerFactory.WaitForCacheSync(stopCh)
	cache.WaitForCacheSync(stopCh, nsInformer.HasSynced)

	// start services & ctrl
	go appSyncCtrl.Run(stopCh)
	go mappingCtrl.Run(stopCh)
	go relistRequester.Run(stopCh)

	fatalOnError(srv.Run(ctx, fmt.Sprintf(":%d", cfg.Port)))
}

func fatalOnError(err error) {
	if err != nil {
		logrus.Fatal(err.Error())
	}
}

// cancelOnInterrupt closes given channel and also calls cancel func when os.Interrupt or SIGTERM is received
func cancelOnInterrupt(ctx context.Context, ch chan<- struct{}, cancel context.CancelFunc) {
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
