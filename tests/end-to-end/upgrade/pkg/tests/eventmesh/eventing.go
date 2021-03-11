package eventmesh

import (
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/runner"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/eventmesh/helpers"
)

const (
	kymaSystemNamespace  = "kyma-system"
	publisherServiceName = "eventing-event-publisher-proxy"
)

var _ runner.UpgradeTest = &EventingUpgradeTest{}

type step func() error

type EventingUpgradeTest struct {
	k8sClient     kubernetes.Interface
	dynamicClient dynamic.Interface

	// test config
	namespace           string
	eventType           string
	subscriberName      string
	applicationName     string
	eventTypeVersion    string
	subscriptionName    string
	serviceInstanceName string
}

func NewEventingUpgradeTest(k8sClient kubernetes.Interface, dynamicClient dynamic.Interface) *EventingUpgradeTest {
	return &EventingUpgradeTest{
		k8sClient:     k8sClient,
		dynamicClient: dynamicClient,

		// test config same as the EventMesh test config
		eventTypeVersion:    "v1",
		eventType:           defaultName,
		subscriberName:      defaultName,
		applicationName:     defaultName,
		subscriptionName:    defaultName,
		serviceInstanceName: defaultName,
	}
}

func (e *EventingUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	// TODO implement when Kyma is fully integrated with the new Eventing solution https://github.com/kyma-project/kyma/issues/10866
	log.Info("CreateResources for EventingUpgradeTest is not implemented yet")
	return nil
}

func (e *EventingUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	// set the target namespace for the test
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
			return err
		}
	}

	return nil
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
	check := func() error {
		return helpers.CheckEvent(fmt.Sprintf("http://%s.%s.svc.cluster.local:9000/ce/%v/%v/%v",
			e.subscriberName, e.namespace, e.applicationName, e.eventType, e.eventTypeVersion))
	}

	return retry.Do(func() error {
		if err := check(); err != nil {
			return err
		}

		return nil
	})
}
