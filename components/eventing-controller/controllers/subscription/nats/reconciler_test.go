package nats

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"
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

	smallTimeout         = 10 * time.Second
	smallPollingInterval = 1 * time.Second

	timeout         = 60 * time.Second
	pollingInterval = 5 * time.Second

	namespaceName          = "test"
	subscriptionNameFormat = "nats-sub-%d"
	subscriberNameFormat   = "subscriber-%d"
)



// testNATSUnavailabilityReflectedInSubscriptionStatus tests if the reconciler can correctly resolve a Subscription in
// the case of a NATS server that becomes unavailable.
// The test is conducted in the following steps:
// 1. the NATS server that is available
// 2. the NATS server that is no longer available
// 3. the NATS server is available again
func testNATSUnavailabilityReflectedInSubscriptionStatus(id int, eventTypePrefix, _, eventTypeToSubscribe string) bool {
	return When("NATS server is not reachable and max retries are exceeded", func() {
		It("should mark the subscription as not ready until NATS is reachable again", func() {
			ctx := context.Background()
			natsPort := natsPort + id
			natsServer, natsURL := startNATS(natsPort)
			defer reconcilertesting.ShutDownNATSServer(natsServer)
			cancel = startReconciler(eventTypePrefix, natsURL)
			defer cancel()

			// Create a subscriber service
			subscriberName := fmt.Sprintf(subscriberNameFormat, id) + "-valid"
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subscriberName, namespaceName)
			ensureSubscriberSvcCreated(ctx, subscriberSvc)

			// Create a subscription
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id) + "-valid"
			subscription := reconcilertesting.NewSubscription(subscriptionName, subscriberSvc.Namespace,
				reconcilertesting.WithFilter("", eventTypeToSubscribe),
				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
			)
			ensureSubscriptionCreated(ctx, subscription)

			Context("testing against an available NATS server", func() {
				expectedConditions := eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionTrue, "")
				getSubscription(ctx, subscription).Should(And(
					reconcilertesting.HaveSubscriptionName(subscriptionName),
					reconcilertesting.HaveCondition(expectedConditions),
				))
			})

			Context("testing against an unavailable NATS server", func() {
				natsServer.Shutdown()
				getSubscription(ctx, subscription, timeout, pollingInterval).Should(And(
					reconcilertesting.HaveSubscriptionName(subscriptionName),
					reconcilertesting.HaveSubscriptionNotReady()),
				)
			})

			Context("testing against a NATS server that is available again", func() {
				_, _ = startNATS(natsPort)
				getSubscription(ctx, subscription, timeout, pollingInterval).Should(And(
					reconcilertesting.HaveSubscriptionName(subscriptionName),
					reconcilertesting.HaveSubscriptionReady()),
				)
			})
		})
	})
}

