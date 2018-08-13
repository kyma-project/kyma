package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	serviceCatalogClientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	serviceCatalogInformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/usagekind"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/platform/logger"
	bindingUsageClientset "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned"
	bindingUsageInformers "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/signal"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	k8sClientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// informerResyncPeriod defines how often informer will execute relist action. Setting to zero disable resync.
// BEWARE: too short period time will increase the CPU load.
const informerResyncPeriod = 30 * time.Minute

// Config holds application configuration
type Config struct {
	Logger                       logger.Config
	Port                         int    `envconfig:"default=8080"`
	KubeconfigPath               string `envconfig:"optional"`
	AppliedSBUConfigMapName      string `envconfig:"default=applied-sbu-spec"`
	AppliedSBUConfigMapNamespace string `envconfig:"default=kyma-system"`
}

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(errors.Wrap(err, "while reading configuration from environment variables"))

	log := logger.New(&cfg.Logger)
	// set up signals so we can handle the first shutdown signal gracefully
	stopCh := signal.SetupChannel()

	k8sConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	fatalOnError(err)

	// k8s informers
	k8sCli, err := k8sClientset.NewForConfig(k8sConfig)
	fatalOnError(err)
	k8sInformersFactory := informers.NewSharedInformerFactory(k8sCli, informerResyncPeriod)

	// ServiceBindingUsage informers
	bindingUsageCli, err := bindingUsageClientset.NewForConfig(k8sConfig)
	fatalOnError(err)
	bindingUsageInformerFactory := bindingUsageInformers.NewSharedInformerFactory(bindingUsageCli, informerResyncPeriod)

	// ServiceCatalog informers
	serviceCatalogCli, err := serviceCatalogClientset.NewForConfig(k8sConfig)
	fatalOnError(err)
	serviceCatalogInformerFactory := serviceCatalogInformers.NewSharedInformerFactory(serviceCatalogCli, informerResyncPeriod)

	podPresetModifier := controller.NewPodPresetModifier(k8sCli.SettingsV1alpha1())

	aggregator := controller.NewResourceSupervisorAggregator()
	sbuInformer := bindingUsageInformerFactory.Servicecatalog().V1alpha1().ServiceBindingUsages()

	cp, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		fatalOnError(err)
	}

	kindController := usagekind.NewKindController(
		bindingUsageInformerFactory.Servicecatalog().V1alpha1().UsageKinds(),
		aggregator,
		cp,
		log)
	ukProtectionController := usagekind.NewProtectionController(
		bindingUsageInformerFactory.Servicecatalog().V1alpha1().UsageKinds(),
		sbuInformer,
		bindingUsageCli.ServicecatalogV1alpha1(),
		log,
	)

	labelsFetcher := controller.NewBindingLabelsFetcher(serviceCatalogInformerFactory.Servicecatalog().V1beta1().ServiceInstances().Lister(), serviceCatalogInformerFactory.Servicecatalog().V1beta1().ClusterServiceClasses().Lister())

	cfgMapClient := k8sCli.CoreV1().ConfigMaps(cfg.AppliedSBUConfigMapNamespace)
	usageSpecStorage := controller.NewBindingUsageSpecStorage(cfgMapClient, cfg.AppliedSBUConfigMapName)

	ctr := controller.NewServiceBindingUsage(
		usageSpecStorage,
		bindingUsageCli.ServicecatalogV1alpha1(),
		sbuInformer,
		serviceCatalogInformerFactory.Servicecatalog().V1beta1().ServiceBindings(),
		aggregator,
		podPresetModifier,
		labelsFetcher,
		log,
	)
	ctr.AddOnDeleteListener(ukProtectionController)

	// TODO consider to extract here the cache sync logic from controller
	// and use WaitForCacheSync() method defined on factories
	k8sInformersFactory.Start(stopCh)
	bindingUsageInformerFactory.Start(stopCh)
	serviceCatalogInformerFactory.Start(stopCh)

	go runStatuszHTTPServer(stopCh, fmt.Sprintf(":%d", cfg.Port), log)
	go kindController.Run(stopCh)
	go ukProtectionController.Run(stopCh)

	ctr.Run(stopCh)
}

func runStatuszHTTPServer(stop <-chan struct{}, addr string, log logrus.FieldLogger) {
	mux := http.NewServeMux()
	mux.HandleFunc("/statusz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		<-stop
		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Errorf("HTTP server Shutdown: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Errorf("HTTP server ListenAndServe: %v", err)
	}
}

func fatalOnError(err error) {
	if err != nil {
		logrus.Fatal(err.Error())
	}
}

func newRestClientConfig(kubeConfigPath string) (*restclient.Config, error) {
	if kubeConfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}

	return restclient.InClusterConfig()
}
