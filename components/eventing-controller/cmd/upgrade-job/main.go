package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"log"
	"os"
	"time"

	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/secret"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/subscription"

	eventmesh "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/event-mesh"
	eventingbackend "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/eventing-backend"
	jobprocess "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/process"
)

// Config for env variables
type Config struct {
	ReleaseName     		string `envconfig:"RELEASE" required:"true"`
	KymaNamespace    		string `envconfig:"KYMA_NAMESPACE" default:"kyma-system"`
	EventingControllerName  string `envconfig:"EVENTING_CONTROLLER_NAME" required:"true"`
	EventingPublisherName	string `envconfig:"EVENTING_PUBLISHER_NAME" required:"true"`
	LogFormat 				string `envconfig:"APP_LOG_FORMAT" default:"json"`
	LogLevel  				string `envconfig:"APP_LOG_LEVEL" default:"warn"`
}

func main() {
	// Read Env variables
	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatalf("Start handler failed with error: %s", err)
	}

	// Create logger instance
	ctrLogger, err := logger.New(cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %s", err)
	}
	defer func() {
		if err := ctrLogger.WithContext().Sync(); err != nil {
			log.Printf("failed to flush logger:: %s\n", err)
		}
	}()


	// Generate dynamic clients
	k8sConfig := config.GetConfigOrDie()

	// Create dynamic client
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)

	// setup clients
	deploymentClient := deployment.NewClient(dynamicClient)
	subscriptionClient := subscription.NewClient(dynamicClient)
	eventingBackendClient := eventingbackend.NewClient(dynamicClient)
	secretClient := secret.NewClient(dynamicClient)
	eventMeshClient := eventmesh.NewClient()

	// Create process
	p := jobprocess.Process{
		Logger: ctrLogger.Logger,
		TimeoutPeriod: 60 * time.Second,
		ReleaseName:  cfg.ReleaseName,
		KymaNamespace: cfg.KymaNamespace,
		ControllerName: cfg.EventingControllerName,
		PublisherName: cfg.EventingPublisherName,
		Clients: jobprocess.Clients{
			Deployment: deploymentClient,
			Subscription: subscriptionClient,
			EventingBackend: eventingBackendClient,
			Secret: secretClient,
			EventMesh: eventMeshClient,
		},
	}

	// First check if BEB is enabled for Kyma cluster
	checkBebJob := jobprocess.NewCheckIsBebEnabled(&p)
	err = checkBebJob.Do()
	if err != nil {
		ctrLogger.Logger.WithContext().Error(errors.Wrapf(err, "failed to check: %s", checkBebJob.ToString()))
		os.Exit(1)
	}

	// If BEB is not enabled then stop the upgrade-job
	// because we don't need this upgrade-job
	if !p.State.IsBebEnabled {
		ctrLogger.Logger.WithContext().Info("BEB not enabled for Kyma cluster! Exiting upgrade-job")
		return
	}

	// BEB is enabled, execute the steps for this upgrade-job
	// Add steps to process
	p.AddSteps()

	// Execute process
	err = p.Execute()
	if err != nil {
		ctrLogger.Logger.WithContext().Error(err)
	}

	ctrLogger.Logger.WithContext().Info("upgrade-job completed")
}
