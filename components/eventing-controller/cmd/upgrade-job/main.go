package main

import (
	"github.com/kelseyhightower/envconfig"

	"log"
	"os"
	"time"

	"github.com/pkg/errors"

	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/configmap"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/eventingbackend"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/eventmesh"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/secret"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/subscription"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/env"
	jobprocess "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/process"
)

func main() {
	// Read Env variables
	cfg := new(env.Config)
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
	configMapClient := configmap.NewClient(dynamicClient)
	eventMeshClient := eventmesh.NewClient()

	// Create process
	p := jobprocess.Process{
		Logger:         ctrLogger.Logger,
		TimeoutPeriod:  60 * time.Second,
		ReleaseName:    cfg.ReleaseName,
		KymaNamespace:  cfg.KymaNamespace,
		ControllerName: cfg.EventingControllerName,
		PublisherName:  cfg.EventingPublisherName,
		Clients: jobprocess.Clients{
			Deployment:      deploymentClient,
			Subscription:    subscriptionClient,
			EventingBackend: eventingBackendClient,
			Secret:          secretClient,
			ConfigMap:       configMapClient,
			EventMesh:       eventMeshClient,
		},
		State: jobprocess.State{},
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
