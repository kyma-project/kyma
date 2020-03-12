package migrateeventmesh

import (
	appconnector "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	serving "knative.dev/serving/pkg/client/clientset/versioned"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	appBroker "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	ebClientSet "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/runner"
	"github.com/sirupsen/logrus"
	eventingclientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	messagingclientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
)

type MigrateFromEventMeshUpgradeTest struct {
	k8sInterface kubernetes.Interface

	appConnectorInterface appconnector.Interface
	messagingClient       messagingclientv1alpha1.MessagingV1alpha1Interface
	servingClient         serving.Interface
	appBrokerCli          appBroker.Interface
	scCli                 sc.Interface
	eventingCli           eventingclientv1alpha1.EventingV1alpha1Interface
	ebCli                 ebClientSet.Interface
}

// compile time assertion
var _ runner.UpgradeTest = &MigrateFromEventMeshUpgradeTest{}

func NewMigrateFromEventMeshUpgradeTest(appConnectorInterface appconnector.Interface, k8sInterface kubernetes.Interface, messagingCli messagingclientv1alpha1.MessagingV1alpha1Interface, servingCli serving.Interface, appBrokerCli appBroker.Interface, scCli sc.Interface, eventingCli eventingclientv1alpha1.EventingV1alpha1Interface, ebCli ebClientSet.Interface) runner.UpgradeTest {
	return &MigrateFromEventMeshUpgradeTest{
		k8sInterface:          k8sInterface,
		messagingClient:       messagingCli,
		appConnectorInterface: appConnectorInterface,
		servingClient:         servingCli,
		appBrokerCli:          appBrokerCli,
		scCli:                 scCli,
		eventingCli:           eventingCli,
		ebCli:                 ebCli,
	}
}

func (e *MigrateFromEventMeshUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	f := newMigrateEventMeshFlow(e, stop, log, namespace)

	for _, fn := range []func() error{
		f.CreateApplication,
		f.CreateSubscriber,
		f.WaitForApplication,
		f.WaitForSubscriber,
		f.CreateApplicationMapping,
		f.CreateServiceInstance,
		f.WaitForServiceInstance,
		f.CreateSubscription,
		f.CheckSubscriptionReady,
		//f.WaitForBroker,
		//f.CreateTrigger,
		//f.WaitForTrigger,
		//f.publishTestEvent,
		//f.checkSubscriberReceivedEvent,
	} {
		err := fn()
		if err != nil {
			f.log.WithField("error", err).Error("CreateResources() failed")
			return err
		}
	}

	return nil
}

func (e *MigrateFromEventMeshUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	f := newMigrateEventMeshFlow(e, stop, log, namespace)

	for _, fn := range []func() error{
		// Steps to test:
		// Check subscriber is ready or not
		// Check readiness for Brokers
		// Check readiness for Triggers
		// Check readiness for EventActivation
		// Publish an event to the event service
		// Check event reached subscriber
		// Clean up stuff e.g. subscriber, trigger, eventactivation, broker (optional)
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
			//f.log.WithField("error", err).Error("TestResources() failed")
			return err
		}
	}

	return nil
}
