package main

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/connector"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/director"
	config_provider "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/config"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/secrets"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/typed/compass/v1alpha1"
	"k8s.io/client-go/kubernetes"

	"os"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compassconnection"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	apis "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

func main() {
	log.Infoln("Starting Runtime Agent")

	var options Config
	err := envconfig.InitWithPrefix(&options, "APP") // TODO - refactor in chart
	if err != nil {
		log.Error("Failed to process environment variables")
	}
	log.Infof("Env config: %s", options)

	var connectionConfig EnvConfig
	err = envconfig.InitWithPrefix(&connectionConfig, "")
	if err != nil {
		log.Error("Failed to process environment variables for connection")
	}
	log.Infof("Connection config: %s", connectionConfig)

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
		log.Error(err, "Unable add APIs to scheme")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	compassConnectionCRClient, err := v1alpha1.NewForConfig(cfg)
	if err != nil {
		log.Error("Unable to setup Compass Connection CR client")
		os.Exit(1)
	}

	// TODO - rework this part
	coreClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Errorf("Failed to initialize core clientset: %s", err.Error())
		os.Exit(1)
	}

	coreClientSet := coreClientset.CoreV1()

	// TODO - one secret repo instead of 2
	secretsRepository := secrets.NewRepository(func(namespace string) secrets.Manager {
		return coreClientSet.Secrets(namespace)
	})

	clusterCertSecret := parseNamespacedName(options.ClusterCertificatesSecret)
	caCertSecret := parseNamespacedName(options.CaCertificatesSecret)

	certManager := certificates.NewCredentialsManager(clusterCertSecret, caCertSecret, secretsRepository)
	compassConfigClient := director.NewConfigurationClient(graphql.New, options.InsecureConfigurationFetch)
	syncService, err := createNewSynchronizationService(cfg, options.IntegrationNamespace, options.GatewayPort, options.UploadServiceUrl)
	if err != nil {
		log.Errorf("Failed to create synchronization service, %s", err.Error())
		os.Exit(1)
	}

	configProvider := config_provider.NewConfigProvider(options.ConfigFile)

	clientsProvider := compass.NewClientsProvider(graphql.New, options.InsecureConnectorCommunication, options.InsecureConfigurationFetch, options.QueryLogging)

	compassConnector := newCompassConnector(options.InsecureConnectorCommunication)
	connectionSupervisor := compassconnection.NewSupervisor(
		compassConnector,
		compassConnectionCRClient.CompassConnections(),
		certManager,
		compassConfigClient,
		syncService,
		configProvider,
		options.CertValidityRenewalThreshold,
		options.MinimalCompassSyncTime)

	// Setup all Controllers
	log.Info("Setting up controller")
	if err := compassconnection.InitCompassConnectionController(mgr, connectionSupervisor, options.MinimalCompassSyncTime); err != nil {
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

func newCompassConnector(insecureConnection bool) compassconnection.Connector {
	csrProvider := certificates.NewCSRProvider()
	return compassconnection.NewCompassConnector(
		csrProvider,
		connector.NewTokenSecuredConnectorClient,
		connector.NewConnectorClient,
		insecureConnection)
}
