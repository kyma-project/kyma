package eventmesh

import (
	"fmt"
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	k8s "k8s.io/client-go/kubernetes"
	serving "knative.dev/serving/pkg/client/clientset/versioned"

	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	appbroker "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appconnector "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"

	messaging "knative.dev/eventing/pkg/client/clientset/versioned"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/config"
)

const (
	eventServicePort = 8081
)

type EventMeshTest struct {
	k8s            k8s.Interface
	appConnector   appconnector.Interface
	serviceCatalog servicecatalog.Interface
	messaging      messaging.Interface
	appBroker      appbroker.Interface
	serving        serving.Interface
}

type eventMeshFlow struct {
	namespace string

	applicationName     string
	serviceInstanceName string
	subscriberName      string
	eventTypeVersion    string
	eventType           string
	triggerName         string
	brokerName          string

	log logrus.FieldLogger

	k8s            k8s.Interface
	appConnector   appconnector.Interface
	serviceCatalog servicecatalog.Interface
	messaging      messaging.Interface
	appBroker      appbroker.Interface
	serving        serving.Interface

	subscriberURL string
}

func NewEventMeshTest() (*EventMeshTest, error) {
	k8sConfig, err := config.NewRestClientConfig()
	if err != nil {
		return nil, err
	}

	k8s, err := k8s.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	serviceCatalog, err := servicecatalog.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	appConnector, err := appconnector.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	appBroker, err := appbroker.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	messaging, err := messaging.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	serving, err := serving.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return &EventMeshTest{
		k8s:            k8s,
		appConnector:   appConnector,
		serviceCatalog: serviceCatalog,
		messaging:      messaging,
		appBroker:      appBroker,
		serving:        serving,
	}, nil
}

func (em *EventMeshTest) newFlow(namespace string) *eventMeshFlow {

	logger := logrus.New()
	// configure logger with text instead of json for easier reading in CI logs
	logger.Formatter = &logrus.TextFormatter{}
	// show file and line number
	logger.SetReportCaller(true)
	res := &eventMeshFlow{
		namespace:           namespace,
		applicationName:     "event-mesh-app",
		serviceInstanceName: "event-mesh-event-service",
		subscriberName:      "event-mesh-subscriber",
		eventTypeVersion:    "v1",
		eventType:           "event.mesh",
		triggerName:         "event-mesh-trigger",
		brokerName:          "default",
		log:                 logger,
		k8s:                 em.k8s,
		appConnector:        em.appConnector,
		serviceCatalog:      em.serviceCatalog,
		messaging:           em.messaging,
		appBroker:           em.appBroker,
		serving:             em.serving,
		subscriberURL:       "",
	}
	return res
}

func (em *EventMeshTest) CreateResources(t *testing.T, namespace string) {
	err := em.newFlow(namespace).createResources()
	require.NoError(t, err)
}

func (em *EventMeshTest) TestResources(t *testing.T, namespace string) {
	err := em.newFlow(namespace).testResources()
	require.NoError(t, err)
}

func (f *eventMeshFlow) CreateApplication() error {
	return CreateApplication(f.appConnector, f.applicationName,
		WithAccessLabel(f.applicationName),
		WithEventService(f.serviceInstanceName),
	)
}

func (f *eventMeshFlow) CreateSubscriber() error {
	return CreateSubscriber(f.k8s, f.subscriberName, f.namespace)
}

func (f *eventMeshFlow) WaitForSubscriber() error {
	return WaitForSubscriber(f.k8s, f.subscriberName, f.namespace)
}

func (f *eventMeshFlow) WaitForApplication() error {
	return WaitForApplication(f.appConnector, f.messaging, f.serving, f.applicationName)
}

func (f *eventMeshFlow) CreateApplicationMapping() error {
	return CreateApplicationMapping(f.appBroker, f.applicationName, f.namespace)
}

func (f *eventMeshFlow) CreateServiceInstance() error {
	return CreateServiceInstance(f.serviceCatalog, f.serviceInstanceName, f.namespace)
}

func (f *eventMeshFlow) CreateTrigger() error {
	return CreateTrigger(f.messaging, f.triggerName, f.namespace,
		WithFilter(f.eventTypeVersion, f.eventType, f.applicationName),
		WithURISubscriber(fmt.Sprintf("http://%s.%s.svc.cluster.local:9000/v3/events", f.subscriberName, f.namespace)))
}

func (f *eventMeshFlow) CheckEvent() error {
	return CheckEvent(fmt.Sprintf("http://%s.%s.svc.cluster.local:9000/v3/results", f.subscriberName, f.namespace), f.eventType, f.eventTypeVersion)
}

func (f *eventMeshFlow) createResources() error {
	for _, fn := range []func() error{
		f.CreateApplication,
		f.CreateSubscriber,
		f.WaitForApplication,
		f.WaitForSubscriber,
		f.CreateApplicationMapping,
		f.CreateServiceInstance,
		f.WaitForServiceInstance,
		f.WaitForBroker,
		f.CreateTrigger,
		f.WaitForTrigger,
	} {
		if err := fn(); err != nil {
			return fmt.Errorf("CreateResources() failed with: %w", err)
		}
	}
	return nil
}

func (f *eventMeshFlow) testResources() error {
	for _, fn := range []func() error{
		f.WaitForApplication,
		f.WaitForSubscriber,
		f.WaitForServiceInstance,
		f.WaitForBroker,
		f.WaitForTrigger,
		f.PublishTestEventToMesh,
		f.CheckEvent,
		f.PublishTestEventToCompatibilityLayer,
		f.CheckEvent,
	} {
		if err := fn(); err != nil {
			return fmt.Errorf("TestResources() failed with: %w", err)
		}
	}
	return nil
}

func (f *eventMeshFlow) WaitForTrigger() error {
	return WaitForTrigger(f.messaging, f.triggerName, f.namespace)
}

func (f *eventMeshFlow) WaitForBroker() error {
	return WaitForBroker(f.messaging, f.brokerName, f.namespace, retry.Delay(10*time.Second), retry.DelayType(retry.FixedDelay), retry.Attempts(10))
}

func (f *eventMeshFlow) PublishTestEventToMesh() error {
	return SendCloudEvent(fmt.Sprintf("http://%s.%s.svc.cluster.local", f.applicationName, integrationNamespace), "Dumbidu", f.eventType, f.eventTypeVersion)
}

func (f *eventMeshFlow) PublishTestEventToCompatibilityLayer() error {
	return SendLegacyEvent(fmt.Sprintf("http://%s-event-service.%s.svc.cluster.local:%v/%s/v1/events", f.applicationName, integrationNamespace, eventServicePort, f.applicationName), "Dumbidu", f.eventType, f.eventTypeVersion)
}

func (f *eventMeshFlow) WaitForServiceInstance() error {
	return WaitForServiceInstance(f.serviceCatalog, f.serviceInstanceName, f.namespace)
}
