/*
Copyright 2020 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package eventmesh

import (
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"
	"github.com/smartystreets/goconvey/convey"
	k8sCS "k8s.io/client-go/kubernetes"
	servingCS "knative.dev/serving/pkg/client/clientset/versioned"

	serviceCatalogCS "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	appBrokerCS "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appConnectorCS "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"

	messagingCS "knative.dev/eventing/pkg/client/clientset/versioned"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/config"
)

type EventMeshTest struct {
	k8s            k8sCS.Interface
	appConnector   appConnectorCS.Interface
	serviceCatalog serviceCatalogCS.Interface
	messaging      messagingCS.Interface
	appBroker      appBrokerCS.Interface
	serving        servingCS.Interface
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

	k8s            k8sCS.Interface
	appConnector   appConnectorCS.Interface
	serviceCatalog serviceCatalogCS.Interface
	messaging      messagingCS.Interface
	appBroker      appBrokerCS.Interface
	serving        servingCS.Interface

	subscriberURL string
}

func NewEventMeshTest() (*EventMeshTest, error) {
	k8sConfig, err := config.NewRestClientConfig()
	if err != nil {
		return nil, err
	}

	k8s, err := k8sCS.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	serviceCatalog, err := serviceCatalogCS.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	appConnector, err := appConnectorCS.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	appBroker, err := appBrokerCS.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	messaging, err := messagingCS.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	serving, err := servingCS.NewForConfig(k8sConfig)
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

func (em *EventMeshTest) CreateResources(namespace string) {
	err := em.newFlow(namespace).createResources()
	convey.So(err, convey.ShouldBeNil)
}

func (em *EventMeshTest) TestResources(namespace string) {
	err := em.newFlow(namespace).testResources()
	convey.So(err, convey.ShouldBeNil)
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
		f.PublishTestEvent,
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

func (f *eventMeshFlow) PublishTestEvent() error {
	return SendEvent(fmt.Sprintf("http://%s.%s.svc.cluster.local", f.applicationName, integrationNamespace), "Dumbidu", f.eventType, f.eventTypeVersion)
}

func (f *eventMeshFlow) WaitForServiceInstance() error {
	return WaitForServiceInstance(f.serviceCatalog, f.serviceInstanceName, f.namespace)
}
