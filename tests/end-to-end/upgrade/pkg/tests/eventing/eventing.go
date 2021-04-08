package eventing

import (
	"fmt"
	"net/http"

	scclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	appbrokerclientset "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appconnectorclientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/runner"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/eventing/helpers"
)

const (
	kymaSystemNamespace     = "kyma-system"
	publisherServiceName    = "eventing-event-publisher-proxy"
	defaultName             = "eventupgrade"
	defaultEventType        = "order.created"
	defaultEventTypeVersion = "v1"
)

var _ runner.UpgradeTest = &EventingUpgradeTest{}

type step func() error

type EventingUpgradeTest struct {
	k8sClient          kubernetes.Interface
	dynamicClient      dynamic.Interface
	appConnectorClient appconnectorclientset.Interface
	appBrokerClient    appbrokerclientset.Interface
	scClient           scclientset.Interface
	subscriberImage    string

	// test config
	namespace           string
	eventType           string
	subscriberName      string
	applicationName     string
	eventTypeVersion    string
	subscriptionName    string
	serviceInstanceName string
}

func NewEventingUpgradeTest(k8sClient kubernetes.Interface, appConnectorCli appconnectorclientset.Interface, appBrokerClient appbrokerclientset.Interface, scClient scclientset.Interface, dynamicClient dynamic.Interface, subscriberImage string) *EventingUpgradeTest {
	return &EventingUpgradeTest{
		k8sClient:           k8sClient,
		dynamicClient:       dynamicClient,
		appConnectorClient:  appConnectorCli,
		appBrokerClient:     appBrokerClient,
		scClient:            scClient,
		subscriberImage:     subscriberImage,
		eventTypeVersion:    defaultEventTypeVersion,
		eventType:           defaultEventType,
		subscriberName:      defaultName,
		applicationName:     defaultName,
		subscriptionName:    defaultName,
		serviceInstanceName: defaultName,
	}
}

func (e *EventingUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	for _, fn := range []func() error{
		e.createApplication,
		e.createSubscriber,
		e.waitForApplication,
		e.waitForSubscriber,
		e.createApplicationMapping,
		e.createServiceInstance,
		e.waitForServiceInstance,
		e.createSubscription,
		e.waitForSubscriptionReady,
	} {
		err := fn()
		if err != nil {
			log.WithField("error", err).Error("CreateResources() failed")
			return err
		}
	}
	return nil
}

func (e *EventingUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	// set the target namespace for the test with the same namespace that the EventMeshUpgradeTest used
	e.withNamespace(namespace)

	// prepare steps
	steps := []step{
		e.waitForSubscriptionReady,
		e.waitForSubscriber,
		e.publishTestEvent,
		e.checkEvent,
	}

	// execute steps
	for _, s := range steps {
		if err := s(); err != nil {
			log.WithField("error", err).Error("TestResources() failed")
			return err
		}
	}

	return nil
}

func (e *EventingUpgradeTest) createApplication() error {
	return helpers.CreateApplication(e.appConnectorClient, e.applicationName,
		helpers.WithAccessLabel(e.applicationName),
		helpers.WithEventService(e.serviceInstanceName),
	)
}

func (e *EventingUpgradeTest) waitForApplication() error {
	return helpers.WaitForApplication(e.appConnectorClient, e.applicationName)
}

func (e *EventingUpgradeTest) createSubscriber() error {
	return helpers.CreateSubscriber(e.k8sClient, e.subscriberName, e.namespace, helpers.WithSubscriberImage(e.subscriberImage))
}

func (e *EventingUpgradeTest) createApplicationMapping() error {
	return helpers.CreateApplicationMapping(e.appBrokerClient, e.applicationName, e.namespace)
}

func (e *EventingUpgradeTest) createServiceInstance() error {
	return helpers.CreateServiceInstance(e.scClient, e.serviceInstanceName, e.namespace)
}

func (e *EventingUpgradeTest) waitForServiceInstance() error {
	return helpers.WaitForServiceInstance(e.scClient, e.serviceInstanceName, e.namespace)
}

func (e *EventingUpgradeTest) createSubscription() error {
	return helpers.CreateSubscription(e.dynamicClient, e.subscriptionName, e.namespace,
		helpers.WithFilter(e.eventTypeVersion, e.eventType, e.applicationName),
		helpers.WithSink(fmt.Sprintf("http://%s.%s.svc.cluster.local:9000/ce", e.subscriberName, e.namespace)))

}

func (e *EventingUpgradeTest) withNamespace(namespace string) {
	e.namespace = namespace
}

func (e EventingUpgradeTest) waitForSubscriptionReady() error {
	return helpers.WaitForSubscriptionReady(e.dynamicClient, e.subscriptionName, e.namespace)
}

func (e EventingUpgradeTest) waitForSubscriber() error {
	return helpers.WaitForSubscriber(e.k8sClient, e.subscriberName, e.namespace)
}

func (e EventingUpgradeTest) publishTestEvent() error {
	return helpers.SendEvent(fmt.Sprintf("http://%s.%s/%s/v1/events", publisherServiceName, kymaSystemNamespace, e.applicationName), e.eventType, e.eventTypeVersion)
}

func (e EventingUpgradeTest) checkEvent() error {
	if err := helpers.CheckEvent(fmt.Sprintf("http://%s.%s.svc.cluster.local:9000/ce/%v/%v/%v",
		e.subscriberName, e.namespace, e.applicationName, e.eventType, e.eventTypeVersion), http.StatusNoContent); err != nil {
		return err
	}

	return nil
}