// testCleanEventTypes tests if the reconciler can resolve the correct cleanEventTypes from the corresponding filters
// of a Subscription.
// The test is conducted by changing the filters of a Subscription in the following steps:
// 1. no filters
// 2. adding two filters
// 3. changing one of the filters
// 4. deleting one of the filters
func testCleanEventTypes(id int, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("changing the filters of a subscription", func() {
		It("should have the correct corresponding clean event types", func() {
			// Set the expectations that are common to all test steps
			expectedCondition := eventingv1alpha1.MakeCondition(
				eventingv1alpha1.ConditionSubscriptionActive,
				eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
				v1.ConditionTrue, "")
			expectedConfiguration := &eventingv1alpha1.SubscriptionConfig{
				MaxInFlightMessages: defaultSubsConfig.MaxInFlightMessages}

			ctx := context.Background()
			cancel = startReconciler(eventTypePrefix, natsURL)
			defer cancel()

			// Create a subscriber service
			subscriberName := fmt.Sprintf(subscriberNameFormat, id)
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subscriberName, namespaceName)
			ensureSubscriberSvcCreated(ctx, subscriberSvc)

			// Create a Subscription
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)
			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
				reconcilertesting.WithEmptyFilter(),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
			)
			ensureSubscriptionCreated(ctx, subscription)

			Context("A subscription without filters", func() {
				By("checking the expected status of the subscription")
				getSubscription(ctx, subscription).Should(And(
					reconcilertesting.HaveSubscriptionName(subscriptionName),
					reconcilertesting.HaveCondition(expectedCondition),
					reconcilertesting.HaveSubsConfiguration(expectedConfiguration),
				))

				By("especially checking that the subscription has no clean event types")
				getSubscription(ctx, subscription).Should(reconcilertesting.HaveCleanEventTypes(nil))
			})

			Context("A subscription without filters to which filters are added", func() {
				expectedCleanEventTypes := []string{
					fmt.Sprintf("%s0", natsSubjectToPublish),
					fmt.Sprintf("%s1", natsSubjectToPublish),
				}

				By("adding filters to a subscription")
				eventTypes := []string{
					fmt.Sprintf("%s0", eventTypeToSubscribe),
					fmt.Sprintf("%s1", eventTypeToSubscribe),
				}
				for _, eventType := range eventTypes {
					reconcilertesting.AddFilter(reconcilertesting.EventSource, eventType, subscription)
				}
				ensureSubscriptionUpdated(ctx, subscription)

				By("checking the expected status of the subscription")
				getSubscription(ctx, subscription).Should(And(
					reconcilertesting.HaveSubscriptionName(subscriptionName)),
					reconcilertesting.HaveSubsConfiguration(expectedConfiguration),
					reconcilertesting.HaveCondition(expectedCondition),
				)

				By("checking that the clean event types correspond to the added filters")
				getSubscription(ctx, subscription).Should(reconcilertesting.HaveCleanEventTypes(expectedCleanEventTypes))
			})

			Context("A subscription with filters that are being modified", func() {
				expectedCleanEventTypes := []string{
					fmt.Sprintf("%s0alpha", natsSubjectToPublish),
					fmt.Sprintf("%s1alpha", natsSubjectToPublish),
				}

				By("modifying the existing filters")
				for _, f := range subscription.Spec.Filter.Filters {
					f.EventType.Value = fmt.Sprintf("%salpha", f.EventType.Value)
				}
				ensureSubscriptionUpdated(ctx, subscription)

				By("checking the expected status if the subscription")
				getSubscription(ctx, subscription).Should(And(
					reconcilertesting.HaveSubscriptionName(subscriptionName)),
					reconcilertesting.HaveSubsConfiguration(expectedConfiguration),
					reconcilertesting.HaveCondition(expectedCondition),
				)

				By("checking that the clean event types correspond to the modified")
				getSubscription(ctx, subscription).Should(reconcilertesting.HaveCleanEventTypes(expectedCleanEventTypes))
			})

			Context("A subscription with filters of which one is getting deleted", func() {
				expectedCleanEventTypes := []string{
					fmt.Sprintf("%s0alpha", natsSubjectToPublish),
				}

				By("deleting one if the filters")
				subscription.Spec.Filter.Filters = subscription.Spec.Filter.Filters[:1]
				ensureSubscriptionUpdated(ctx, subscription)

				By("checking the expected status if the subscription")
				getSubscription(ctx, subscription).Should(And(
					reconcilertesting.HaveSubscriptionName(subscriptionName)),
					reconcilertesting.HaveSubsConfiguration(expectedConfiguration),
					reconcilertesting.HaveCondition(expectedCondition),
				)

				By("checking that the clean event types correspond to the modified")
				getSubscription(ctx, subscription).Should(reconcilertesting.HaveCleanEventTypes(expectedCleanEventTypes))
			})
		})
	})
}

