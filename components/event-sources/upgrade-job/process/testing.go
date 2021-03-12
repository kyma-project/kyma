package process

import (
	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	eventsourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
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
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/processtest"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
)

type E2ESetup struct {
	applications     *applicationconnectorv1alpha1.ApplicationList
	validators       *appsv1.DeploymentList
	eventServices    *appsv1.DeploymentList
	httpSources      *eventsourcesv1alpha1.HTTPSourceList
	appSubscriptions *messagingv1alpha1.SubscriptionList
	channels         *messagingv1alpha1.ChannelList
	brokers          *eventingv1alpha1.BrokerList
	triggers         *eventingv1alpha1.TriggerList
	namespaces       *corev1.NamespaceList
}

func getProcessClients(e2eSetup E2ESetup, g *gomega.GomegaWithT) Clients {
	fakeTriggerClient, err := trigger.NewFakeClient(e2eSetup.triggers)
	g.Expect(err).Should(gomega.BeNil())
	fakeBrokerClient, err := broker.NewFakeClient(e2eSetup.brokers)
	g.Expect(err).Should(gomega.BeNil())
	fakeSubClient, err := subscription.NewFakeClient(e2eSetup.appSubscriptions)
	g.Expect(err).Should(gomega.BeNil())
	fakeAppClient, err := application.NewFakeClient(e2eSetup.applications)
	g.Expect(err).Should(gomega.BeNil())
	fakeChannelClient, err := channel.NewFakeClient(e2eSetup.channels)
	g.Expect(err).Should(gomega.BeNil())
	fakeNsClient, err := namespace.NewFakeClient(e2eSetup.namespaces)
	g.Expect(err).Should(gomega.BeNil())
	fakeDeploymentClient, err := deployment.NewFakeClient(combineValidatorsAndEventServices(e2eSetup.validators, e2eSetup.eventServices))
	g.Expect(err).Should(gomega.BeNil())
	fakeHTTPSourceClient, err := httpsource.NewFakeClient(e2eSetup.httpSources)
	g.Expect(err).Should(gomega.BeNil())
	fakeConfigMapClient, err := configmap.NewFakeClient(nil)
	g.Expect(err).Should(gomega.BeNil())
	fakeKymaSubClient, err := kymasubscription.NewFakeClient(nil)
	g.Expect(err).Should(gomega.BeNil())
	return Clients{
		Trigger:          fakeTriggerClient,
		Broker:           fakeBrokerClient,
		Subscription:     fakeSubClient,
		Channel:          fakeChannelClient,
		Application:      fakeAppClient,
		Deployment:       fakeDeploymentClient,
		ConfigMap:        fakeConfigMapClient,
		HttpSource:       fakeHTTPSourceClient,
		Namespace:        fakeNsClient,
		KymaSubscription: fakeKymaSubClient,
	}
}

func newE2ESetup() E2ESetup {
	newApps := processtest.NewApps()
	newValidators := processtest.NewValidators()
	newEventServices := processtest.NewEventServices()
	newTriggers := processtest.NewTriggers()
	newHTTPSources := processtest.NewHTTPSources()
	newAppChannels := processtest.NewAppChannels()
	newAppSubscriptions := processtest.NewAppSubscriptions()
	newBrokers := processtest.NewBrokers()
	newNamespaces := processtest.NewNamespaces()

	e2eSetup := E2ESetup{
		applications:     newApps,
		validators:       newValidators,
		eventServices:    newEventServices,
		triggers:         newTriggers,
		httpSources:      newHTTPSources,
		channels:         newAppChannels,
		appSubscriptions: newAppSubscriptions,
		brokers:          newBrokers,
		namespaces:       newNamespaces,
	}
	return e2eSetup
}

func combineValidatorsAndEventServices(validators, eventServices *appsv1.DeploymentList) *appsv1.DeploymentList {
	result := new(appsv1.DeploymentList)
	for _, v := range validators.Items {
		result.Items = append(result.Items, v)
	}
	for _, es := range eventServices.Items {
		result.Items = append(result.Items, es)
	}
	return result
}
