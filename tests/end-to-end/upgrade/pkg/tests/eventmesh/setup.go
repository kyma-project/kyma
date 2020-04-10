package eventmesh

import (
	"fmt"
	"time"

	"github.com/avast/retry-go"

	"github.com/sirupsen/logrus"

	. "github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/eventmesh/helpers"
)

const (
	integrationNamespace = "kyma-integration"
	eventServiceSuffix   = "event-service"
	eventServicePort     = "8081"
	defaultName          = "eventmesh-upgrade"
)

type eventMeshFlow struct {
	EventMeshUpgradeTest
	namespace string

	applicationName     string
	serviceInstanceName string
	subscriberName      string
	subscriptionName    string
	eventTypeVersion    string
	eventType           string
	brokerName          string

	log  logrus.FieldLogger
	stop <-chan struct{}
}

func newEventMeshFlow(e *EventMeshUpgradeTest,
	stop <-chan struct{}, log logrus.FieldLogger, namespace string) *eventMeshFlow {
	return &eventMeshFlow{
		EventMeshUpgradeTest: *e,
		stop:                 stop,
		log:                  log,
		namespace:            namespace,
		applicationName:      defaultName,
		serviceInstanceName:  defaultName,
		subscriberName:       defaultName,
		eventTypeVersion:     "v1",
		eventType:            defaultName,
		subscriptionName:     defaultName,
		brokerName:           "default",
	}
}

func (f *eventMeshFlow) CreateApplication() error {
	return CreateApplication(f.appConnectorInterface, f.applicationName,
		WithAccessLabel(f.applicationName),
		WithEventService(f.serviceInstanceName),
	)
}

func (f *eventMeshFlow) CreateSubscriber() error {
	return CreateSubscriber(f.k8sInterface, f.subscriberName, f.namespace)
}

func (f *eventMeshFlow) WaitForSubscriber() error {
	return WaitForSubscriber(f.k8sInterface, f.subscriberName, f.namespace)
}

func (f *eventMeshFlow) WaitForApplication() error {
	return WaitForApplication(f.appConnectorInterface, f.messagingClient, f.servingClient, f.applicationName)
}

func (f *eventMeshFlow) CreateApplicationMapping() error {
	return CreateApplicationMapping(f.appBrokerCli, f.applicationName, f.namespace)
}

func (f *eventMeshFlow) CreateServiceInstance() error {
	return CreateServiceInstance(f.scCli, f.serviceInstanceName, f.namespace)
}

func (f *eventMeshFlow) CreateTrigger() error {
	return CreateTrigger(f.eventingCli, f.subscriptionName, f.namespace,
		WithFilter(f.eventTypeVersion, f.eventType, f.applicationName),
		WithURISubscriber(fmt.Sprintf("http://%s.%s.svc.cluster.local:9000/v3/events", f.subscriberName, f.namespace)))
}

func (f *eventMeshFlow) CheckEvent() error {
	return CheckEvent(fmt.Sprintf("http://%s.%s.svc.cluster.local:9000/v3/results", f.subscriberName, f.namespace), f.eventType, f.eventTypeVersion)
}

func (f *eventMeshFlow) WaitForServiceInstance() error {
	return WaitForServiceInstance(f.scCli, f.serviceInstanceName, f.namespace)
}

func (f *eventMeshFlow) WaitForBroker() error {
	return WaitForBroker(f.eventingCli, f.brokerName, f.namespace, retry.Delay(10*time.Second), retry.DelayType(retry.FixedDelay), retry.Attempts(10))
}

func (f *eventMeshFlow) WaitForTrigger() error {
	return WaitForTrigger(f.eventingCli, f.subscriptionName, f.namespace)
}

func (f *eventMeshFlow) PublishTestEvent() error {
	return SendEvent(fmt.Sprintf("http://%s-%s.%s.svc.cluster.local:%s/%s/v1/events", f.applicationName, eventServiceSuffix, integrationNamespace, eventServicePort, f.applicationName), f.eventType, f.eventTypeVersion)
}