// testUpdateSubscriptionStatus tests if the reconciler can create and update the Status of a Subscription as expected.
// This especially tests that the subscription does not have multiple conditions after modifying and updating it.
func testUpdateSubscriptionStatus(id int, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("updating the clean event types in the subscription status", func() {
		It("should mark the subscription as ready", func() {
			expectedCondition := eventingv1alpha1.MakeCondition(
				eventingv1alpha1.ConditionSubscriptionActive,
				eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
				v1.ConditionTrue, "")
			expectedConfiguration := &eventingv1alpha1.SubscriptionConfig{
				MaxInFlightMessages: defaultSubsConfig.MaxInFlightMessages}

			ctx := context.Background()
			cancel = startReconciler(eventTypePrefix, natsURL)
			defer cancel()

			// Create a subscriber service
			subscriberName := fmt.Sprintf(subscriberNameFormat, id)
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subscriberName, namespaceName)
			ensureSubscriberSvcCreated(ctx, subscriberSvc)

			// Create a subscription
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)
			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
				reconcilertesting.WithEmptyFilter(),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithMultipleConditions(),
				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
			)
			multipleConditions := subscription.Status.Conditions
			ensureSubscriptionCreated(ctx, subscription)

			Context("Checking the subscription got created properly", func() {
				getSubscription(ctx, subscription).Should(And(
					reconcilertesting.HaveSubscriptionName(subscriptionName),
					reconcilertesting.HaveCleanEventTypes(nil),
					reconcilertesting.HaveCondition(expectedCondition),
					reconcilertesting.HaveSubsConfiguration(expectedConfiguration),
					reconcilertesting.HaveSubscriptionReady(),
				))
			})

			Context("Changing an existing subscription and checking, that it got updated properly", func() {
				By("adding a filter")
				expectedCleanEventType := []string{fmt.Sprintf("%stest", natsSubjectToPublish)}

				eventType := fmt.Sprintf("%stest", eventTypeToSubscribe)
				reconcilertesting.AddFilter(reconcilertesting.EventSource, eventType, subscription)
				ensureSubscriptionUpdated(ctx, subscription)

				By("checking, the subscription has the expected Status")
				getSubscription(ctx, subscription).Should(And(
					reconcilertesting.HaveSubscriptionName(subscriptionName),
					reconcilertesting.HaveCondition(expectedCondition),
					reconcilertesting.HaveSubsConfiguration(expectedConfiguration),
					reconcilertesting.HaveSubscriptionReady(),
					reconcilertesting.HaveCleanEventTypes(expectedCleanEventType),
				))

				By("ensuring, that the subscription does not have additional conditions")
				getSubscription(ctx, subscription).ShouldNot(And(
					reconcilertesting.HaveCondition(multipleConditions[0]),
					reconcilertesting.HaveCondition(multipleConditions[1]),
				))
			})
		})
	})
}

// testCreateDeleteSubscription tests if a subscription, after getting created correctly, can be deleted properly.
func testCreateDeleteSubscription(id int, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("creating and then deleting a subscription", func() {
		It("the subscription should get reconciled properly", func() {

			ctx := context.Background()
			cancel = startReconciler(eventTypePrefix, natsURL)
			defer cancel()

			// Create a subscriber svc
			subscriberName := fmt.Sprintf(subscriberNameFormat, id)
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subscriberName, namespaceName)
			ensureSubscriberSvcCreated(ctx, subscriberSvc)

			// Create a subscription
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)
			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
				reconcilertesting.WithFilter(reconcilertesting.EventSource, eventTypeToSubscribe),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
			)
			ensureSubscriptionCreated(ctx, subscription)

			By("checking that the subscription was created properly")
			expectedCondition := eventingv1alpha1.MakeCondition(
				eventingv1alpha1.ConditionSubscriptionActive,
				eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
				v1.ConditionTrue, "")

			expectedConfiguration := &eventingv1alpha1.SubscriptionConfig{
				MaxInFlightMessages: defaultSubsConfig.MaxInFlightMessages}

			getSubscription(ctx, subscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(expectedCondition),
				reconcilertesting.HaveSubsConfiguration(expectedConfiguration),
			))

			By("checking that the subscription was created at the NATS backend")
			backendSubscription := getSubscriptionFromNats(natsBackend.GetAllSubscriptions(), subscriptionName)
			Expect(backendSubscription).NotTo(BeNil())
			Expect(backendSubscription.IsValid()).To(BeTrue())
			Expect(backendSubscription.Subject).Should(Equal(natsSubjectToPublish))

			By("checking, that the subscription gets deleted properly")
			Expect(k8sClient.Delete(ctx, subscription)).Should(BeNil())
			isSubscriptionDeleted(ctx, subscription).Should(BeTrue())
		})
	})
}

