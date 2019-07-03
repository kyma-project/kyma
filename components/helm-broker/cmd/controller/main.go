package main

import (
	"flag"
	"github.com/davecgh/go-spew/spew"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"os"

	"github.com/kyma-project/kyma/components/helm-broker/internal/controller"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"
	//hbConfig "github.com/kyma-project/kyma/components/helm-broker/internal/config"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.Parse()
	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("entrypoint")
	//hbCfg, err := hbConfig.Load(false)
	//if err != nil {
	//	log.Error(err, "unable to set up helm broker config")
	//	os.Exit(1)
	//}

	//storageConfig := storage.ConfigList(hbCfg.Storage)
	storageConfig := storage.ConfigList{
		{
			Driver: storage.DriverMemory,
			Provide: storage.ProviderConfigMap{
				storage.EntityAll: storage.ProviderConfig{},
			},
		},
	}
	sFact, err := storage.NewFactory(&storageConfig)
	if err != nil {
		log.Error(err, "unable to get storage factory")
		os.Exit(1)
	}
	spew.Dump(sFact.Bundle().FindAll("default"))

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

	// Setup all Controllers
	log.Info("Setting up controller")
	var clog = logf.Log.WithName("controller")

	acReconcile := controller.NewReconcileAddonsConfiguration(mgr, sFact, clog)
	acController := controller.NewAddonsConfigurationController(acReconcile)
	err = acController.Start(mgr)
	if err != nil {
		log.Error(err, "unable to start AddonsConfigurationController")
	}

	cacReconcile := controller.NewReconcileClusterAddonsConfiguration(mgr, clog)
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
