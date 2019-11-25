package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	serviceCatalogClientset "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	serviceCatalogInformers "github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/metric"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/usagekind"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/platform/logger"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	bindingUsageClientset "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	bindingUsageInformers "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/signal"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	k8sClientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// informerResyncPeriod defines how often informer will execute relist action. Setting to zero disable resync.
// BEWARE: too short period time will increase the CPU load.
const informerResyncPeriod = 30 * time.Minute

// liveness probe is run in some period of time (e.g. 10s), one of the liveness functionality
// is create ServiceBindingUsage sample, check its state and remove it.
// It is not necessary to run this check too often, that is the reason of livenessInhibitor value,
// it slows down the process, is a time multiplier
// (period of ServiceBindingUsage sample checker = liveness period * livenessInhibitor)
const livenessInhibitor = 10

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

	// k8s Clientset
	k8sCli, err := k8sClientset.NewForConfig(k8sConfig)
	fatalOnError(err)

	// ServiceBindingUsage informers
	bindingUsageCli, err := bindingUsageClientset.NewForConfig(k8sConfig)
	fatalOnError(err)
	bindingUsageInformerFactory := bindingUsageInformers.NewSharedInformerFactory(bindingUsageCli, informerResyncPeriod)

	// ServiceCatalog informers
	serviceCatalogCli, err := serviceCatalogClientset.NewForConfig(k8sConfig)
	fatalOnError(err)
	serviceCatalogInformerFactory := serviceCatalogInformers.NewSharedInformerFactory(serviceCatalogCli, informerResyncPeriod)
	scInformersGroup := serviceCatalogInformerFactory.Servicecatalog().V1beta1()

	// Service Catalog PodPreset client
	// As a temporary solution, client is generated in this repository under /pkg/client.
	// This SHOULD be changed when PodPreset from Service Catalog become production ready.
	svcatPodPresetCli := bindingUsageCli.SettingsV1alpha1()
	podPresetModifier := controller.NewPodPresetModifier(svcatPodPresetCli)

	aggregator := controller.NewResourceSupervisorAggregator()
	sbuInformer := bindingUsageInformerFactory.Servicecatalog().V1alpha1().ServiceBindingUsages()

	cp, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		fatalOnError(err)
	}

	cbm := metric.NewControllerBusinessMetric()
	prometheus.MustRegister(cbm)

	kindController := usagekind.NewKindController(
		bindingUsageInformerFactory.Servicecatalog().V1alpha1().UsageKinds(),
		aggregator,
		cp,
		log,
		cbm)
	ukProtectionController, err := usagekind.NewProtectionController(
		bindingUsageInformerFactory.Servicecatalog().V1alpha1().UsageKinds(),
		sbuInformer,
		bindingUsageCli.ServicecatalogV1alpha1(),
		log,
	)
	fatalOnError(err)

	labelsFetcher := controller.NewBindingLabelsFetcher(scInformersGroup.ServiceInstances().Lister(),
		scInformersGroup.ClusterServiceClasses().Lister(),
		scInformersGroup.ServiceClasses().Lister())

	cfgMapClient := k8sCli.CoreV1().ConfigMaps(cfg.AppliedSBUConfigMapNamespace)
	usageSpecStorage := controller.NewBindingUsageSpecStorage(cfgMapClient, cfg.AppliedSBUConfigMapName)

	ctr := controller.NewServiceBindingUsage(
		usageSpecStorage,
		bindingUsageCli.ServicecatalogV1alpha1(),
		sbuInformer,
		scInformersGroup.ServiceBindings(),
		aggregator,
		podPresetModifier,
		labelsFetcher,
		log,
		cbm,
	)
	ctr.AddOnDeleteListener(ukProtectionController)

	// TODO consider to extract here the cache sync logic from controller
	// and use WaitForCacheSync() method defined on factories
	bindingUsageInformerFactory.Start(stopCh)
	serviceCatalogInformerFactory.Start(stopCh)

	go runHTTPServer(stopCh, fmt.Sprintf(":%d", cfg.Port), bindingUsageCli, cfg.AppliedSBUConfigMapNamespace, log)
	go kindController.Run(stopCh)
	go ukProtectionController.Run(stopCh)

	ctr.Run(stopCh)
}

func runHTTPServer(stop <-chan struct{}, addr string, sbuClient *bindingUsageClientset.Clientset, ns string, log logrus.FieldLogger) {
	counter := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/statusz", func(w http.ResponseWriter, req *http.Request) {
		if counter >= livenessInhibitor {
			if err := informerAvailability(sbuClient, log, ns); err != nil {
				log.Errorf("Cannot apply ServiceBindingUsage sample: %v ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			counter = 0
		}
		counter++

		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprint(w, "OK"); err != nil {
			log.Errorf("Cannot write response body, got err: %v ", err)
		}
	})
	mux.Handle("/metrics", promhttp.Handler())

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

func informerAvailability(sbuClient *bindingUsageClientset.Clientset, log logrus.FieldLogger, namespace string) error {
	deleteSample := func() {
		err := sbuClient.ServicecatalogV1alpha1().ServiceBindingUsages(namespace).Delete(
			controller.LivenessBUCSample,
			&metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("while deleting ServiceBindingUsage sample", err)
		}
	}
	_, err := sbuClient.ServicecatalogV1alpha1().ServiceBindingUsages(namespace).Create(
		&v1alpha1.ServiceBindingUsage{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "servicecatalog.kyma-project.io/v1alpha1",
				Kind:       "ServiceBindingUsage",
			},
			ObjectMeta: metav1.ObjectMeta{Name: controller.LivenessBUCSample},
		})

	switch {
	case k8sErrors.IsAlreadyExists(err):
		deleteSample()
		return nil
	case err != nil:
		return errors.Wrap(err, "while creating ServiceBindingUsage")
	}

	defer func() {
		deleteSample()
	}()
	err = wait.Poll(1*time.Second, 20*time.Second, func() (done bool, err error) {
		sbuSample, err := sbuClient.ServicecatalogV1alpha1().ServiceBindingUsages(namespace).Get(
			controller.LivenessBUCSample,
			metav1.GetOptions{})
		if err != nil {
			return false, errors.Wrap(err, "while fetching ServiceBindingUsage")
		}
		for _, condition := range sbuSample.Status.Conditions {
			if condition.Status == v1alpha1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return errors.Wrap(err, "while checking ServiceBindingUsage status conditions")
	}

	return nil
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
