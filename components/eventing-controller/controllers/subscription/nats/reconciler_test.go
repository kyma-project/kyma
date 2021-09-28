package nats

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	events "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/fake"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

const (
	natsPort = 4221

	smallTimeout         = 5 * time.Second
	smallPollingInterval = 1 * time.Second

	timeout         = 60 * time.Second
	pollingInterval = 5 * time.Second

	namespaceName          = "test"
	subscriptionNameFormat = "sub-%d"
	subscriberNameFormat   = "subscriber-%d"
)

type testCase func(id int, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) bool

var (
	reconcilerTestCases = []testCase{
		testCreateDeleteSubscription,
		testCreateSubscriptionWithInvalidSink,
		testCreateSubscriptionWithEmptyProtocolProtocolSettingsDialect,
		testChangeSubscriptionConfiguration,
		testCreateSubscriptionWithEmptyEventType,
	}

	dispatcherTestCases = []testCase{
		testDispatcherWithOneSubscriber,
		testDispatcherWithMultipleSubscribers,
	}
)

func testCreateDeleteSubscription(id int, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("Create/Delete Subscription", func() {
		It("Should create/delete NATS Subscription", func() {
			ctx := context.Background()
			cancel = startReconciler(eventTypePrefix, DefaultSinkValidator)
			defer cancel()
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)
			subscriberName := fmt.Sprintf(subscriberNameFormat, id)

			// create subscriber svc
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subscriberName, namespaceName)
			ensureSubscriberSvcCreated(ctx, subscriberSvc)

			// create subscription
			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithFilter(reconcilertesting.EventSource, eventTypeToSubscribe), reconcilertesting.WithWebhookForNats)
			reconcilertesting.WithValidSink(namespaceName, subscriberSvc.Name, subscription)
			ensureSubscriptionCreated(ctx, subscription)

			getSubscription(ctx, subscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionTrue, "")),
				reconcilertesting.HaveSubsConfiguration(&eventingv1alpha1.SubscriptionConfig{
					MaxInFlightMessages: defaultSubsConfig.MaxInFlightMessages,
				}),
			))

			// check for subscription at nats
			backendSubscription := getSubscriptionFromNats(natsBackend.GetAllSubscription(), subscriptionName)
			Expect(backendSubscription).NotTo(BeNil())
			Expect(backendSubscription.IsValid()).To(BeTrue())
			Expect(backendSubscription.Subject).Should(Equal(natsSubjectToPublish))

			Expect(k8sClient.Delete(ctx, subscription)).Should(BeNil())
			isSubscriptionDeleted(ctx, subscription).Should(reconcilertesting.HaveNotFoundSubscription(true))
		})
	})
}

