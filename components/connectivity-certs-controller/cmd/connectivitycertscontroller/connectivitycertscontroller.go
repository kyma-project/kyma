package main

import (
	"os"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/centralconnection"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/connectorservice"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificaterequest"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/pkg/apis/applicationconnector/v1alpha1"

	"github.com/pkg/errors"

	"k8s.io/client-go/kubernetes"

	log "github.com/sirupsen/logrus"
	restclient "k8s.io/client-go/rest"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {

	log.Info("Starting Certificates Manager.")

	options := parseArgs()
	log.Infof("Options: %s", options)

	// Get a config to talk to the apiserver
	log.Info("Setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "Unable to set up client config")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	log.Info("Setting up manager")
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		log.Error(err, "Unable to set up overall controller manager")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	secretsRepository, err := newSecretsRepository(options.namespace)
	if err != nil {
		log.Errorf("Failed to initialize secret repository, %s", err.Error())
		return
	}

	// Setup Scheme for all resources
	log.Info("Setting up scheme")
	if err := v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "Unable add APIs to scheme")
		os.Exit(1)
	}

	csrProvider := certificates.NewCSRProvider(options.clusterCertificatesSecret, options.caCertificatesSecret, secretsRepository)
	certPreserver := certificates.NewCertificatePreserver(options.clusterCertificatesSecret, options.caCertificatesSecret, secretsRepository)
	connectorClient := connectorservice.NewConnectorClient(csrProvider)

	// Setup Certificate Request Controller
	log.Info("Setting up Certificate Request controller")
	if err := certificaterequest.InitCertificatesRequestController(mgr, options.appName, connectorClient, certPreserver); err != nil {
		log.Error(err, "Unable to register controllers to the manager")
		os.Exit(1)
	}

	// Setup Master Connection Controller
	log.Info("Setting up Master Connection controller")
	if err := centralconnection.InitMasterConnectionsController(mgr, options.appName, connectorClient, certPreserver); err != nil {
		log.Error(err, "Unable to register controllers to the manager")
		os.Exit(1)
	}

	// Start the Cmd
	log.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Unable to run the manager")
		os.Exit(1)
	}
}

func newSecretsRepository(namespace string) (secrets.Repository, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read in cluster Kubernetes config")
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize core clientset")
	}

	sei := coreClientset.CoreV1().Secrets(namespace)

	return secrets.NewRepository(sei), nil
}
