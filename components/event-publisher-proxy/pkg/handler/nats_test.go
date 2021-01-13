package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func TestNatsHandler(t *testing.T) {
	t.Parallel()

	logger := logrus.New()
	logger.Info("TestNatsHandler started")

	// a cancelable context to be used
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port, err := testingutils.GeneratePort()
	if err != nil {
		t.Fatalf("failed to generate port: %v", err)
	}
	cfgNats := newEnvConfig(port)

	// configure message receiver
	messageReceiver := receiver.NewHttpMessageReceiver(cfgNats.Port)
	assert.NotNil(t, messageReceiver)

	// Start Nats server
	natsServer := eventingtesting.RunNatsServerOnPort(cfgNats.Port + 1)
	assert.NotNil(t, natsServer)
	defer natsServer.Shutdown()

	// create a Nats sender
	natsUrl := natsServer.ClientURL()
	assert.NotEmpty(t, natsUrl)
	sender := sender.NewNatsMessageSender(ctx, natsUrl, logger)

	// configure legacyTransformer
	legacyTransformer := legacy.NewTransformer(
		cfgNats.ToConfig().BEBNamespace,
		cfgNats.ToConfig().EventTypePrefix,
	)

	// Setting up fake informers
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 to scheme: %v", err)
	}
	if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add eventing v1alpha1 to scheme: %v", err)
	}
	subscription := testingutils.NewSubscription()

	subUnstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(subscription)
	if err != nil {
		t.Fatalf("failed to convert subscription to unstructured obj: %v", err)
	}
	// Creating unstructured subscriptions
	subUnstructured := &unstructured.Unstructured{
		Object: subUnstructuredMap,
	}
	// Setting Kind information in unstructured subscription
	subscriptionGVK := schema.GroupVersionKind{
		Group:   subscribed.GVR.Group,
		Version: subscribed.GVR.Version,
		Kind:    "Subscription",
	}
	subUnstructured.SetGroupVersionKind(subscriptionGVK)
	// Configuring fake dynamic client
	dynamicTestClient := dynamicfake.NewSimpleDynamicClient(scheme, subUnstructured)

	dFilteredSharedInfFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicTestClient,
		10*time.Second,
		v1.NamespaceAll,
		nil,
	)
	genericInf := dFilteredSharedInfFactory.ForResource(subscribed.GVR)
	t.Logf("Waiting for cache to resync")
	subscribed.WaitForCacheSyncOrDie(ctx, dFilteredSharedInfFactory)
	t.Logf("Informers resynced successfully")
	subLister := genericInf.Lister()
	subscribedProcessor := &subscribed.Processor{
		SubscriptionLister: &subLister,
		Config:             cfgNats.ToConfig(),
		Logger:             logrus.New(),
	}

	// start handler which blocks until it receives a shutdown signal
	opts := &options.Options{MaxRequestSize: 65536}
	natsHandler := NewNatsHandler(messageReceiver, sender, cfgNats.RequestTimeout, legacyTransformer, opts, subscribedProcessor, logger)
	assert.NotNil(t, natsHandler)
	go func() {
		if err := natsHandler.Start(ctx); err != nil {
			t.Errorf("Failed to start handler with error: %v", err)
		}
	}()

	// test environment
	healthEndpoint := fmt.Sprintf("http://localhost:%d/healthz", port)
	publishEndpoint := fmt.Sprintf("http://localhost:%d/publish", port)
	publishLegacyEndpoint := fmt.Sprintf("http://localhost:%d/app/v1/events", port)
	subscribedEndpointFormat := "http://localhost:%d/%s/v1/events/subscribed"
	// wait that the embedded servers are started
	testingutils.WaitForHandlerToStart(t, healthEndpoint)

	// run the tests for publishing cloudevents
	for _, testCase := range testCasesForCloudEvents {
		t.Run(testCase.name, func(t *testing.T) {
			body, headers := testCase.provideMessage()
			resp, err := testingutils.SendEvent(publishEndpoint, body, headers)
			if err != nil {
				t.Errorf("Failed to send event with error: %v", err)
			}
			_ = resp.Body.Close()
			if testCase.wantStatusCode != resp.StatusCode {
				t.Errorf("Test failed, want status code:%d but got:%d", testCase.wantStatusCode, resp.StatusCode)
			}
		})
	}
	// run the tests for publishing legacy events
	for _, testCase := range testCasesForLegacyEvents {
		t.Run(testCase.name, func(t *testing.T) {
			body, headers := testCase.provideMessage()

			resp, err := testingutils.SendEvent(publishLegacyEndpoint, body, headers)
			if err != nil {
				t.Errorf("Failed to send event with error: %v", err)
			}

			if testCase.wantStatusCode != resp.StatusCode {
				t.Errorf("Test failed, want status code:%d but got:%d", testCase.wantStatusCode, resp.StatusCode)
			}

			if testCase.wantStatusCode == http.StatusOK {
				testingutils.ValidateOkResponse(t, *resp, &testCase.wantResponse)
			} else {
				testingutils.ValidateErrorResponse(t, *resp, &testCase.wantResponse)
			}
		})
	}
	// run the tests for subscribed endpoint
	for _, testCase := range testCasesForSubscribedEndpoit {
		t.Run(testCase.name, func(t *testing.T) {
			subscribedURL := fmt.Sprintf(subscribedEndpointFormat, port, testCase.appName)
			resp, err := testingutils.QuerySubscribedEndpoint(subscribedURL)
			if err != nil {
				t.Errorf("failed to send event with error: %v", err)
			}

			if testCase.wantStatusCode != resp.StatusCode {
				t.Errorf("test failed, want status code:%d but got:%d", testCase.wantStatusCode, resp.StatusCode)
			}
			defer func() { _ = resp.Body.Close() }()
			respBodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("failed to convert body to bytes: %v", err)
			}
			gotEventsResponse := subscribed.Events{}
			err = json.Unmarshal(respBodyBytes, &gotEventsResponse)
			if err != nil {
				t.Errorf("failed to unmarshal body bytes to events response: %v", err)
			}
			if !reflect.DeepEqual(testCase.wantResponse, gotEventsResponse) {
				t.Errorf("incorrect response, wanted: %v, got: %v", testCase.wantResponse, gotEventsResponse)
			}
		})
	}
}

func newEnvConfig(port int) *env.NatsConfig {
	return &env.NatsConfig{
		Port:                  port,
		NatsPublishURL:        fmt.Sprintf("http://localhost:%d", port+1),
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   2,
		RequestTimeout:        2 * time.Second,
		LegacyNamespace:       "beb.namespace", //"kyma",
		LegacyEventTypePrefix: "event.type.prefix",
	}
}