func testCreateSubscriptionWithInvalidSink(id int, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("Create Subscription with invalid sink", func() {
		It("Should mark the Subscription as not ready", func() {
			ctx := context.Background()
			cancel = startReconciler(eventTypePrefix, DefaultSinkValidator)
			defer cancel()
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)

			// Create subscription
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithFilter(reconcilertesting.EventSource, eventTypeToSubscribe), reconcilertesting.WithWebhookForNats)
			givenSubscription.Spec.Sink = "invalid"
			ensureSubscriptionCreated(ctx, givenSubscription)

			getSubscription(ctx, givenSubscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionFalse, "sink URL scheme should be 'http' or 'https'")),
			))

			var subscriptionEvents = v1.EventList{}
			subscriptionEvent := v1.Event{
				Reason:  string(events.ReasonValidationFailed),
				Message: "Sink URL scheme should be HTTP or HTTPS: invalid",
				Type:    v1.EventTypeWarning,
			}
			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertesting.HaveEvent(subscriptionEvent))

			By("Updating the subscription configuration in the spec with invalid URL scheme")
			changedSub := givenSubscription.DeepCopy()
			changedSub.Spec.Sink = "http://127.0.0. 1"
			Expect(k8sClient.Update(ctx, changedSub)).Should(BeNil())

			getSubscription(ctx, givenSubscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionFalse, "not able to parse sink url with error: parse \"http://127.0.0. 1\": invalid character \" \" in host name")),
			))

			subscriptionEvent = v1.Event{
				Reason:  string(events.ReasonValidationFailed),
				Message: "Not able to parse Sink URL with error: parse \"http://127.0.0. 1\": invalid character \" \" in host name",
				Type:    v1.EventTypeWarning,
			}
			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertesting.HaveEvent(subscriptionEvent))

			By("Updating the subscription configuration in the spec with valid URL scheme")
			changedSub = givenSubscription.DeepCopy()
			changedSub.Spec.Sink = "http://127.0.0.1"
			Expect(k8sClient.Update(ctx, changedSub)).Should(BeNil())

			getSubscription(ctx, givenSubscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionFalse, "sink does not contain suffix: svc.cluster.local in the URL")),
			))

			subscriptionEvent = v1.Event{
				Reason:  string(events.ReasonValidationFailed),
				Message: "Sink does not contain suffix: svc.cluster.local",
				Type:    v1.EventTypeWarning,
			}
			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertesting.HaveEvent(subscriptionEvent))

			By("Updating the subscription configuration in the spec with invalid service name scheme")
			changedSub = givenSubscription.DeepCopy()
			changedSub.Spec.Sink = fmt.Sprintf("https://%s.%s.%s.svc.cluster.local", "testapp", "testsub", "test")
			Expect(k8sClient.Update(ctx, changedSub)).Should(BeNil())

			getSubscription(ctx, givenSubscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionFalse, "sink should contain 5 sub-domains: testapp.testsub.test.svc.cluster.local")),
			))

			subscriptionEvent = v1.Event{
				Reason:  string(events.ReasonValidationFailed),
				Message: "Sink should contain 5 sub-domains: testapp.testsub.test.svc.cluster.local",
				Type:    v1.EventTypeWarning,
			}
			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertesting.HaveEvent(subscriptionEvent))

			By("Updating the subscription configuration in the spec with invalid sink URL with subscriber namespace")
			changedSub = givenSubscription.DeepCopy()
			changedSub.Spec.Sink = fmt.Sprintf("https://%s.%s.svc.cluster.local", "testapp", "test-ns")
			Expect(k8sClient.Update(ctx, changedSub)).Should(BeNil())

			getSubscription(ctx, givenSubscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionFalse, "namespace of subscription: test and the namespace of subscriber: test-ns are different")),
			))

			subscriptionEvent = v1.Event{
				Reason:  string(events.ReasonValidationFailed),
				Message: "Namespace of subscription: test and the subscriber: test-ns are different",
				Type:    v1.EventTypeWarning,
			}
			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertesting.HaveEvent(subscriptionEvent))

			By("Updating the subscription configuration in the spec with invalid sink URL with non-existing service name")
			changedSub = givenSubscription.DeepCopy()
			reconcilertesting.WithValidSink(namespaceName, "testapp", changedSub)
			Expect(k8sClient.Update(ctx, changedSub)).Should(BeNil())

			getSubscription(ctx, givenSubscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionFalse, "sink is not valid cluster local svc, failed with error: Service \"testapp\" not found")),
			))

			subscriptionEvent = v1.Event{
				Reason:  string(events.ReasonValidationFailed),
				Message: "Sink does not correspond to a valid cluster local svc",
				Type:    v1.EventTypeWarning,
			}
			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertesting.HaveEvent(subscriptionEvent))

			By("Updating the subscription configuration in the spec with invalid sink URL with valid subscriber service name")
			changedSub = givenSubscription.DeepCopy()

			subscriberName := fmt.Sprintf(subscriberNameFormat, id)
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subscriberName, namespaceName)
			ensureSubscriberSvcCreated(ctx, subscriberSvc)

			reconcilertesting.WithValidSink(namespaceName, subscriberSvc.Name, changedSub)
			Expect(k8sClient.Update(ctx, changedSub)).Should(BeNil())

			getSubscription(ctx, givenSubscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionTrue, "")),
			))

			// check for subscription at nats
			backendSubscription := getSubscriptionFromNats(natsBackend.GetAllSubscription(), subscriptionName)
			Expect(backendSubscription).NotTo(BeNil())
			Expect(backendSubscription.IsValid()).To(BeTrue())
			Expect(backendSubscription.Subject).Should(Equal(natsSubjectToPublish))

			Expect(k8sClient.Delete(ctx, givenSubscription)).Should(BeNil())
			isSubscriptionDeleted(ctx, givenSubscription).Should(reconcilertesting.HaveNotFoundSubscription(true))
		})
	})
}

