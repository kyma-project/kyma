package migrateeventmesh

import (
	"github.com/sirupsen/logrus"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	"k8s.io/client-go/kubernetes"

	appbrokerclientset "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appconnector "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	ebclientset "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"

	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/runner"

	eventingclientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	messagingclientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
	serving "knative.dev/serving/pkg/client/clientset/versioned"
)

type MigrateFromEventMeshUpgradeTest struct {
	k8sInterface kubernetes.Interface

	appConnectorInterface appconnector.Interface
	messagingClient       messagingclientv1alpha1.MessagingV1alpha1Interface
	servingClient         serving.Interface
	appBrokerCli          appbrokerclientset.Interface
	scCli                 sc.Interface
	eventingCli           eventingclientv1alpha1.EventingV1alpha1Interface
	ebCli                 ebclientset.Interface
}

// compile time assertion
var _ runner.UpgradeTest = &MigrateFromEventMeshUpgradeTest{}

func NewMigrateFromEventMeshUpgradeTest(appConnectorCli appconnector.Interface, k8sCli kubernetes.Interface, messagingCli messagingclientv1alpha1.MessagingV1alpha1Interface, servingCli serving.Interface, appBrokerCli appbrokerclientset.Interface, scCli sc.Interface, eventingCli eventingclientv1alpha1.EventingV1alpha1Interface, ebCli ebclientset.Interface) runner.UpgradeTest {
	return &MigrateFromEventMeshUpgradeTest{
		k8sInterface:          k8sCli,
		messagingClient:       messagingCli,
		appConnectorInterface: appConnectorCli,
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
