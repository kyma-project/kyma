package main

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/typed/compass/v1alpha1"

	"os"
	"time"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compassconnection"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/kelseyhightower/envconfig"
	apis "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Infoln("Starting Runtime Agent")
	options := parseArgs()
	log.Infof("Options: %s", options)

	var envConfig EnvConfig
	err := envconfig.Process("", &envConfig)
	if err != nil {
		log.Error("Failed to process environment variables")
	}
	log.Infof("Env config: %s", envConfig)

	// Get a config to talk to the apiserver
	log.Info("Setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "unable to set up client config")
		os.Exit(1)
	}

	syncPeriod := time.Second * time.Duration(options.controllerSyncPeriod)

	log.Info("Setting up manager")
	mgr, err := manager.New(cfg, manager.Options{SyncPeriod: &syncPeriod})
	if err != nil {
		log.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	// Setup Scheme for all resources
	log.Info("Setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "Unable add APIs to scheme")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	compassConnectionCRClient, err := v1alpha1.NewForConfig(cfg)
	if err != nil {
		log.Error("Unable to setup Compass Connection CR client")
		os.Exit(1)
	}

	certManager := certificates.NewCredentialsManager()
	compassConfigClient := compass.NewConfigurationClient(graphql.New, options.insecureConfigurationFetch)
	syncService, err := createNewSynchronizationService(cfg, options.integrationNamespace, options.gatewayPort, options.uploadServiceUrl)
	if err != nil {
		log.Errorf("Failed to create synchronization service, %s", err.Error())
		os.Exit(1)
	}

	compassConnector := compass.NewCompassConnector(envConfig.DirectorURL)
	connectionSupervisor := compassconnection.NewSupervisor(
		compassConnector,
		compassConnectionCRClient.CompassConnections(),
		certManager,
		compassConfigClient,
		syncService)

	minimalConfigSyncTime := time.Duration(options.minimalConfigSyncTime) * time.Second

	// Setup all Controllers
	log.Info("Setting up controller")
	if err := compassconnection.InitCompassConnectionController(mgr, connectionSupervisor, minimalConfigSyncTime); err != nil {
		log.Error(err, "Unable to register controllers to the manager")
		os.Exit(1)
	}

	// Initialize Compass Connection CR
	log.Infoln("Initializing Compass Connection CR")
	_, err = connectionSupervisor.InitializeCompassConnection()
	if err != nil {
		log.Error("Unable to initialize Compass Connection CR")
	}

	// Start the Cmd
	log.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Unable to run the manager")
		os.Exit(1)
	}
}
