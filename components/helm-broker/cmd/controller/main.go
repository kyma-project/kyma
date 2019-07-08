package main

import (
	"flag"
	"os"

	scCs "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"
	//hbConfig "github.com/kyma-project/kyma/components/helm-broker/internal/config"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker"
	ctrlCfg "github.com/kyma-project/kyma/components/helm-broker/internal/controller/config"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	//restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	verbose := flag.Bool("verbose", false, "specify if log verbosely loading configuration")
	flag.Parse()

	ctrlCfg, err := ctrlCfg.Load(*verbose)
	fatalOnError(err)

	log := logger.New(&ctrlCfg.Logger)

	storageConfig := storage.ConfigList(ctrlCfg.Storage)
	sFact, err := storage.NewFactory(&storageConfig)
	if err != nil {
		log.Error(err, "unable to get storage factory")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	log.Info("setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "unable to set up client config")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	log.Info("setting up manager")
	mgr, err := manager.New(cfg, manager.Options{MetricsBindAddress: metricsAddr})
	if err != nil {
		log.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	log.Info("setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "unable add APIs to scheme")
		os.Exit(1)
	}

	// TODO: change to InClusterConfig
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", "/Users/i355812/.kube/config")
	//k8sConfig, err := restclient.InClusterConfig()
	fatalOnError(err)

	scClientSet, err := scCs.NewForConfig(k8sConfig)
	fatalOnError(err)

	brokerSyncer := broker.NewServiceBrokerSyncer(scClientSet.ServicecatalogV1beta1(), log)
	sbFacade := broker.NewBrokersFacade(scClientSet.ServicecatalogV1beta1(), brokerSyncer, ctrlCfg.Namespace, ctrlCfg.ServiceName)
	csbFacade := broker.NewClusterBrokersFacade(scClientSet.ServicecatalogV1beta1(), brokerSyncer, ctrlCfg.Namespace, ctrlCfg.ServiceName)

	bundleProvider := bundle.NewProvider(bundle.NewHTTPRepository(), bundle.NewLoader(ctrlCfg.TmpDir, log), log)

	// Setup all Controllers
	log.Info("Setting up controller")
	acReconcile := controller.NewReconcileAddonsConfiguration(mgr, bundleProvider, sbFacade, sFact, ctrlCfg.DevelopMode)
	acController := controller.NewAddonsConfigurationController(acReconcile)
	err = acController.Start(mgr)
	if err != nil {
		log.Error(err, "unable to start AddonsConfigurationController")
	}

	cacReconcile := controller.NewReconcileClusterAddonsConfiguration(mgr, csbFacade)
	cacController := controller.NewClusterAddonsConfigurationController(cacReconcile)
	err = cacController.Start(mgr)
	if err != nil {
		log.Error(err, "unable to start ClusterAddonsConfigurationController")
	}

	// Start the Cmd
	log.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "unable to run the manager")
		os.Exit(1)
	}
}

func fatalOnError(err error) {
	if err != nil {
		logrus.Fatal(err.Error())
	}
}
