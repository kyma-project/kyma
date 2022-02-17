// This file contains unit tests for the NATS subscription reconciler.
// It uses the testing.T and gomega testing libraries to perform the assertions.
// TestEnvironment struct mocks the required resources.
package nats

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/mocks"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"k8s.io/apimachinery/pkg/types"
)

const namespaceName = "test"

var testEnv *envtest.Environment

func TestHandleSubscriptionDeletion(t *testing.T) {
	g := NewGomegaWithT(t)
	testEnvironment := setupTestEnvironment(t)
	ctx, r, mockedBackend := testEnvironment.Context, testEnvironment.Reconciler, testEnvironment.Backend

	testCases := []struct {
		name               string
		givenFinalizers    []string
		expectedDeleteCall bool
		expectedFinalizers []string
	}{
		{
			name:               "With no finalizers the NATS subscription should not be deleted",
			givenFinalizers:    []string{},
			expectedDeleteCall: false,
			expectedFinalizers: []string{},
		},
		{
			name:               "With eventing finalizer the NATS subscription should be deleted and the finalizer should be cleared",
			givenFinalizers:    []string{Finalizer},
			expectedDeleteCall: true,
			expectedFinalizers: []string{},
		},
		{
			name:               "With wrong finalizer the NATS subscription should not be deleted",
			givenFinalizers:    []string{"eventing2.kyma-project.io"},
			expectedDeleteCall: false,
			expectedFinalizers: []string{"eventing2.kyma-project.io"},
		},
	}

	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			// given
			subscription := NewSubscription(WithFinalizers(testCase.givenFinalizers))
			err := r.Client.Create(testEnvironment.Context, subscription)
			g.Expect(err).Should(BeNil())

			mockedBackend.On("DeleteSubscription", subscription).Return(nil)

			// when
			err = r.handleSubscriptionDeletion(ctx, subscription, r.namedLogger())
			g.Expect(err).Should(BeNil())

			// then
			if testCase.expectedDeleteCall {
				mockedBackend.AssertCalled(t, "DeleteSubscription", subscription)
			} else {
				mockedBackend.AssertNotCalled(t, "DeleteSubscription", subscription)
			}

			ensureFinalizerMatch(g, subscription, testCase.expectedFinalizers)

			// check the changes were made on the kubernetes server
			fetchedSub, err := fetchTestSubscription(ctx, r)
			g.Expect(err).Should(BeNil())
			ensureFinalizerMatch(g, &fetchedSub, testCase.expectedFinalizers)

			// clean up
			err = r.Client.Delete(ctx, subscription)
			g.Expect(err).Should(BeNil())
		})
	}
}

