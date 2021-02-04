package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/handlertest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/informers"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

type Test struct {
	logger     *logrus.Logger
	natsConfig *env.NatsConfig
	natsServer *server.Server
}

func (t *Test) setup() {
	port := testingutils.GeneratePortOrDie()
	natsPort := testingutils.GeneratePortOrDie()
	t.natsConfig = newEnvConfig(port, natsPort)
	t.logger = logrus.New()
	t.natsServer = eventingtesting.RunNatsServerOnPort(port + 1)
}

func (t *Test) cleanup() {
	t.natsServer.Shutdown()
}

var test = Test{}

func TestMain(m *testing.M) {
	test.setup()
	code := m.Run()
	test.cleanup()
	os.Exit(code)
}

func TestNatsHandlerForCloudEvents(t *testing.T) {
	exec := func(t *testing.T, applicationName, expectedNatsSubject string) {
		// test environment
		var (
			healthEndpoint  = fmt.Sprintf("http://localhost:%d/healthz", test.natsConfig.Port)
			publishEndpoint = fmt.Sprintf("http://localhost:%d/publish", test.natsConfig.Port)
		)

		test.logger.Info("TestNatsHandler started")

		// a cancelable context to be used
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// configure message receiver
		messageReceiver := receiver.NewHttpMessageReceiver(test.natsConfig.Port)
		assert.NotNil(t, messageReceiver)

		// create a Nats sender
		natsUrl := test.natsServer.ClientURL()
		assert.NotEmpty(t, natsUrl)
		msgSender := sender.NewNatsMessageSender(ctx, natsUrl, test.logger)

		// configure legacyTransformer
		appLister := handlertest.NewApplicationListerOrDie(ctx, applicationName)
		legacyTransformer := legacy.NewTransformer(
			test.natsConfig.ToConfig().BEBNamespace,
			test.natsConfig.ToConfig().EventTypePrefix,
			appLister,
		)

		// Setting up fake informers
		scheme := runtime.NewScheme()
		if err := corev1.AddToScheme(scheme); err != nil {
			t.Fatalf("failed to add corev1 to scheme: %v", err)
		}
		if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
			t.Fatalf("failed to add eventing v1alpha1 to scheme: %v", err)
		}
		subscription := testingutils.NewSubscription(testingutils.SubscriptionWithFilter(testingutils.MessagingNamespace, testingutils.CloudEventTypeNotClean))

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

		dFilteredSharedInfFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicTestClient, 10*time.Second, v1.NamespaceAll, nil)
		genericInf := dFilteredSharedInfFactory.ForResource(subscribed.GVR)
		t.Logf("Waiting for cache to resync")
		informers.WaitForCacheSyncOrDie(ctx, dFilteredSharedInfFactory)
		t.Logf("Informers resynced successfully")
		subLister := genericInf.Lister()
		subscribedProcessor := &subscribed.Processor{
			SubscriptionLister: &subLister,
			Config:             test.natsConfig.ToConfig(),
			Logger:             logrus.New(),
		}

		// start handler which blocks until it receives a shutdown signal
		opts := &options.Options{MaxRequestSize: 65536}
		natsHandler := NewNatsHandler(messageReceiver, msgSender, test.natsConfig.RequestTimeout, legacyTransformer, opts, subscribedProcessor, test.logger)
		assert.NotNil(t, natsHandler)
		go func() {
			if err := natsHandler.Start(ctx); err != nil {
				t.Errorf("Failed to start handler with error: %v", err)
			}
		}()

		// wait that the embedded servers are started
		testingutils.WaitForHandlerToStart(t, healthEndpoint)

		// prepare event type from subscription
		assert.NotNil(t, subscription.Spec.Filter)
		assert.NotEmpty(t, subscription.Spec.Filter.Filters)
		eventType := subscription.Spec.Filter.Filters[0].EventType.Value

		// publish a message to NATS and validate it
		connection := connectToNatsOrFail(t, natsUrl)
		validator := validateNatsSubject(t, expectedNatsSubject)
		subscribeToEventOrFail(t, connection, eventType, validator)

		// run the tests for publishing cloudevents
		for _, testCase := range handlertest.TestCasesForCloudEvents {
			t.Run(testCase.Name, func(t *testing.T) {
				body, headers := testCase.ProvideMessage()
				resp, err := testingutils.SendEvent(publishEndpoint, body, headers)
				if err != nil {
					t.Errorf("Failed to send event with error: %v", err)
				}
				_ = resp.Body.Close()
				if testCase.WantStatusCode != resp.StatusCode {
					t.Errorf("Test failed, want status code:%d but got:%d", testCase.WantStatusCode, resp.StatusCode)
				}
			})
		}
	}

	// make sure not to change the cloudevent, even if its event-type contains none-alphanumeric characters
	exec(t, testingutils.ApplicationName, testingutils.CloudEventTypeNotClean)
	exec(t, testingutils.ApplicationNameNotClean, testingutils.CloudEventTypeNotClean)
}

