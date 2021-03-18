package eventmesh

import (
	"github.com/sirupsen/logrus"

	scclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	"k8s.io/client-go/kubernetes"

	appbrokerclientset "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appconnectorclientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"

	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/runner"

	eventingv1alpha1clientset "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	messagingv1alpha1clientset "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
)

// TODO remove this test when Kyma is fully integrated with the new Eventing solution https://github.com/kyma-project/kyma/issues/10866
type EventMeshUpgradeTest struct {
	k8sInterface kubernetes.Interface

	appConnectorInterface appconnectorclientset.Interface
	messagingClient       messagingv1alpha1clientset.MessagingV1alpha1Interface
	sourcesClient         sourcesv1alpha1.SourcesV1alpha1Interface
	appBrokerCli          appbrokerclientset.Interface
	scCli                 scclientset.Interface
	eventingCli           eventingv1alpha1clientset.EventingV1alpha1Interface
	subscriberImage       string
	skipTestResources     bool
}

// compile time assertion
var _ runner.UpgradeTest = &EventMeshUpgradeTest{}

func NewEventMeshUpgradeTest(
	appConnectorCli appconnectorclientset.Interface,
	k8sCli kubernetes.Interface,
	messagingCli messagingv1alpha1clientset.MessagingV1alpha1Interface,
	sourcesCli sourcesv1alpha1.SourcesV1alpha1Interface,
	appBrokerCli appbrokerclientset.Interface,
	scCli scclientset.Interface,
	eventingCli eventingv1alpha1clientset.EventingV1alpha1Interface,
	subscriberImage string) runner.UpgradeTest {
	return &EventMeshUpgradeTest{
		k8sInterface:          k8sCli,
		messagingClient:       messagingCli,
		appConnectorInterface: appConnectorCli,
		sourcesClient:         sourcesCli,
		appBrokerCli:          appBrokerCli,
		scCli:                 scCli,
		eventingCli:           eventingCli,
		subscriberImage:       subscriberImage,
		skipTestResources:     true,
	}
}

func (e *EventMeshUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	f := newEventMeshFlow(e, stop, log, namespace)

	for _, fn := range []func() error{
		f.CreateApplication,
		f.CreateSubscriber,
		f.WaitForApplication,
		f.WaitForSubscriber,
		f.CreateApplicationMapping,
		f.CreateServiceInstance,
		f.WaitForServiceInstance,
		f.CreateTrigger,
		f.WaitForTrigger,
	} {
		err := fn()
		if err != nil {
			f.log.WithField("error", err).Error("CreateResources() failed")
			return err
		}
	}

	return nil
}

func (e *EventMeshUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	if e.skipTestResources {
		log.Info("TestResources for EventMeshUpgradeTest is skipped")
		return nil
	}

	f := newEventMeshFlow(e, stop, log, namespace)

	for _, fn := range []func() error{
		// Steps to test:
		// Check subscriber is ready or not
		// Check readiness for Brokers
		// Check readiness for Triggers
		// Check readiness for EventActivation
		// Publish an event to the event service
		// Check event reached subscriber
		f.WaitForApplication,
		f.WaitForSubscriber,
		f.WaitForServiceInstance,
		f.WaitForBroker,
		f.WaitForTrigger,
		f.PublishTestEvent,
		f.CheckEvent,
	} {
		err := fn()
		if err != nil {
			// f.log.WithField("error", err).Error("TestResources() failed")
			return err
		}
	}

	return nil
}