func TestSyncSubscriptionStatus(t *testing.T) {
	g := NewGomegaWithT(t)
	testEnvironment := setupTestEnvironment(t)
	ctx, r := testEnvironment.Context, testEnvironment.Reconciler

	message := "message is not required for tests"
	falseNatsSubActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
		corev1.ConditionFalse, message)
	trueNatsSubActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
		corev1.ConditionTrue, message)

	testCases := []struct {
		name               string
		givenSub           *eventingv1alpha1.Subscription
		givenNatsSubReady  bool
		givenForceStatus   bool
		expectedConditions []eventingv1alpha1.Condition
		expectedStatus     bool
	}{
		{
			name: "Subscription with no conditions should stay not ready with false condition",
			givenSub: NewSubscription(
				WithConditions([]eventingv1alpha1.Condition{}),
				WithStatus(true),
			),
			givenNatsSubReady:  false,
			givenForceStatus:   false,
			expectedConditions: []eventingv1alpha1.Condition{falseNatsSubActiveCondition},
			expectedStatus:     false,
		},
		{
			name: "Subscription with false ready condition should stay not ready with false condition",
			givenSub: NewSubscription(
				WithConditions([]eventingv1alpha1.Condition{falseNatsSubActiveCondition}),
				WithStatus(false),
			),
			givenNatsSubReady:  false,
			givenForceStatus:   false,
			expectedConditions: []eventingv1alpha1.Condition{falseNatsSubActiveCondition},
			expectedStatus:     false,
		},
		{
			name: "Subscription should become ready because of isNatsSubReady flag",
			givenSub: NewSubscription(
				WithConditions([]eventingv1alpha1.Condition{falseNatsSubActiveCondition}),
				WithStatus(false),
			),
			givenNatsSubReady:  true, // isNatsSubReady
			givenForceStatus:   false,
			expectedConditions: []eventingv1alpha1.Condition{trueNatsSubActiveCondition},
			expectedStatus:     true,
		},
		{
			name: "Subscription should stay with ready condition and status",
			givenSub: NewSubscription(
				WithConditions([]eventingv1alpha1.Condition{trueNatsSubActiveCondition}),
				WithStatus(true),
			),
			givenNatsSubReady:  true,
			givenForceStatus:   false,
			expectedConditions: []eventingv1alpha1.Condition{trueNatsSubActiveCondition},
			expectedStatus:     true,
		},
		{
			name: "Subscription should become not ready because of false isNatsSubReady flag",
			givenSub: NewSubscription(
				WithConditions([]eventingv1alpha1.Condition{trueNatsSubActiveCondition}),
				WithStatus(true),
			),
			givenNatsSubReady:  false, // isNatsSubReady
			givenForceStatus:   false,
			expectedConditions: []eventingv1alpha1.Condition{falseNatsSubActiveCondition},
			expectedStatus:     false,
		},
		{
			name: "Subscription should stay with the same condition, but still updated because of the forceUpdateStatus flag",
			givenSub: NewSubscription(
				WithConditions([]eventingv1alpha1.Condition{trueNatsSubActiveCondition}),
				WithStatus(true),
			),
			givenNatsSubReady:  true,
			givenForceStatus:   true,
			expectedConditions: []eventingv1alpha1.Condition{trueNatsSubActiveCondition},
			expectedStatus:     true,
		},
	}
	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			// given
			sub := testCase.givenSub
			err := r.Client.Create(ctx, sub)
			g.Expect(err).Should(BeNil())

			// when
			err = r.syncSubscriptionStatus(ctx, sub, testCase.givenNatsSubReady, testCase.givenForceStatus, message)
			g.Expect(err).To(BeNil())

			// then
			ensureSubscriptionMatchesConditionsAndStatus(g, *sub, testCase.expectedConditions, testCase.expectedStatus)

			// fetch the sub also from k8s server in order to check whether changes were done both in-memory and on k8s server
			fetchedSub, err := fetchTestSubscription(ctx, r)
			g.Expect(err).Should(BeNil())
			ensureSubscriptionMatchesConditionsAndStatus(g, fetchedSub, testCase.expectedConditions, testCase.expectedStatus)

			// clean up
			err = r.Client.Delete(ctx, sub)
			g.Expect(err).To(BeNil())
		})
	}
}

func TestDefaultSinkValidator(t *testing.T) {
	g := NewGomegaWithT(t)
	testEnvironment := setupTestEnvironment(t)
	ctx, r, log := testEnvironment.Context, testEnvironment.Reconciler, testEnvironment.Logger

	testCases := []struct {
		name                  string
		givenSubscriptionSink string
		givenSvcNameToCreate  string
		expectedErrString     string
	}{
		{
			name:                  "With invalid scheme",
			givenSubscriptionSink: "invalid Sink",
			expectedErrString:     "sink URL scheme should be 'http' or 'https'",
		},
		{
			name:                  "With invalid URL",
			givenSubscriptionSink: "http://invalid Sink",
			expectedErrString:     "not able to parse sink url with error",
		},
		{
			name:                  "With invalid suffix",
			givenSubscriptionSink: "https://svc2.test.local",
			expectedErrString:     "sink does not contain suffix",
		},
		{
			name:                  "With invalid suffix and port",
			givenSubscriptionSink: "https://svc2.test.local:8080",
			expectedErrString:     "sink does not contain suffix",
		},
		{
			name:                  "With invalid number of subdomains",
			givenSubscriptionSink: "https://svc.cluster.local:8080", // right suffix but 3 subdomains
			expectedErrString:     "sink should contain 5 sub-domains",
		},
		{
			name:                  "With different namespaces in subscription and sink name",
			givenSubscriptionSink: "https://eventing-nats.kyma-system.svc.cluster.local:8080", // sub is in test ns
			expectedErrString:     "namespace of subscription: test and the namespace of subscriber: kyma-system are different",
		},
		{
			name:                  "With no existing svc in the cluster",
			givenSubscriptionSink: "https://eventing-nats.test.svc.cluster.local:8080",
			expectedErrString:     "sink is not valid cluster local svc, failed with error",
		},
		{
			name:                  "With no existing svc in the cluster, service has the wrong name",
			givenSubscriptionSink: "https://eventing-nats.test.svc.cluster.local:8080",
			givenSvcNameToCreate:  "test", // wrong name
			expectedErrString:     "sink is not valid cluster local svc, failed with error",
		},
		{
			name:                  "With no existing svc in the cluster, service has the wrong name",
			givenSubscriptionSink: "https://eventing-nats.test.svc.cluster.local:8080",
			givenSvcNameToCreate:  "eventing-nats",
			expectedErrString:     "",
		},
	}
	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			// given
			sub := NewSubscription(
				WithConditions([]eventingv1alpha1.Condition{}),
				WithStatus(true),
				WithSink(testCase.givenSubscriptionSink),
			)

			// create the service if required for test
			if testCase.givenSvcNameToCreate != "" {
				svc := &corev1.Service{
					ObjectMeta: v1.ObjectMeta{
						Name:      testCase.givenSvcNameToCreate,
						Namespace: namespaceName,
					},
				}

				err := r.Client.Create(ctx, svc)
				g.Expect(err).To(BeNil())
			}

			// when
			// call the defaultSinkValidator function
			err := r.sinkValidator(ctx, r, sub)
			log.WithContext().Infof("Result of defaultSinkValidatorCall: %+v", err)

			// then
			// given error should match expected error
			if testCase.expectedErrString == "" {
				g.Expect(err).To(BeNil())
			} else {
				substringResult := strings.Contains(err.Error(), testCase.expectedErrString)
				g.Expect(substringResult).Should(BeTrue())
			}
		})
	}
}