func testCreateSubscriptionWithEmptyProtocolProtocolSettingsDialect(id int, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("Create Subscription with empty protocol, protocolsettings and dialect", func() {
		It("Should mark the Subscription as ready", func() {
			ctx := context.Background()
			cancel = startReconciler(eventTypePrefix, DefaultSinkValidator)
			defer cancel()
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)
			subscriberName := fmt.Sprintf(subscriberNameFormat, id)

			// create subscriber svc
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subscriberName, namespaceName)
			ensureSubscriberSvcCreated(ctx, subscriberSvc)

			// create subscription
			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithFilter("", eventTypeToSubscribe))
			reconcilertesting.WithValidSink(namespaceName, subscriberSvc.Name, subscription)
			ensureSubscriptionCreated(ctx, subscription)

			getSubscription(ctx, subscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionTrue, "")),
			))

			// check for subscription at nats
			backendSubscription := getSubscriptionFromNats(natsBackend.GetAllSubscription(), subscriptionName)
			Expect(backendSubscription).NotTo(BeNil())
			Expect(backendSubscription.IsValid()).To(BeTrue())
			Expect(backendSubscription.Subject).Should(Equal(natsSubjectToPublish))
		})
	})
}

func testChangeSubscriptionConfiguration(id int, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("Change Subscription configuration", func() {
		It("Should reflect the new config in the subscription status", func() {
			By("Creating the subscription using the default config")
			ctx := context.Background()
			cancel = startReconciler(eventTypePrefix, DefaultSinkValidator)
			defer cancel()
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)
			subscriberName := fmt.Sprintf(subscriberNameFormat, id)

			// create subscriber svc
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subscriberName, namespaceName)
			ensureSubscriberSvcCreated(ctx, subscriberSvc)

			// create subscription
			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithFilter(reconcilertesting.EventSource, eventTypeToSubscribe), reconcilertesting.WithWebhookForNats)
			reconcilertesting.WithValidSink(namespaceName, subscriberSvc.Name, subscription)
			ensureSubscriptionCreated(ctx, subscription)

			getSubscription(ctx, subscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionTrue, "")),
				reconcilertesting.HaveSubsConfiguration(&eventingv1alpha1.SubscriptionConfig{
					MaxInFlightMessages: defaultSubsConfig.MaxInFlightMessages,
				}),
			))

			By("Updating the subscription configuration in the spec")

			newMaxInFlight := defaultSubsConfig.MaxInFlightMessages + 1
			changedSub := subscription.DeepCopy()
			changedSub.Spec.Config = &eventingv1alpha1.SubscriptionConfig{
				MaxInFlightMessages: newMaxInFlight,
			}
			Expect(k8sClient.Update(ctx, changedSub)).Should(BeNil())

			Eventually(subscriptionGetter(ctx, subscription.Name, subscription.Namespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveSubscriptionName(subscriptionName),
					reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
						eventingv1alpha1.ConditionSubscriptionActive,
						eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
						v1.ConditionTrue, ""),
					),
					reconcilertesting.HaveSubsConfiguration(&eventingv1alpha1.SubscriptionConfig{
						MaxInFlightMessages: newMaxInFlight,
					}),
				))

			// check for subscription at nats
			backendSubscription := getSubscriptionFromNats(natsBackend.GetAllSubscription(), subscriptionName)
			Expect(backendSubscription).NotTo(BeNil())
			Expect(backendSubscription.IsValid()).To(BeTrue())
			Expect(backendSubscription.Subject).Should(Equal(natsSubjectToPublish))

			Expect(k8sClient.Delete(ctx, subscription)).Should(BeNil())
			isSubscriptionDeleted(ctx, subscription).Should(reconcilertesting.HaveNotFoundSubscription(true))
		})
	})
}

func testCreateSubscriptionWithEmptyEventType(id int, eventTypePrefix, _, _ string) bool {
	return When("Create Subscription with empty event type", func() {
		It("Should mark the subscription as not ready", func() {
			ctx := context.Background()
			cancel = startReconciler(eventTypePrefix, DefaultSinkValidator)
			defer cancel()
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)
			subscriberName := fmt.Sprintf(subscriberNameFormat, id)

			// create subscriber svc
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subscriberName, namespaceName)
			ensureSubscriberSvcCreated(ctx, subscriberSvc)

			// Create subscription
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithFilter(reconcilertesting.EventSource, ""), reconcilertesting.WithWebhookForNats)
			reconcilertesting.WithValidSink(namespaceName, subscriberName, givenSubscription)
			ensureSubscriptionCreated(ctx, givenSubscription)

			getSubscription(ctx, givenSubscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionFalse, nats.ErrBadSubject.Error())),
			))
		})
	})
}

