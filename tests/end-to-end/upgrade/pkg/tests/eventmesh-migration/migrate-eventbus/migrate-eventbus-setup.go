package migrateeventbus

import (
	"fmt"
	"time"

	"github.com/avast/retry-go"

	"github.com/sirupsen/logrus"

	. "github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/eventmesh-migration/helpers"
)

const (
	integrationNamespace = "kyma-integration"
	eventServiceSuffix   = "event-service"
	eventServicePort     = "8081"
)

type migrateEventBusFlow struct {
	MigrateFromEventBusUpgradeTest
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

func newMigrateEventBusFlow(e *MigrateFromEventBusUpgradeTest,
	stop <-chan struct{}, log logrus.FieldLogger, namespace string) *migrateEventBusFlow {
	return &migrateEventBusFlow{
		MigrateFromEventBusUpgradeTest: *e,
		stop:                           stop,
		log:                            log,
		namespace:                      namespace,
		applicationName:                "migrate-eventbus-upgrade",
		serviceInstanceName:            "migrate-eventbus-upgrade",
		subscriberName:                 "migrate-eventbus-upgrade",
		eventTypeVersion:               "v1",
		eventType:                      "migrate-eventbus-upgrade",
		subscriptionName:               "migrate-eventbus-upgrade",
		brokerName:                     "default",
	}
}

func (f *migrateEventBusFlow) CreateApplication() error {
	return CreateApplication(f.appConnectorInf, f.applicationName,
		WithAccessLabel(f.applicationName),
		WithEventService(f.serviceInstanceName),
	)
}

func (f *migrateEventBusFlow) CreateSubscriber() error {
	return CreateSubscriber(f.k8sInf, f.subscriberName, f.namespace)
}

func (f *migrateEventBusFlow) WaitForSubscriber() error {
	return WaitForSubscriber(f.k8sInf, f.subscriberName, f.namespace)
}

func (f *migrateEventBusFlow) WaitForApplication() error {
	return WaitForApplication(f.appConnectorInf, f.messagingInf, f.servingInf, f.applicationName)
}

func (f *migrateEventBusFlow) CreateApplicationMapping() error {
	return CreateApplicationMapping(f.appBrokerInf, f.applicationName, f.namespace)
}

func (f *migrateEventBusFlow) CreateServiceInstance() error {
	return CreateServiceInstance(f.scInf, f.serviceInstanceName, f.namespace)
}

func (f *migrateEventBusFlow) CreateTrigger() error {
	return CreateTrigger(f.eventingInf, f.subscriptionName, f.namespace,
		WithFilter(f.eventTypeVersion, f.eventType, f.applicationName),
		WithURISubscriber(fmt.Sprintf("http://%s.%s.svc.cluster.local:9000/v3/events", f.subscriberName, f.namespace)))
}

func (f *migrateEventBusFlow) CheckEvent() error {
	return CheckEvent(fmt.Sprintf("http://%s.%s.svc.cluster.local:9000/v3/results", f.subscriberName, f.namespace), f.eventType, f.eventTypeVersion)
}

func (f *migrateEventBusFlow) WaitForServiceInstance() error {
	return WaitForServiceInstance(f.scInf, f.serviceInstanceName, f.namespace)
}

func (f *migrateEventBusFlow) WaitForBroker() error {
	return WaitForBroker(f.eventingInf, f.brokerName, f.namespace, retry.Delay(10*time.Second), retry.DelayType(retry.FixedDelay), retry.Attempts(10))
}

func (f *migrateEventBusFlow) DeleteBroker() error {
	return DeleteBroker(f.eventingInf, f.brokerName, f.namespace, retry.Delay(10*time.Second), retry.DelayType(retry.FixedDelay), retry.Attempts(10))
}

func (f *migrateEventBusFlow) RemoveBrokerInjectionLabel() error {
	return RemoveBrokerInjectionLabel(f.k8sInf, f.namespace, retry.Delay(10*time.Second), retry.DelayType(retry.FixedDelay), retry.Attempts(10))
}

func (f *migrateEventBusFlow) WaitForTrigger() error {
	return WaitForTrigger(f.eventingInf, f.subscriptionName, f.namespace)
}

func (f *migrateEventBusFlow) CreateSubscription() error {
	return CreateSubscription(f.ebInf, f.subscriberName, f.namespace, f.eventType, f.eventTypeVersion, f.applicationName)
}

func (f *migrateEventBusFlow) CheckSubscriptionReady() error {
	return CheckSubscriptionReady(f.ebInf, f.subscriberName, f.namespace)
}

func (f *migrateEventBusFlow) PublishTestEvent() error {
	return SendEvent(fmt.Sprintf("http://%s-%s.%s.svc.cluster.local:%s/%s/v1/events", f.applicationName, eventServiceSuffix, integrationNamespace, eventServicePort, f.applicationName), f.eventType, f.eventTypeVersion)
}