// helper function and structs

// TestEnvironment provides mocked resources for tests.
type TestEnvironment struct {
	Context    context.Context
	Client     *client.WithWatch
	Backend    *mocks.MessagingBackend
	Reconciler *Reconciler
	Logger     *logger.Logger
	Recorder   *record.FakeRecorder
}

// setupTestEnvironment is a TestEnvironment constructor
func setupTestEnvironment(t *testing.T) *TestEnvironment {
	g := NewGomegaWithT(t)
	mockedBackend := &mocks.MessagingBackend{}
	ctx := context.Background()
	fakeClient := createFakeClient(g)
	recorder := &record.FakeRecorder{}

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	r := Reconciler{
		Backend:       mockedBackend,
		Client:        fakeClient,
		logger:        defaultLogger,
		recorder:      recorder,
		sinkValidator: defaultSinkValidator,
	}

	return &TestEnvironment{
		Context:    ctx,
		Client:     &fakeClient,
		Backend:    mockedBackend,
		Reconciler: &r,
		Logger:     defaultLogger,
		Recorder:   recorder,
	}
}

func createFakeClient(g *GomegaWithT) client.WithWatch {
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../", "config", "crd", "bases"),
			filepath.Join("../../", "config", "crd", "external"),
		},
	}
	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	g.Expect(err).Should(BeNil())
	return fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
}

func NewSubscription(options ...func(subscription *eventingv1alpha1.Subscription)) *eventingv1alpha1.Subscription {
	sub := &eventingv1alpha1.Subscription{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: namespaceName,
		},
	}
	for _, o := range options {
		o(sub)
	}
	return sub
}
func WithSink(sink string) func(subscription *eventingv1alpha1.Subscription) {
	return func(sub *eventingv1alpha1.Subscription) {
		sub.Spec.Sink = sink
	}
}
func WithConditions(conditions []eventingv1alpha1.Condition) func(subscription *eventingv1alpha1.Subscription) {
	return func(sub *eventingv1alpha1.Subscription) {
		sub.Status.Conditions = conditions
	}
}
func WithStatus(status bool) func(subscription *eventingv1alpha1.Subscription) {
	return func(sub *eventingv1alpha1.Subscription) {
		sub.Status.Ready = status
	}
}
func WithFinalizers(finalizers []string) func(subscription *eventingv1alpha1.Subscription) {
	return func(sub *eventingv1alpha1.Subscription) {
		sub.ObjectMeta.Finalizers = finalizers
	}
}

func fetchTestSubscription(ctx context.Context, r *Reconciler) (eventingv1alpha1.Subscription, error) {
	var fetchedSub eventingv1alpha1.Subscription
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: namespaceName,
	}, &fetchedSub)
	return fetchedSub, err
}

func ensureFinalizerMatch(g *GomegaWithT, subscription *eventingv1alpha1.Subscription, expectedFinalizers []string) {
	if len(expectedFinalizers) == 0 {
		g.Expect(subscription.ObjectMeta.Finalizers).Should(BeEmpty())
	} else {
		g.Expect(subscription.ObjectMeta.Finalizers).Should(Equal(expectedFinalizers))
	}
}

func ensureSubscriptionMatchesConditionsAndStatus(g *GomegaWithT, subscription eventingv1alpha1.Subscription, expectedConditions []eventingv1alpha1.Condition, expectedStatus bool) {
	g.Expect(len(expectedConditions)).Should(Equal(len(subscription.Status.Conditions)))
	for i := range expectedConditions {
		comparisonResult := equalsConditionsIgnoringTime(expectedConditions[i], subscription.Status.Conditions[i])
		g.Expect(comparisonResult).To(BeTrue())
	}
	g.Expect(expectedStatus).Should(Equal(subscription.Status.Ready))
}
