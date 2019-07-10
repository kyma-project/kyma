package main

import (
	"flag"
	"os"

	scCs "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"

	//hbConfig "github.com/kyma-project/kyma/components/helm-broker/internal/config"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker"
	ctrlCfg "github.com/kyma-project/kyma/components/helm-broker/internal/config"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	verbose := flag.Bool("verbose", false, "specify if log verbosely loading configuration")
	flag.Parse()

	ctrCfg, err := ctrlCfg.LoadControllerConfig(*verbose)
	fatalOnError(err)

	log := logger.New(&ctrCfg.Logger)

	storageConfig := storage.ConfigList(ctrCfg.Storage)
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
	log.Info("setting up schemes")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "unable to add Addons APIs to scheme")
		os.Exit(1)
	}
	if err := v1beta1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "unable to add ServiceCatalog APIs to scheme")
		os.Exit(1)
	}
	if err := v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "unable to add CMS APIs to scheme")
		os.Exit(1)
	}

	scClientSet, err := scCs.NewForConfig(cfg)
	fatalOnError(err)

	dynamicClient, err := client.New(cfg, client.Options{Scheme: mgr.GetScheme()})
	fatalOnError(err)

	docsProvider := bundle.NewDocsProvider(dynamicClient)
	// TODO: change SC client to generic client
	brokerSyncer := broker.NewServiceBrokerSyncer(scClientSet.ServicecatalogV1beta1(), scClientSet.ServicecatalogV1beta1(), ctrCfg.ClusterServiceBrokerName, log)
	sbFacade := broker.NewBrokersFacade(scClientSet.ServicecatalogV1beta1(), brokerSyncer, ctrCfg.Namespace, ctrCfg.ServiceName)
	csbFacade := broker.NewClusterBrokersFacade(scClientSet.ServicecatalogV1beta1(), brokerSyncer, ctrCfg.Namespace, ctrCfg.ServiceName, ctrCfg.ClusterServiceBrokerName)

	bundleProvider := bundle.NewProvider(bundle.NewHTTPRepository(), bundle.NewLoader(ctrCfg.TmpDir, log), log)

	// Setup all Controllers
	log.Info("Setting up controller")
	acReconcile := controller.NewReconcileAddonsConfiguration(mgr, bundleProvider, sbFacade, sFact, ctrCfg.DevelopMode, docsProvider, brokerSyncer)
	acController := controller.NewAddonsConfigurationController(acReconcile)
	err = acController.Start(mgr)
	if err != nil {
		log.Error(err, "unable to start ClusterAddonsConfigurationController")
	}

	cacReconcile := controller.NewReconcileClusterAddonsConfiguration(mgr, bundleProvider, sFact, csbFacade, docsProvider, brokerSyncer, ctrCfg.DevelopMode)
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