// testCreateSubscriptionWithValidSink tests if a subscription with a valid sink can get resolved correctly.
func testCreateSubscriptionWithValidSink(id int, eventTypePrefix, _, eventTypeToSubscribe string) bool {
	subscriptionName := fmt.Sprintf(subscriptionNameFormat, id) + "-valid"
	subscriberName := fmt.Sprintf(subscriberNameFormat, id) + "-valid"
	sink := reconcilertesting.ValidSinkURL(namespaceName, subscriberName)

	testCreatingSubscription := func(sink string) {
		ctx := context.Background()
		cancel = startReconciler(eventTypePrefix, natsURL)
		defer cancel()

		// Create a subscriber service
		subscriberSvc := reconcilertesting.NewSubscriberSvc(subscriberName, namespaceName)
		ensureSubscriberSvcCreated(ctx, subscriberSvc)

		// Create a subscription
		subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
			reconcilertesting.WithFilter("", eventTypeToSubscribe),
			reconcilertesting.WithSinkURL(sink),
		)
		ensureSubscriptionCreated(ctx, subscription)

		// Validate the subscription
		expectedConditions := eventingv1alpha1.MakeCondition(
			eventingv1alpha1.ConditionSubscriptionActive,
			eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
			v1.ConditionTrue, "")
		getSubscription(ctx, subscription).Should(And(
			reconcilertesting.HaveSubscriptionName(subscriptionName),
			reconcilertesting.HaveCondition(expectedConditions),
			reconcilertesting.HaveSubscriptionReady(),
		))

		Expect(k8sClient.Delete(ctx, subscription)).Should(BeNil())
		isSubscriptionDeleted(ctx, subscription).Should(reconcilertesting.HaveNotFoundSubscription(true))

		Expect(k8sClient.Delete(ctx, subscriberSvc)).Should(BeNil())
	}
	return When("Create subscription with valid sink", func() {
		It("Should mark the Subscription with a valid sink as ready", func() {
			testCreatingSubscription(sink)
		})

		It("Should mark the subscription with a valid sink that contains the port suffix as ready", func() {
			testCreatingSubscription(sink + ":8080")
		})

		It("Should mark the subscription with a valid sink that contains the port suffix and a path as ready", func() {
			testCreatingSubscription(sink + ":8080" + "/myEndpoint")
		})
	})
}

// testCreateSubscriptionWithInvalidSink tests if a subscription with an invalid sink can get resolved correctly.
func testCreateSubscriptionWithInvalidSink(id int, eventTypePrefix, _, eventTypeToSubscribe string) bool {
	invalidSinkMsgCheck := func(sink, subConditionMsg, k8sEventMsg string) {
		ctx := context.Background()
		cancel = startReconciler(eventTypePrefix, natsURL)
		defer cancel()
		subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)

		// Create a subscription
		subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
			reconcilertesting.WithFilter(reconcilertesting.EventSource, eventTypeToSubscribe),
			reconcilertesting.WithWebhookForNATS(),
			reconcilertesting.WithSinkURL(sink),
		)
		ensureSubscriptionCreated(ctx, subscription)

		// Validate the subscription
		expectedCondition := eventingv1alpha1.MakeCondition(
			eventingv1alpha1.ConditionSubscriptionActive,
			eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
			v1.ConditionFalse, subConditionMsg)
		getSubscription(ctx, subscription).Should(And(
			reconcilertesting.HaveSubscriptionName(subscriptionName),
			reconcilertesting.HaveCondition(expectedCondition),
		))

		var subscriptionEvents = v1.EventList{}
		subscriptionEvent := v1.Event{
			Reason:  string(events.ReasonValidationFailed),
			Message: k8sEventMsg,
			Type:    v1.EventTypeWarning,
		}
		getK8sEvents(&subscriptionEvents, subscription.Namespace).Should(reconcilertesting.HaveEvent(subscriptionEvent))

		Expect(k8sClient.Delete(ctx, subscription)).Should(BeNil())
		isSubscriptionDeleted(ctx, subscription).Should(reconcilertesting.HaveNotFoundSubscription(true))
	}

	return When("Create subscription with invalid sink", func() {
		It("Should mark the Subscription as not ready if sink URL scheme is not 'http' or 'https'", func() {
			invalidSinkMsgCheck(
				"invalid",
				"sink URL scheme should be 'http' or 'https'",
				"Sink URL scheme should be HTTP or HTTPS: invalid",
			)
		})
		It("Should mark the subscription as not ready if sink contains invalid characters", func() {
			invalidSinkMsgCheck(
				"http://127.0.0. 1",
				"not able to parse sink url with error: parse \"http://127.0.0. 1\": invalid character \" \" in host name",
				"Not able to parse Sink URL with error: parse \"http://127.0.0. 1\": invalid character \" \" in host name",
			)
		})

		It("Should mark the subscription as not ready if sink does not contain suffix 'svc.cluster.local'", func() {
			invalidSinkMsgCheck(
				"http://127.0.0.1",
				"sink does not contain suffix: svc.cluster.local in the URL",
				"Sink does not contain suffix: svc.cluster.local",
			)
		})

		It("Should mark the subscription as not ready if sink does not contain 5 sub-domains", func() {
			invalidSinkMsgCheck(
				fmt.Sprintf("https://%s.%s.%s.svc.cluster.local", "testapp", "testsub", "test"),
				"sink should contain 5 sub-domains: testapp.testsub.test.svc.cluster.local",
				"Sink should contain 5 sub-domains: testapp.testsub.test.svc.cluster.local",
			)
		})

		It("Should mark the subscription as not ready if sink points to different namespace", func() {
			invalidSinkMsgCheck(
				fmt.Sprintf("https://%s.%s.svc.cluster.local", "testapp", "test-ns"),
				"namespace of subscription: test and the namespace of subscriber: test-ns are different",
				"natsNamespace of subscription: test and the subscriber: test-ns are different",
			)
		})

		It("Should mark the subscription as not ready if sink is not a valid cluster local service", func() {
			invalidSinkMsgCheck(
				reconcilertesting.ValidSinkURL(namespaceName, "testapp"),
				"sink is not valid cluster local svc, failed with error: Service \"testapp\" not found",
				"Sink does not correspond to a valid cluster local svc",
			)
		})
	})
}

