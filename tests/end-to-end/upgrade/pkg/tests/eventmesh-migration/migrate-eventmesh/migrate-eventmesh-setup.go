package migrateeventmesh

import (
	"fmt"
	"time"

	"github.com/avast/retry-go"

	"github.com/sirupsen/logrus"

	. "github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/eventmesh-migration/helpers"
)

type migrateEventMeshFlow struct {
	MigrateFromEventMeshUpgradeTest
	namespace string

	applicationName     string
	serviceInstanceName string
	subscriberName      string
	eventTypeVersion    string
	eventType           string
	triggerName         string
	brokerName          string

	log  logrus.FieldLogger
	stop <-chan struct{}
}

func newMigrateEventMeshFlow(e *MigrateFromEventMeshUpgradeTest,
	stop <-chan struct{}, log logrus.FieldLogger, namespace string) *migrateEventMeshFlow {
	return &migrateEventMeshFlow{
		MigrateFromEventMeshUpgradeTest: *e,
		stop:                            stop,
		log:                             log,
		namespace:                       namespace,
		applicationName:                 "migrate-eventmesh-upgrade",
		serviceInstanceName:             "migrate-eventmesh-upgrade",
		subscriberName:                  "migrate-eventmesh-upgrade",
		eventTypeVersion:                "migrate-eventmesh-upgrade",
		eventType:                       "migrate-eventmesh-upgrade",
		triggerName:                     "migrate-eventmesh-upgrade",
		brokerName:                      "migrate-eventmesh-upgrade",
	}
}

func (f *migrateEventMeshFlow) CreateApplication() error {
	return CreateApplication(f.appConnectorInterface, f.applicationName,
		WithAccessLabel(f.applicationName),
		WithEventService(f.serviceInstanceName),
	)
}

func (f *migrateEventMeshFlow) CreateSubscriber() error {
	return CreateSubscriber(f.k8sInterface, f.subscriberName, f.namespace)
}

func (f *migrateEventMeshFlow) WaitForSubscriber() error {
	return WaitForSubscriber(f.k8sInterface, f.subscriberName, f.namespace)
}

func (f *migrateEventMeshFlow) WaitForApplication() error {
	return WaitForApplication(f.appConnectorInterface, f.messagingClient, f.servingClient, f.applicationName)
}

func (f *migrateEventMeshFlow) CreateApplicationMapping() error {
	return CreateApplicationMapping(f.appBrokerCli, f.applicationName, f.namespace)
}

func (f *migrateEventMeshFlow) CreateServiceInstance() error {
	return CreateServiceInstance(f.scCli, f.serviceInstanceName, f.namespace)
}

func (f *migrateEventMeshFlow) CreateTrigger() error {
	return CreateTrigger(f.eventingCli, f.triggerName, f.namespace,
		WithFilter(f.eventTypeVersion, f.eventType, f.applicationName),
		WithURISubscriber(fmt.Sprintf("http://%s.%s.svc.cluster.local:9000/v3/events", f.subscriberName, f.namespace)))
}

func (f *migrateEventMeshFlow) CheckEvent() error {
	return CheckEvent(fmt.Sprintf("http://%s.%s.svc.cluster.local:9000/v3/results", f.subscriberName, f.namespace), f.eventType, f.eventTypeVersion)
}

func (f *migrateEventMeshFlow) WaitForServiceInstance() error {
	return WaitForServiceInstance(f.scCli, f.serviceInstanceName, f.namespace)
}

func (f *migrateEventMeshFlow) WaitForBroker() error {
	return WaitForBroker(f.eventingCli, f.brokerName, f.namespace, retry.Delay(10*time.Second), retry.DelayType(retry.FixedDelay), retry.Attempts(10))
}

func (f *migrateEventMeshFlow) WaitForTrigger() error {
	return WaitForTrigger(f.eventingCli, f.triggerName, f.namespace)
}

func (f *migrateEventMeshFlow) CreateSubscription() error {
	return CreateSubscription(f.ebCli, f.subscriberName, f.namespace, f.eventType, f.applicationName)
}

func (f *migrateEventMeshFlow) CheckSubscriptionReady() error {
	return CheckSubscriptionReady(f.ebCli, f.subscriberName, f.namespace)
}