func TestNatsHandlerForLegacyEvents(t *testing.T) {
	exec := func(t *testing.T, applicationName string, expectedNatsSubject string) {
		// test environment
		var (
			healthEndpoint        = fmt.Sprintf("http://localhost:%d/healthz", test.natsConfig.Port)
			publishLegacyEndpoint = fmt.Sprintf("http://localhost:%d/%s/v1/events", test.natsConfig.Port, applicationName)
		)

		test.logger.Info("TestNatsHandler started")

		// a cancelable context to be used
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// configure message receiver
		messageReceiver := receiver.NewHttpMessageReceiver(test.natsConfig.Port)
		assert.NotNil(t, messageReceiver)

		// create a Nats sender
		natsUrl := test.natsServer.ClientURL()
		assert.NotEmpty(t, natsUrl)
		msgSender := sender.NewNatsMessageSender(ctx, natsUrl, test.logger)

		// configure legacyTransformer
		appLister := handlertest.NewApplicationListerOrDie(ctx, applicationName)
		legacyTransformer := legacy.NewTransformer(
			test.natsConfig.ToConfig().BEBNamespace,
			test.natsConfig.ToConfig().EventTypePrefix,
			appLister,
		)

		// Setting up fake informers
		scheme := runtime.NewScheme()
		if err := corev1.AddToScheme(scheme); err != nil {
			t.Fatalf("failed to add corev1 to scheme: %v", err)
		}
		if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
			t.Fatalf("failed to add eventing v1alpha1 to scheme: %v", err)
		}
		subscription := testingutils.NewSubscription(testingutils.SubscriptionWithFilter(testingutils.MessagingNamespace, testingutils.CloudEventTypeNotClean))

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

		dFilteredSharedInfFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicTestClient, 10*time.Second, v1.NamespaceAll, nil)
		genericInf := dFilteredSharedInfFactory.ForResource(subscribed.GVR)
		t.Logf("Waiting for cache to resync")
		informers.WaitForCacheSyncOrDie(ctx, dFilteredSharedInfFactory)
		t.Logf("Informers resynced successfully")
		subLister := genericInf.Lister()
		subscribedProcessor := &subscribed.Processor{
			SubscriptionLister: &subLister,
			Config:             test.natsConfig.ToConfig(),
			Logger:             logrus.New(),
		}

		// start handler which blocks until it receives a shutdown signal
		opts := &options.Options{MaxRequestSize: 65536}
		natsHandler := NewNatsHandler(messageReceiver, msgSender, test.natsConfig.RequestTimeout, legacyTransformer, opts, subscribedProcessor, test.logger)
		assert.NotNil(t, natsHandler)
		go func() {
			if err := natsHandler.Start(ctx); err != nil {
				t.Errorf("Failed to start handler with error: %v", err)
			}
		}()

		// wait that the embedded servers are started
		testingutils.WaitForHandlerToStart(t, healthEndpoint)

		// prepare event type from subscription
		assert.NotNil(t, subscription.Spec.Filter)
		assert.NotEmpty(t, subscription.Spec.Filter.Filters)
		eventType := subscription.Spec.Filter.Filters[0].EventType.Value

		// publish a message to NATS and validate it
		connection := connectToNatsOrFail(t, natsUrl)
		validator := validateNatsSubject(t, expectedNatsSubject)
		subscribeToEventOrFail(t, connection, eventType, validator)

		// run the tests for publishing legacy events
		for _, testCase := range handlertest.TestCasesForLegacyEvents {
			t.Run(testCase.Name, func(t *testing.T) {
				body, headers := testCase.ProvideMessage()
				resp, err := testingutils.SendEvent(publishLegacyEndpoint, body, headers)
				if err != nil {
					t.Fatalf("Failed to send event with error: %v", err)
				}

				if testCase.WantStatusCode != resp.StatusCode {
					t.Fatalf("Test failed, want status code:%d but got:%d", testCase.WantStatusCode, resp.StatusCode)
				}

				if testCase.WantStatusCode == http.StatusOK {
					handlertest.ValidateOkResponse(t, *resp, &testCase.WantResponse)
				} else {
					handlertest.ValidateErrorResponse(t, *resp, &testCase.WantResponse)
				}
			})
		}
	}

	// make sure to clean the legacy event, so that its event-type is free from none-alphanumeric characters
	exec(t, testingutils.ApplicationName, testingutils.CloudEventType)
	exec(t, testingutils.ApplicationNameNotClean, testingutils.CloudEventType)
}