// testCreateSubscriptionWithEmptyProtocolProtocolSettingsDialect
func testCreateSubscriptionWithEmptyProtocolProtocolSettingsDialect(id int, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("Create subscription with empty protocol, protocolsettings and dialect", func() {
		It("should mark the subscription as ready", func() {
			ctx := context.Background()
			cancel = startReconciler(eventTypePrefix, natsURL)
			defer cancel()
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)
			subscriberName := fmt.Sprintf(subscriberNameFormat, id)

			// Create a subscriber service
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subscriberName, namespaceName)
			ensureSubscriberSvcCreated(ctx, subscriberSvc)

			// Create a subscription
			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
				reconcilertesting.WithFilter("", eventTypeToSubscribe),
				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
			)
			ensureSubscriptionCreated(ctx, subscription)

			// Validating the Subscription
			expectedCondition := eventingv1alpha1.MakeCondition(
				eventingv1alpha1.ConditionSubscriptionActive,
				eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
				v1.ConditionTrue, "")
			getSubscription(ctx, subscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(expectedCondition),
			))

			// Check that the subscription was created at the NATS backend
			backendSubscription := getSubscriptionFromNats(natsBackend.GetAllSubscriptions(), subscriptionName)
			Expect(backendSubscription).NotTo(BeNil())
			Expect(backendSubscription).To(reconcilertesting.BeValid())
			Expect(backendSubscription.Subject).Should(Equal(natsSubjectToPublish))

			Expect(backendSubscription).To(And(
				reconcilertesting.BeNotNil(),
				reconcilertesting.BeValid(),
				reconcilertesting.HaveSubject(natsSubjectToPublish),
			))

		})
	})
}

// testChangeSubscriptionConfiguration tests if changes to the configuration of an existing subscription get resolved
// properly.
func testChangeSubscriptionConfiguration(id int, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("Change Subscription configuration", func() {
		It("should reflect the new config in the subscription status", func() {
			By("creating the subscription using the default config")
			ctx := context.Background()
			cancel = startReconciler(eventTypePrefix, natsURL)
			defer cancel()
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)
			subscriberName := fmt.Sprintf(subscriberNameFormat, id)

			// Create a subscriber service
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subscriberName, namespaceName)
			ensureSubscriberSvcCreated(ctx, subscriberSvc)

			// Create a subscription
			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
				reconcilertesting.WithFilter(reconcilertesting.EventSource, eventTypeToSubscribe),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
			)
			ensureSubscriptionCreated(ctx, subscription)

			// Validating the Subscription
			expectedConditions := eventingv1alpha1.MakeCondition(
				eventingv1alpha1.ConditionSubscriptionActive,
				eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
				v1.ConditionTrue, "",
			)
			expectedConfiguration := &eventingv1alpha1.SubscriptionConfig{
				MaxInFlightMessages: defaultSubsConfig.MaxInFlightMessages,
			}
			getSubscription(ctx, subscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(expectedConditions),
				reconcilertesting.HaveSubsConfiguration(expectedConfiguration),
			))

			By("updating the subscription configuration in the spec")
			newMaxInFlight := defaultSubsConfig.MaxInFlightMessages + 1
			changedSub := subscription.DeepCopy()
			changedSub.Spec.Config = &eventingv1alpha1.SubscriptionConfig{
				MaxInFlightMessages: newMaxInFlight,
			}
			Expect(k8sClient.Update(ctx, changedSub)).Should(BeNil())

			// Validating the updated Subscription
			expectedUpgradedConfiguration := &eventingv1alpha1.SubscriptionConfig{
				MaxInFlightMessages: newMaxInFlight,
			}
			getSubscription(ctx, subscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(expectedConditions),
				reconcilertesting.HaveSubsConfiguration(expectedUpgradedConfiguration),
			))

			// Check that the subscription was created at the NATS backend
			backendSubscription := getSubscriptionFromNats(natsBackend.GetAllSubscriptions(), subscriptionName)
			Expect(backendSubscription).NotTo(BeNil())
			Expect(backendSubscription.IsValid()).To(BeTrue())
			Expect(backendSubscription.Subject).Should(Equal(natsSubjectToPublish))

			// Clean everything up
			Expect(k8sClient.Delete(ctx, subscription)).Should(BeNil())
			isSubscriptionDeleted(ctx, subscription).Should(reconcilertesting.HaveNotFoundSubscription(true))
		})
	})
}

