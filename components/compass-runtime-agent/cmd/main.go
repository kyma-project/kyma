package main

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"kyma-project.io/compass-runtime-agent/internal/certificates"
	"kyma-project.io/compass-runtime-agent/internal/compass"
	confProvider "kyma-project.io/compass-runtime-agent/internal/config"
	"kyma-project.io/compass-runtime-agent/internal/graphql"
	"kyma-project.io/compass-runtime-agent/internal/secrets"

	"os"

	"kyma-project.io/compass-runtime-agent/internal/compassconnection"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	apis "kyma-project.io/compass-runtime-agent/pkg/apis/compass/v1alpha1"
)

func main() {
	log.Infoln("Starting Runtime Agent")

	var options Config
	err := envconfig.InitWithPrefix(&options, "APP")
	if err != nil {
		log.Error("Failed to process environment variables")
	}
	log.Infof("Env config: %s", options.String())

	// Get a config to talk to the apiserver
	log.Info("Setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "unable to set up client config")
		os.Exit(1)
	}

	log.Info("Setting up manager")
	mgr, err := manager.New(cfg, manager.Options{SyncPeriod: &options.ControllerSyncPeriod})
	if err != nil {
		log.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	// Setup Scheme for all resources
	log.Info("Setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "Unable add API to scheme")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	k8sResourceClientSets, err := k8sResourceClients(cfg)
	if err != nil {
		log.Errorf("Failed to initialize K8s resource clients: %s", err.Error())
		os.Exit(1)
	}

	secretsManagerConstructor := func(namespace string) secrets.Manager {
		return k8sResourceClientSets.core.CoreV1().Secrets(namespace)
	}

	secretsRepository := secrets.NewRepository(secretsManagerConstructor)

	clusterCertSecret := parseNamespacedName(options.ClusterCertificatesSecret)
	caCertSecret := parseNamespacedName(options.CaCertificatesSecret)

	certManager := certificates.NewCredentialsManager(clusterCertSecret, caCertSecret, secretsRepository)
	syncService, err := createNewSynchronizationService(
		k8sResourceClientSets,
		secretsManagerConstructor(options.IntegrationNamespace),
		options.IntegrationNamespace,
		options.GatewayPort,
		options.UploadServiceUrl)
	if err != nil {
		log.Errorf("Failed to create synchronization service, %s", err.Error())
		os.Exit(1)
	}

	configMapNamespacedName := parseNamespacedName(options.ConnectionConfigMap)
	configMapClient := k8sResourceClientSets.core.CoreV1().ConfigMaps(configMapNamespacedName.Namespace)

	configProvider := confProvider.NewConfigProvider(configMapNamespacedName.Name, configMapClient)
	clientsProvider := compass.NewClientsProvider(graphql.New, options.InsecureConnectorCommunication, options.InsecureConfigurationFetch, options.QueryLogging)

	log.Infoln("Setting up Controller")
	controllerDependencies := compassconnection.DependencyConfig{
		K8sConfig:                    cfg,
		ControllerManager:            mgr,
		ClientsProvider:              clientsProvider,
		CredentialsManager:           certManager,
		SynchronizationService:       syncService,
		ConfigProvider:               configProvider,
		RuntimeURLsConfig:            options.Runtime,
		CertValidityRenewalThreshold: options.CertValidityRenewalThreshold,
		MinimalCompassSyncTime:       options.MinimalCompassSyncTime,
	}

	compassConnectionSupervisor, err := controllerDependencies.InitializeController()
	if err != nil {
		log.Error(err, "Unable to initialize Controller")
		os.Exit(1)
	}

	// Initialize Compass Connection CR
	log.Infoln("Initializing Compass Connection CR")
	_, err = compassConnectionSupervisor.InitializeCompassConnection()
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
