package main

import (
	"flag"

	"github.com/kyma-project/kyma/components/helm-broker/internal/controller"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"

	envs "github.com/kyma-project/kyma/components/helm-broker/internal/config"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/broker"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger"
	"github.com/sirupsen/logrus"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	verbose := flag.Bool("verbose", false, "specify if log verbosely loading configuration")
	flag.Parse()

	ctrCfg, err := envs.LoadControllerConfig(*verbose)
	fatalOnError(err, "while loading config")

	storageConfig := storage.ConfigList(ctrCfg.Storage)
	sFact, err := storage.NewFactory(&storageConfig)
	fatalOnError(err, "while setting up a storage")

	log := logger.New(&ctrCfg.Logger)

	// Get a config to talk to the apiserver
	log.Info("Setting up client for manager")
	cfg, err := config.GetConfig()
	fatalOnError(err, "while setting up a client")

	// Create a new Cmd to provide shared dependencies and start components
	log.Info("Setting up manager")
	mgr, err := manager.New(cfg, manager.Options{MetricsBindAddress: metricsAddr})
	fatalOnError(err, "while setting up a manager")

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	log.Info("Setting up schemes")
	fatalOnError(apis.AddToScheme(mgr.GetScheme()), "while adding AC scheme")
	fatalOnError(v1beta1.AddToScheme(mgr.GetScheme()), "while adding SC scheme")
	fatalOnError(v1alpha1.AddToScheme(mgr.GetScheme()), "while adding CMS scheme")

	docsProvider := controller.NewDocsProvider(mgr.GetClient())
	brokerSyncer := broker.NewServiceBrokerSyncer(mgr.GetClient(), ctrCfg.ClusterServiceBrokerName, log)
	sbFacade := broker.NewBrokersFacade(mgr.GetClient(), brokerSyncer, ctrCfg.Namespace, ctrCfg.ServiceName, log)
	csbFacade := broker.NewClusterBrokersFacade(mgr.GetClient(), brokerSyncer, ctrCfg.Namespace, ctrCfg.ServiceName, ctrCfg.ClusterServiceBrokerName, log)
	bundleProvider := bundle.NewProvider(bundle.NewHTTPRepository(), bundle.NewLoader(ctrCfg.TmpDir, log), log)

	log.Info("Setting up controller")
	acReconcile := controller.NewReconcileAddonsConfiguration(mgr, bundleProvider, sFact.Chart(), sFact.Bundle(), sbFacade, docsProvider, brokerSyncer, ctrCfg.DevelopMode)
	acController := controller.NewAddonsConfigurationController(acReconcile)
	err = acController.Start(mgr)
	fatalOnError(err, "unable to start AddonsConfigurationController")

	cacReconcile := controller.NewReconcileClusterAddonsConfiguration(mgr, bundleProvider, sFact.Chart(), sFact.Bundle(), csbFacade, docsProvider, brokerSyncer, ctrCfg.DevelopMode)
	cacController := controller.NewClusterAddonsConfigurationController(cacReconcile)
	err = cacController.Start(mgr)
	fatalOnError(err, "unable to start ClusterAddonsConfigurationController")

	log.Info("Starting the Controller.")
	err = mgr.Start(signals.SetupSignalHandler())
	fatalOnError(err, "unable to run the manager")
}

func fatalOnError(err error, msg string) {
	if err != nil {
		logrus.Fatalf("%s: %s", msg, err.Error())
	}
}