func testDispatcherWithOneSubscriber(id int, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("Sending Events through Dispatcher", func() {
		It("Should receive events in subscriber", func() {
			ctx := context.Background()

			// Start reconciler with empty checkSink function
			cancel = startReconciler(eventTypePrefix, func(ctx context.Context, r *Reconciler, subscription *eventingv1alpha1.Subscription) error {
				return nil
			})
			defer cancel()

			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)

			result := make(chan []byte)
			url, shutdown := newSubscriber(result)
			defer shutdown()

			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithFilter(reconcilertesting.EventSource, eventTypeToSubscribe), reconcilertesting.WithWebhookForNats)
			subscription.Spec.Sink = url
			ensureSubscriptionCreated(ctx, subscription)

			getSubscription(ctx, subscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionTrue, "")),
				reconcilertesting.HaveSubsConfiguration(&eventingv1alpha1.SubscriptionConfig{
					MaxInFlightMessages: defaultSubsConfig.MaxInFlightMessages,
				}),
			))

			connection, err := connectToNats(natsURL)
			Expect(err).ShouldNot(HaveOccurred())
			err = connection.Publish(natsSubjectToPublish, []byte(reconcilertesting.StructuredCloudEvent))
			Expect(err).ShouldNot(HaveOccurred())

			// make sure that the subscriber received the message
			sent := fmt.Sprintf(`"%s"`, reconcilertesting.EventData)
			Eventually(func() ([]byte, error) {
				return getFromChanOrTimeout(result, smallPollingInterval)
			}, timeout, pollingInterval).Should(WithTransform(bytesStringer, Equal(sent)))
		})
	})
}

func testDispatcherWithMultipleSubscribers(id int, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("Sending Events through Dispatcher for multiple subscribers", func() {
		It("Should receive events in subscribers", func() {
			ctx := context.Background()

			var subscription2 *eventingv1alpha1.Subscription
			eventType := reconcilertesting.EventTypePrefix + "." + reconcilertesting.ApplicationNameNotClean + "." + reconcilertesting.OrderUpdatedV1Event
			emptyEventType := reconcilertesting.ApplicationNameNotClean + "." + reconcilertesting.OrderUpdatedV1Event
			natsSubject := reconcilertesting.EventTypePrefix + "." + reconcilertesting.ApplicationName + "." + reconcilertesting.OrderUpdatedV1Event
			emptyNatsSubject := reconcilertesting.ApplicationName + "." + reconcilertesting.OrderUpdatedV1Event

			// Start reconciler with empty checkSink function
			cancel = startReconciler(eventTypePrefix, func(ctx context.Context, r *Reconciler, subscription *eventingv1alpha1.Subscription) error {
				return nil
			})
			defer cancel()

			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)

			subChan1 := make(chan []byte)
			url, shutdown := newSubscriber(subChan1)
			defer shutdown()

			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithFilter(reconcilertesting.EventSource, eventTypeToSubscribe), reconcilertesting.WithWebhookForNats)
			subscription.Spec.Sink = url
			ensureSubscriptionCreated(ctx, subscription)

			getSubscription(ctx, subscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionTrue, "")),
				reconcilertesting.HaveSubsConfiguration(&eventingv1alpha1.SubscriptionConfig{
					MaxInFlightMessages: defaultSubsConfig.MaxInFlightMessages,
				}),
			))

			subName2 := fmt.Sprintf("subb-%d", id)

			subChan2 := make(chan []byte)
			url2, close := newSubscriber(subChan2)
			defer close()

			if eventTypePrefix != "" {
				subscription2 = reconcilertesting.NewSubscription(subName2, namespaceName, reconcilertesting.WithFilter(reconcilertesting.EventSource, eventType), reconcilertesting.WithWebhookForNats)
			} else {
				subscription2 = reconcilertesting.NewSubscription(subName2, namespaceName, reconcilertesting.WithFilter(reconcilertesting.EventSource, emptyEventType), reconcilertesting.WithWebhookForNats)
			}

			subscription2.Spec.Sink = url2
			ensureSubscriptionCreated(ctx, subscription2)

			getSubscription(ctx, subscription2).Should(And(
				reconcilertesting.HaveSubscriptionName(subName2),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionTrue, "")),
				reconcilertesting.HaveSubsConfiguration(&eventingv1alpha1.SubscriptionConfig{
					MaxInFlightMessages: defaultSubsConfig.MaxInFlightMessages,
				}),
			))

			connection, err := connectToNats(natsURL)
			Expect(err).ShouldNot(HaveOccurred())
			err = connection.Publish(natsSubjectToPublish, []byte(reconcilertesting.StructuredCloudEvent))
			Expect(err).ShouldNot(HaveOccurred())

			if eventTypePrefix != "" {
				err = connection.Publish(natsSubject, []byte(reconcilertesting.StructuredCloudEventUpdated))
				Expect(err).ShouldNot(HaveOccurred())
			} else {
				err = connection.Publish(emptyNatsSubject, []byte(reconcilertesting.StructuredCloudEventUpdated))
				Expect(err).ShouldNot(HaveOccurred())
			}

			// make sure that the subscriber received the message
			sent := fmt.Sprintf(`"%s"`, reconcilertesting.EventData)
			Eventually(func() ([]byte, error) {
				return getFromChanOrTimeout(subChan1, smallPollingInterval)
			}, timeout, pollingInterval).Should(WithTransform(bytesStringer, Equal(sent)))

			Eventually(func() ([]byte, error) {
				return getFromChanOrTimeout(subChan2, smallPollingInterval)
			}, timeout, pollingInterval).Should(WithTransform(bytesStringer, Equal(sent)))
		})
	})
}