func TestNatsHandlerForSubscribedEndpoint(t *testing.T) {
	// test environment
	var (
		applicationName          = testingutils.ApplicationName
		healthEndpoint           = fmt.Sprintf("http://localhost:%d/healthz", test.natsConfig.Port)
		subscribedEndpointFormat = "http://localhost:%d/%s/v1/events/subscribed"
	)

	test.logger.Info("TestNatsHandler started")

	// a cancelable context to be used
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// configure message receiver
	messageReceiver := receiver.NewHttpMessageReceiver(test.natsConfig.Port)
	assert.NotNil(t, messageReceiver)

	// create a Nats sender
	natsUrl := test.natsServer.ClientURL()
	assert.NotEmpty(t, natsUrl)
	msgSender := sender.NewNatsMessageSender(ctx, natsUrl, test.logger)

	// configure legacyTransformer
	appLister := handlertest.NewApplicationListerOrDie(ctx, applicationName)
	legacyTransformer := legacy.NewTransformer(
		test.natsConfig.ToConfig().BEBNamespace,
		test.natsConfig.ToConfig().EventTypePrefix,
		appLister,
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
	informers.WaitForCacheSyncOrDie(ctx, dFilteredSharedInfFactory)
	t.Logf("Informers resynced successfully")
	subLister := genericInf.Lister()
	subscribedProcessor := &subscribed.Processor{
		SubscriptionLister: &subLister,
		Config:             test.natsConfig.ToConfig(),
		Logger:             logrus.New(),
	}

	// start handler which blocks until it receives a shutdown signal
	opts := &options.Options{MaxRequestSize: 65536}
	natsHandler := NewNatsHandler(messageReceiver, msgSender, test.natsConfig.RequestTimeout, legacyTransformer, opts, subscribedProcessor, test.logger)
	assert.NotNil(t, natsHandler)
	go func() {
		if err := natsHandler.Start(ctx); err != nil {
			t.Errorf("Failed to start handler with error: %v", err)
		}
	}()

	// wait that the embedded servers are started
	testingutils.WaitForHandlerToStart(t, healthEndpoint)

	// run the tests for subscribed endpoint
	for _, testCase := range handlertest.TestCasesForSubscribedEndpoint {
		t.Run(testCase.Name, func(t *testing.T) {
			subscribedURL := fmt.Sprintf(subscribedEndpointFormat, test.natsConfig.Port, testCase.AppName)
			resp, err := testingutils.QuerySubscribedEndpoint(subscribedURL)
			if err != nil {
				t.Fatalf("failed to send event with error: %v", err)
			}

			if testCase.WantStatusCode != resp.StatusCode {
				t.Fatalf("test failed, want status code:%d but got:%d", testCase.WantStatusCode, resp.StatusCode)
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
			if !reflect.DeepEqual(testCase.WantResponse, gotEventsResponse) {
				t.Errorf("incorrect response, wanted: %v, got: %v", testCase.WantResponse, gotEventsResponse)
			}
		})
	}
}

func newEnvConfig(port, natsPort int) *env.NatsConfig {
	return &env.NatsConfig{
		Port:                  port,
		NatsPublishURL:        fmt.Sprintf("http://localhost:%d", natsPort),
		RequestTimeout:        2 * time.Second,
		LegacyNamespace:       testingutils.MessagingNamespace,
		LegacyEventTypePrefix: testingutils.MessagingEventTypePrefix,
	}
}

func connectToNatsOrFail(t *testing.T, natsUrl string) *nats.Conn {
	connection, err := nats.Connect(natsUrl, nats.RetryOnFailedConnect(true), nats.MaxReconnects(3), nats.ReconnectWait(time.Second))
	if err != nil {
		t.Fatalf("Failed to connect to NATS server with error: %v", err)
	}
	if connection.Status() != nats.CONNECTED {
		t.Fatal("Failed to connect to NATS server")
	}
	return connection
}

func subscribeToEventOrFail(t *testing.T, connection *nats.Conn, eventType string, validator nats.MsgHandler) {
	if _, err := connection.Subscribe(eventType, validator); err != nil {
		t.Fatalf("Failed to subscribe to event with error: %v", err)
	}
}

func validateNatsSubject(t *testing.T, subject string) nats.MsgHandler {
	return func(msg *nats.Msg) {
		if msg != nil && msg.Subject != subject {
			t.Errorf("invalid NATS subject, expected [%s] but found [%s]", subject, msg.Subject)
		}
	}
}
