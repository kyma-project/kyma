package main

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/cache"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/director"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compassconnection"
	confProvider "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/config"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/healthz"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/secrets"
	apis "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
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

	log.Info("Migrating certificate if needed")
	k8sResourceClientSets, err := k8sResourceClients(cfg)
	exitOnError(err, "Failed to initialize K8s resource clients")

	secretsManagerConstructor := func(namespace string) secrets.Manager {
		return k8sResourceClientSets.core.CoreV1().Secrets(namespace)
	}

	caCertSecret := parseNamespacedName(options.CaCertificatesSecret)
	caCertSecretToMigrate := parseNamespacedName(options.CaCertSecretToMigrate)

	secretsRepository := secrets.NewRepository(secretsManagerConstructor)

	err = migrateSecret(secretsRepository, caCertSecretToMigrate, caCertSecret, options.CaCertSecretKeysToMigrate)
	exitOnError(err, "Failed to migrate ")

	log.Info("Migrating credentials if needed")
	clusterCertSecret := parseNamespacedName(options.ClusterCertificatesSecret)
	agentConfigSecret := parseNamespacedName(options.AgentConfigurationSecret)
	oldClusterCertSecret := parseNamespacedName(options.ClusterCertificatesSecretToMigrate)
	oldAgentConfigSecret := parseNamespacedName(options.AgentConfigurationSecretToMigrate)

	err = migrateSecretAllKeys(secretsRepository, oldClusterCertSecret, clusterCertSecret)
	exitOnError(err, "Failed to migrate ")

	err = migrateSecretAllKeys(secretsRepository, oldAgentConfigSecret, agentConfigSecret)
	exitOnError(err, "Failed to migrate ")

	log.Info("Setting up manager")
	mgr, err := manager.New(cfg, manager.Options{SyncPeriod: &options.ControllerSyncPeriod})
	exitOnError(err, "Failed to set up overall controller manager")

	// Setup Scheme for all resources
	log.Info("Setting up scheme")
	err = apis.AddToScheme(mgr.GetScheme())
	exitOnError(err, "Failed to add APIs to scheme")

	log.Info("Registering Components.")

	certManager := certificates.NewCredentialsManager(clusterCertSecret, caCertSecret, secretsRepository)

	syncService, err := createSynchronisationService(k8sResourceClientSets, options)
	exitOnError(err, "Failed to create synchronization service")

	connectionDataCache := cache.NewConnectionDataCache()

	configProvider := confProvider.NewConfigProvider(agentConfigSecret, secretsRepository)
	clientsProvider := compass.NewClientsProvider(graphql.New, options.SkipCompassTLSVerify, options.QueryLogging)
	connectionDataCache.AddSubscriber(clientsProvider.UpdateConnectionData)

	log.Infoln("Setting up Director Proxy Service")
	directorProxy := director.NewProxy(options.DirectorProxy)
	err = mgr.Add(directorProxy)
	exitOnError(err, "Failed to create director proxy")
	connectionDataCache.AddSubscriber(directorProxy.SetURLAndCerts)

	log.Infoln("Setting up Controller")
	controllerDependencies := compassconnection.DependencyConfig{
		K8sConfig:                    cfg,
		ControllerManager:            mgr,
		ClientsProvider:              clientsProvider,
		CredentialsManager:           certManager,
		SynchronizationService:       syncService,
		ConfigProvider:               configProvider,
		ConnectionDataCache:          connectionDataCache,
		RuntimeURLsConfig:            options.Runtime,
		CertValidityRenewalThreshold: options.CertValidityRenewalThreshold,
		MinimalCompassSyncTime:       options.MinimalCompassSyncTime,
	}

	compassConnectionSupervisor, err := controllerDependencies.InitializeController()
	exitOnError(err, "Failed to initialize Controller")

	correlationID := uuid.New().String()
	ctx := correlation.SaveCorrelationIDHeaderToContext(context.Background(), str.Ptr(correlation.RequestIDHeaderKey), str.Ptr(correlationID))

	log.Infoln("Initializing Compass Connection CR")
	_, err = compassConnectionSupervisor.InitializeCompassConnection(ctx)
	exitOnError(err, "Failed to initialize Compass Connection CR")

	log.Infoln("Initializing metrics logger")
	metricsLogger, err := newMetricsLogger(options.MetricsLoggingTimeInterval)
	exitOnError(err, "Failed to create metrics logger")
	err = mgr.Add(metricsLogger)
	exitOnError(err, "Failed to add metrics logger to manager")

	go func() {
		log.Info("Starting Healthcheck Server")
		healthz.StartHealthCheckServer(log.StandardLogger(), options.HealthPort)
	}()

	log.Info("Starting the Cmd.")
	err = mgr.Start(signals.SetupSignalHandler())
	exitOnError(err, "Failed to run the manager")
}

func migrateSecretAllKeys(secretRepo secrets.Repository, sourceSecret, targetSecret types.NamespacedName) error {

	includeAllKeysFunc := func(k string) bool {
		return true
	}

	migrator := certificates.NewMigrator(secretRepo, includeAllKeysFunc)
	return migrator.Do(sourceSecret, targetSecret)
}

func migrateSecret(secretRepo secrets.Repository, sourceSecret, targetSecret types.NamespacedName, keysToInclude string) error {
	unmarshallKeysList := func(keys string) (keysArray []string, err error) {
		err = json.Unmarshal([]byte(keys), &keysArray)

		return keysArray, err
	}

	keys, err := unmarshallKeysList(keysToInclude)
	if err != nil {
		log.Errorf("Failed to read secret keys to be migrated")
		return err
	}

	migrator := getMigrator(secretRepo, keys)

	return migrator.Do(sourceSecret, targetSecret)
}

func getMigrator(secretRepo secrets.Repository, keysToInclude []string) certificates.Migrator {
	getIncludeSourceKeyFunc := func() certificates.IncludeKeyFunc {
		if len(keysToInclude) == 0 {
			return func(string) bool {
				return true
			}
		}

		return func(key string) bool {
			for _, k := range keysToInclude {
				if k == key {
					return true
				}
			}

			return false
		}
	}

	return certificates.NewMigrator(secretRepo, getIncludeSourceKeyFunc())
}

func createSynchronisationService(k8sResourceClients *k8sResourceClientSets, options Config) (kyma.Service, error) {

	var syncService kyma.Service
	var err error

	syncService, err = createKymaService(k8sResourceClients, options.IntegrationNamespace, options.CentralGatewayServiceUrl, options.SkipAppsTLSVerify)

	if err != nil {
		return nil, err
	}

	return syncService, nil
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatal(errors.Wrap(err, context))
	}
}