var (
	_ = Describe("NATS Subscription reconciler tests with non-empty eventTypePrefix", testExecutor(reconcilertesting.EventTypePrefix, reconcilertesting.OrderCreatedEventType, reconcilertesting.OrderCreatedEventTypeNotClean))
	_ = Describe("NATS Subscription reconciler tests with empty eventTypePrefix", testExecutor(reconcilertesting.EventTypePrefixEmpty, reconcilertesting.OrderCreatedEventTypePrefixEmpty, reconcilertesting.OrderCreatedEventTypeNotCleanPrefixEmpty))
)

func testExecutor(eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) func() {
	return func() {

		for _, tc := range reconcilerTestCases {
			tc(testID, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe)
			testID++
		}

		for _, tc := range dispatcherTestCases {
			tc(testID, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe)
			testID++
		}
	}
}

// getK8sEvents returns all kubernetes events for the given namespace.
// The result can be used in a gomega assertion.
func getK8sEvents(eventList *v1.EventList, namespace string) AsyncAssertion {
	ctx := context.TODO()
	return Eventually(func() v1.EventList {
		err := k8sClient.List(ctx, eventList, client.InNamespace(namespace))
		if err != nil {
			return v1.EventList{}
		}
		return *eventList
	})
}

func newSubscriber(result chan []byte) (string, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		result <- body
	}))
	return server.URL, server.Close
}

func connectToNats(natsURL string) (*nats.Conn, error) {
	connection, err := nats.Connect(natsURL, nats.RetryOnFailedConnect(true), nats.MaxReconnects(3), nats.ReconnectWait(time.Second))
	if err != nil {
		return nil, err
	}
	if connection.Status() != nats.CONNECTED {
		return nil, err
	}
	return connection, nil
}

func getFromChanOrTimeout(ch <-chan []byte, t time.Duration) ([]byte, error) {
	select {
	case received := <-ch:
		return received, nil
	case <-time.After(t):
		return nil, fmt.Errorf("timed out waiting for a message")
	}
}

func bytesStringer(bs []byte) string {
	return string(bs)
}

func ensureSubscriptionCreated(ctx context.Context, subscription *eventingv1alpha1.Subscription) {
	By(fmt.Sprintf("Ensuring the test namespace %q is created", subscription.Namespace))
	if subscription.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			err := k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				Expect(err).ShouldNot(HaveOccurred())
			}
		}
	}

	By(fmt.Sprintf("Ensuring the subscription %q is created", subscription.Name))
	// create subscription
	err := k8sClient.Create(ctx, subscription)
	Expect(err).Should(BeNil())
}

