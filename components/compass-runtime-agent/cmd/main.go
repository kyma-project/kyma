package main

import (
	"kyma-project.io/compass-runtime-agent/internal/certificates"
	"kyma-project.io/compass-runtime-agent/internal/compass"
	"kyma-project.io/compass-runtime-agent/internal/compass/director"
	"kyma-project.io/compass-runtime-agent/internal/compassconnection"
	confProvider "kyma-project.io/compass-runtime-agent/internal/config"
	"kyma-project.io/compass-runtime-agent/internal/graphql"
	"kyma-project.io/compass-runtime-agent/internal/secrets"
	apis "kyma-project.io/compass-runtime-agent/pkg/apis/compass/v1alpha1"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	log.Infoln("Starting Runtime Agent")

	var options Config
	err := envconfig.InitWithPrefix(&options, "APP")
	exitOnError(err, "Failed to process environment variables")

	log.Infof("Env config: %s", options.String())

	// Get a config to talk to the apiserver
	log.Info("Setting up client for manager")
	cfg, err := config.GetConfig()
	exitOnError(err, "Failed to set up client config")

	log.Info("Setting up manager")
	mgr, err := manager.New(cfg, manager.Options{SyncPeriod: &options.ControllerSyncPeriod})
	exitOnError(err, "Failed to set up overall controller manager")

	// Setup Scheme for all resources
	log.Info("Setting up scheme")
	err = apis.AddToScheme(mgr.GetScheme())
	exitOnError(err, "Failed to add APIs to scheme")

	log.Info("Registering Components.")

	k8sResourceClientSets, err := k8sResourceClients(cfg)
	exitOnError(err, "Failed to initialize K8s resource clients")

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
	exitOnError(err, "Failed to create synchronization service")

	configMapNamespacedName := parseNamespacedName(options.ConnectionConfigMap)
	configMapClient := k8sResourceClientSets.core.CoreV1().ConfigMaps(configMapNamespacedName.Namespace)

	configProvider := confProvider.NewConfigProvider(configMapNamespacedName.Name, configMapClient)
	clientsProvider := compass.NewClientsProvider(graphql.New, options.InsecureConnectorCommunication, options.InsecureConfigurationFetch, options.QueryLogging)

	log.Infoln("Setting up Director Proxy Service")
	directorProxy := director.NewProxy(options.DirectorProxy)
	err = mgr.Add(directorProxy)
	exitOnError(err, "Failed to create director proxy")

	log.Infoln("Setting up Controller")
	controllerDependencies := compassconnection.DependencyConfig{
		K8sConfig:                    cfg,
		ControllerManager:            mgr,
		ClientsProvider:              clientsProvider,
		CredentialsManager:           certManager,
		SynchronizationService:       syncService,
		ConfigProvider:               configProvider,
		DirectorProxyUpdater:         directorProxy,
		RuntimeURLsConfig:            options.Runtime,
		CertValidityRenewalThreshold: options.CertValidityRenewalThreshold,
		MinimalCompassSyncTime:       options.MinimalCompassSyncTime,
	}

	compassConnectionSupervisor, err := controllerDependencies.InitializeController()
	exitOnError(err, "Failed to initialize Controller")

	log.Infoln("Initializing Compass Connection CR")
	_, err = compassConnectionSupervisor.InitializeCompassConnection()
	exitOnError(err, "Failed to initialize Compass Connection CR")

	log.Infoln("Initializing metrics logger")
	metricsLogger, err := newMetricsLogger(options.MetricsLoggingTimeInterval)
	exitOnError(err, "Failed to create metrics logger")
	err = mgr.Add(metricsLogger)

	log.Info("Starting the Cmd.")
	err = mgr.Start(signals.SetupSignalHandler())
	exitOnError(err, "Failed to run the manager")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatal(errors.Wrap(err, context))
	}
}
