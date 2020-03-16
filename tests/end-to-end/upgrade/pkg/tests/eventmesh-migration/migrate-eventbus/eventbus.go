package migrateeventbus

import (
	"k8s.io/client-go/kubernetes"

	"github.com/sirupsen/logrus"

	scclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	appbrokerclientset "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appconnector "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	ebclientset "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/runner"

	eventingv1alpha1clientset "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	messagingv1alpha1clientset "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
	servingclientset "knative.dev/serving/pkg/client/clientset/versioned"
)

type MigrateFromEventBusUpgradeTest struct {
	k8sInf kubernetes.Interface

	appConnectorInf appconnector.Interface
	messagingInf    messagingv1alpha1clientset.MessagingV1alpha1Interface
	servingInf      servingclientset.Interface
	appBrokerInf    appbrokerclientset.Interface
	scInf           scclientset.Interface
	eventingInf     eventingv1alpha1clientset.EventingV1alpha1Interface
	ebInf           ebclientset.Interface
}

// compile time assertion
var _ runner.UpgradeTest = &MigrateFromEventBusUpgradeTest{}

func NewMigrateFromEventBusUpgradeTest(appConnectorInf appconnector.Interface, k8sInf kubernetes.Interface, messagingInf messagingv1alpha1clientset.MessagingV1alpha1Interface, servingInf servingclientset.Interface, appBrokerInf appbrokerclientset.Interface, scInf scclientset.Interface, eventingInf eventingv1alpha1clientset.EventingV1alpha1Interface, ebInf ebclientset.Interface) runner.UpgradeTest {
	return &MigrateFromEventBusUpgradeTest{
		k8sInf:          k8sInf,
		messagingInf:    messagingInf,
		appConnectorInf: appConnectorInf,
		servingInf:      servingInf,
		appBrokerInf:    appBrokerInf,
		scInf:           scInf,
		eventingInf:     eventingInf,
		ebInf:           ebInf,
	}
}

func (e *MigrateFromEventBusUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	f := newMigrateEventBusFlow(e, stop, log, namespace)

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
		f.WaitForBroker,
		// To simulate the upgrade of namespaces without brokers to Kyma 1.12(EventMesh enabled), brokers need to be removed
		f.RemoveBrokerInjectionLabel,
		f.DeleteBroker,
	} {
		err := fn()
		if err != nil {
			f.log.WithField("error", err).Error("CreateResources() failed")
			return err
		}
	}

	return nil
}

func (e *MigrateFromEventBusUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	f := newMigrateEventBusFlow(e, stop, log, namespace)

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
			f.log.WithField("error", err).Error("TestResources() failed")
			return err
		}
	}

	return nil
}