func fixtureNamespace(name string) *v1.Namespace {
	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return &namespace
}

func subscriptionGetter(ctx context.Context, name, namespace string) func() (*eventingv1alpha1.Subscription, error) {
	return func() (*eventingv1alpha1.Subscription, error) {
		lookupKey := types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}
		subscription := &eventingv1alpha1.Subscription{}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return &eventingv1alpha1.Subscription{}, err
		}
		return subscription, nil
	}
}

// getSubscription fetches a subscription using the lookupKey and allows making assertions on it
func getSubscription(ctx context.Context, subscription *eventingv1alpha1.Subscription) AsyncAssertion {
	return Eventually(func() *eventingv1alpha1.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return &eventingv1alpha1.Subscription{}
		}
		return subscription
	}, smallTimeout, smallPollingInterval)
}

// isSubscriptionDeleted checks a subscription is deleted and allows making assertions on it
func isSubscriptionDeleted(ctx context.Context, subscription *eventingv1alpha1.Subscription) AsyncAssertion {
	return Eventually(func() bool {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return k8serrors.IsNotFound(err)
		}
		return false
	}, smallTimeout, smallPollingInterval)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Test Suite setup ////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// These tests use Ginkgo (BDD-style Go controllertesting framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

// TODO: make configurable
const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
)

var testID int
var natsURL string
var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var natsServer *natsserver.Server
var defaultSubsConfig = env.DefaultSubscriptionConfig{MaxInFlightMessages: 1}
var reconciler *Reconciler
var natsBackend *handlers.Nats
var cancel context.CancelFunc

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t, "NATS Controller Suite", []Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	By("bootstrapping test environment")
	natsServer, natsURL = startNATS(natsPort)
	useExistingCluster := useExistingCluster
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../../", "config", "crd", "bases"),
			filepath.Join("../../../", "config", "crd", "external"),
		},
		AttachControlPlaneOutput: attachControlPlaneOutput,
		UseExistingCluster:       &useExistingCluster,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	reconcilertesting.ShutDownNATSServer(natsServer)
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
}, 60)

func startNATS(port int) (*natsserver.Server, string) {
	natsServer := reconcilertesting.RunNatsServerOnPort(port)
	clientURL := natsServer.ClientURL()
	log.Printf("NATS server started %v", clientURL)
	return natsServer, clientURL
}

func startReconciler(eventTypePrefix string, sinkValidator sinkValidator) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: ":7070",
	})
	Expect(err).ToNot(HaveOccurred())

	envConf := env.NatsConfig{
		URL:             natsURL,
		MaxReconnects:   10,
		ReconnectWait:   time.Second,
		EventTypePrefix: eventTypePrefix,
	}

	// prepare application-lister
	app := applicationtest.NewApplication(reconcilertesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), app)

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	Expect(err).To(BeNil())

	reconciler = NewReconciler(
		ctx,
		k8sManager.GetClient(),
		applicationLister,
		k8sManager.GetCache(),
		defaultLogger,
		k8sManager.GetEventRecorderFor("eventing-controller-nats"),
		envConf,
		defaultSubsConfig,
	)
	reconciler.sinkValidator = sinkValidator

	err = reconciler.SetupUnmanaged(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	natsBackend = reconciler.Backend.(*handlers.Nats)

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	return cancel
}

// ensureSubscriberSvcCreated creates a Service in the k8s cluster. If a custom namespace is used, it will be created as well.
func ensureSubscriberSvcCreated(ctx context.Context, svc *v1.Service) {
	By(fmt.Sprintf("Ensuring the test namespace %q is created", svc.Namespace))
	if svc.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(svc.Namespace)
		if namespace.Name != "default" {
			err := k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				Expect(err).ShouldNot(HaveOccurred())
			}
		}
	}

	By(fmt.Sprintf("Ensuring the subscriber service %q is created", svc.Name))
	// create subscription
	err := k8sClient.Create(ctx, svc)
	Expect(err).Should(BeNil())
}

func getSubscriptionFromNats(subscriptionMap map[string]*nats.Subscription, subscriptionName string) *nats.Subscription {
	i := 0
	for key, subscription := range subscriptionMap {
		if strings.Contains(key, subscriptionName) {
			return subscription
		}
		i++
	}
	return nil
}
