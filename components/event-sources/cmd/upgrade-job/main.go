package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/application"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/broker"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/channel"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/configmap"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/deployment"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/httpsource"
	kymasubscription "github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/kyma-subscription"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/namespace"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/subscription"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/trigger"
	jobprocess "github.com/kyma-project/kyma/components/event-sources/upgrade-job/process"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Config struct {
	ReleaseName     string `envconfig:"RELEASE" required:"true"`
	BEBNamespace    string `envconfig:"BEB_NAMESPACE" default:""`
	EventingBackend string `envconfig:"EVENTING_BACKEND" required:"true"`
	EventTypePrefix string `envconfig:"EVENT_TYPE_PREFIX" required:"true"`
}

func main() {

	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	// Env vars
	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		logger.Fatalf("Start handler failed with error: %s", err)
	}

	// Generate dynamic clients
	k8sConfig := config.GetConfigOrDie()

	// setup dynamic clients
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)

	triggerClient := trigger.NewClient(dynamicClient)
	subscriptionClient := subscription.NewClient(dynamicClient)
	brokerClient := broker.NewClient(dynamicClient)
	deploymentClient := deployment.NewClient(dynamicClient)
	applicationClient := application.NewClient(dynamicClient)
	configMapClient := configmap.NewClient(dynamicClient)
	namespaceClient := namespace.NewClient(dynamicClient)
	httpSourceClient := httpsource.NewClient(dynamicClient)
	channelClient := channel.NewClient(dynamicClient)
	kymaSubClient := kymasubscription.NewClient(dynamicClient)

	// Create process
	p := jobprocess.Process{
		ReleaseName:     cfg.ReleaseName,
		BEBNamespace:    cfg.BEBNamespace,
		EventingBackend: cfg.EventingBackend,
		EventTypePrefix: cfg.EventTypePrefix,
		Clients: jobprocess.Clients{
			Trigger:          triggerClient,
			Subscription:     subscriptionClient,
			Application:      applicationClient,
			Deployment:       deploymentClient,
			ConfigMap:        configMapClient,
			Channel:          channelClient,
			Broker:           brokerClient,
			HttpSource:       httpSourceClient,
			Namespace:        namespaceClient,
			KymaSubscription: kymaSubClient,
		},
		Logger: logger,
	}

	// Add steps to process
	p.AddSteps()

	err := p.Execute()
	if err != nil {
		logger.Fatalf("Upgrade process failed: %v", err)
	}

	logger.Info("Upgrade process completed successfully!")
}