// testCreateSubscriptionWithEmptyEventType tests if a subscription with a filter, that is missing an event type,
// gets resolved correctly as faulty.
func testCreateSubscriptionWithEmptyEventType(id int, eventTypePrefix, _, _ string) bool {
	return When("Create subscription with empty event type", func() {
		It("should mark the subscription as not ready", func() {
			ctx := context.Background()
			cancel = startReconciler(eventTypePrefix, natsURL)
			defer cancel()
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)
			subscriberName := fmt.Sprintf(subscriberNameFormat, id)

			// Create a subscriber service
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subscriberName, namespaceName)
			ensureSubscriberSvcCreated(ctx, subscriberSvc)

			// Create a subscription
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
				reconcilertesting.WithFilter(reconcilertesting.EventSource, ""),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
			)
			ensureSubscriptionCreated(ctx, givenSubscription)

			// Validate the subscription
			expectedConditions := eventingv1alpha1.MakeCondition(
				eventingv1alpha1.ConditionSubscriptionActive,
				eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
				v1.ConditionFalse, nats.ErrBadSubject.Error(),
			)
			getSubscription(ctx, givenSubscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(expectedConditions),
			))
		})
	})
}

type testCase func(id int, eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) bool

var (
	reconcilerTestCases = []testCase{
		testCreateDeleteSubscription,
		testCreateSubscriptionWithValidSink,
		testCreateSubscriptionWithInvalidSink,
		testCreateSubscriptionWithEmptyProtocolProtocolSettingsDialect,
		testChangeSubscriptionConfiguration,
		testCreateSubscriptionWithEmptyEventType,
		testCleanEventTypes,
		testUpdateSubscriptionStatus,
		testNATSUnavailabilityReflectedInSubscriptionStatus,
	}
)
var (
	_ = Describe("NATS subscription reconciler tests with non-empty eventTypePrefix", testExecutor(reconcilertesting.EventTypePrefix, reconcilertesting.OrderCreatedEventType, reconcilertesting.OrderCreatedEventTypeNotClean))
	_ = Describe("NATS subscription reconciler tests with empty eventTypePrefix", testExecutor(reconcilertesting.EventTypePrefixEmpty, reconcilertesting.OrderCreatedEventTypePrefixEmpty, reconcilertesting.OrderCreatedEventTypeNotCleanPrefixEmpty))
)

func testExecutor(eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) func() {
	return func() {

		for _, tc := range reconcilerTestCases {
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
	}, smallTimeout, smallPollingInterval)
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

func ensureSubscriptionUpdated(ctx context.Context, subscription *eventingv1alpha1.Subscription) {
	By(fmt.Sprintf("Ensuring the subscription %q is updated", subscription.Name))
	// create subscription
	err := k8sClient.Update(ctx, subscription)
	Expect(err).Should(BeNil())
}

func fixtureNamespace(name string) *v1.Namespace {
	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "natsNamespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return &namespace
}

// getSubscription fetches a subscription using the lookupKey and allows making assertions on it
func getSubscription(ctx context.Context, subscription *eventingv1alpha1.Subscription, intervals ...interface{}) AsyncAssertion {
	if len(intervals) == 0 {
		intervals = []interface{}{smallTimeout, smallPollingInterval}
	}
	return Eventually(func() *eventingv1alpha1.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return &eventingv1alpha1.Subscription{}
		}
		return subscription
	}, intervals...)
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

// //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Test Suite setup ////////////////////////////////////////////////////////////////////////////////////////////////////
// //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
var defaultSubsConfig = env.DefaultSubscriptionConfig{MaxInFlightMessages: 1, DispatcherRetryPeriod: time.Second, DispatcherMaxRetries: 1}
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

func startReconciler(eventTypePrefix string, natsURL string) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: "localhost:7070",
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
		defaultLogger,
		k8sManager.GetEventRecorderFor("eventing-controller-nats"),
		envConf,
		defaultSubsConfig,
	)

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
