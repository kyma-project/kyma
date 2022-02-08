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

func TestHandleSubscriptionDeletion(t *testing.T) {
	g := NewGomegaWithT(t)
	testEnvironment := setupTestEnvironment(t)
	ctx, r, mockedBackend := testEnvironment.Context, testEnvironment.Reconciler, testEnvironment.Backend

	subscription := &eventingv1alpha1.Subscription{
		ObjectMeta: v1.ObjectMeta{
			Name:       "test",
			Namespace:  namespaceName,
			Finalizers: []string{}, // empty finalizers
		},
	}

	err := testEnvironment.Reconciler.Client.Create(testEnvironment.Context, subscription)
	g.Expect(err).Should(BeNil())

	err = r.handleSubscriptionDeletion(ctx, subscription, r.namedLogger())
	g.Expect(err).Should(BeNil())

	// the subscription has no kyma eventing finalizers and hence should not be processed in the function
	mockedBackend.AssertNotCalled(t, "DeleteSubscription", subscription)
	g.Expect(subscription.ObjectMeta.Finalizers).Should(BeEmpty())
	g.Expect(err).Should(BeNil())

	// add the eventing finalizer to test the function's branches
	subscription.ObjectMeta.Finalizers = []string{"eventing.kyma-project.io"}
	mockedBackend.On("DeleteSubscription", subscription).Return(nil)
	err = r.handleSubscriptionDeletion(ctx, subscription, r.namedLogger())
	g.Expect(err).Should(BeNil())

	// the function should have cleared the finalizers
	mockedBackend.AssertCalled(t, "DeleteSubscription", subscription)
	g.Expect(subscription.ObjectMeta.Finalizers).Should(BeEmpty())

	// check the changes were made on the kubernetes server
	var fetchedSub eventingv1alpha1.Subscription
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      "test",
		Namespace: namespaceName,
	}, &fetchedSub)
	g.Expect(err).Should(BeNil())
	g.Expect(fetchedSub.ObjectMeta.Finalizers).Should(BeEmpty())
}

func TestSyncSubscriptionStatus(t *testing.T) {
	g := NewGomegaWithT(t)
	testEnvironment := setupTestEnvironment(t)
	ctx, r := testEnvironment.Context, testEnvironment.Reconciler

	falseNatsSubActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
		corev1.ConditionFalse, "")
	trueNatsSubActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
		corev1.ConditionTrue, "")

	testCases := []struct {
		name               string
		givenSub           *eventingv1alpha1.Subscription
		givenNatsSubReady  bool
		givenForceStatus   bool
		expectedConditions []eventingv1alpha1.Condition
		expectedStatus     bool
	}{
		{
			name:               "Test subscription with no conditions should stay not ready with false condition",
			givenSub:           getSubscriptionWithConditionsAndStatus([]eventingv1alpha1.Condition{}, true, ""),
			givenNatsSubReady:  false,
			givenForceStatus:   false,
			expectedConditions: []eventingv1alpha1.Condition{falseNatsSubActiveCondition},
			expectedStatus:     false,
		},
		{
			name:               "Test subscription with false ready condition should stay not ready with false condition",
			givenSub:           getSubscriptionWithConditionsAndStatus([]eventingv1alpha1.Condition{falseNatsSubActiveCondition}, false, ""),
			givenNatsSubReady:  false,
			givenForceStatus:   false,
			expectedConditions: []eventingv1alpha1.Condition{falseNatsSubActiveCondition},
			expectedStatus:     false,
		},
		{
			name:               "Test subscription should become ready because of isNatsSubReady flag",
			givenSub:           getSubscriptionWithConditionsAndStatus([]eventingv1alpha1.Condition{falseNatsSubActiveCondition}, false, ""),
			givenNatsSubReady:  true,
			givenForceStatus:   false,
			expectedConditions: []eventingv1alpha1.Condition{trueNatsSubActiveCondition},
			expectedStatus:     true,
		},
		{
			name:               "Test subscription should stay with ready condition and status",
			givenSub:           getSubscriptionWithConditionsAndStatus([]eventingv1alpha1.Condition{trueNatsSubActiveCondition}, true, ""),
			givenNatsSubReady:  true,
			givenForceStatus:   false,
			expectedConditions: []eventingv1alpha1.Condition{trueNatsSubActiveCondition},
			expectedStatus:     true,
		},
		{
			name:               "Test subscription should become not ready because of false isNatsSubReady flag",
			givenSub:           getSubscriptionWithConditionsAndStatus([]eventingv1alpha1.Condition{trueNatsSubActiveCondition}, true, ""),
			givenNatsSubReady:  false,
			givenForceStatus:   false,
			expectedConditions: []eventingv1alpha1.Condition{falseNatsSubActiveCondition},
			expectedStatus:     false,
		},
	}
	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			sub := testCase.givenSub
			err := r.Client.Create(ctx, sub)
			g.Expect(err).Should(BeNil())

			err = r.syncSubscriptionStatus(ctx, sub, testCase.givenNatsSubReady, testCase.givenForceStatus, "")
			g.Expect(err).To(BeNil())

			g.Expect(len(testCase.expectedConditions)).Should(Equal(len(sub.Status.Conditions)))
			for i := range testCase.expectedConditions {
				comparisonResult := equalsConditionsIgnoringTime(testCase.expectedConditions[i], sub.Status.Conditions[i])
				g.Expect(comparisonResult).To(BeTrue())
			}
			g.Expect(testCase.expectedStatus).Should(Equal(sub.Status.Ready))

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
		svcNameToCreate       string
		expectedErrString     string
	}{
		{
			name:                  "test with invalid scheme",
			givenSubscriptionSink: "invalid Sink",
			expectedErrString:     "sink URL scheme should be 'http' or 'https'",
		},
		{
			name:                  "test with invalid URL",
			givenSubscriptionSink: "http://invalid Sink",
			expectedErrString:     "not able to parse sink url with error",
		},
		{
			name:                  "test with invalid suffix",
			givenSubscriptionSink: "https://svc2.test.local",
			expectedErrString:     "sink does not contain suffix",
		},
		{
			name:                  "test with invalid suffix and port",
			givenSubscriptionSink: "https://svc2.test.local:8080",
			expectedErrString:     "sink does not contain suffix",
		},
		{
			name:                  "test with invalid number of subdomains",
			givenSubscriptionSink: "https://svc.cluster.local:8080", // right suffix but 4 subdomains
			expectedErrString:     "sink should contain 5 sub-domains:",
		},
		{
			name:                  "test with different namespaces in subscription and sink name",
			givenSubscriptionSink: "https://eventing-nats.kyma-system.svc.cluster.local:8080", // sub is in test ns
			expectedErrString:     "namespace of subscription: test and the namespace of subscriber: kyma-system are different",
		},
		{
			name:                  "test with no existing svc in the cluster",
			givenSubscriptionSink: "https://eventing-nats.test.svc.cluster.local:8080",
			expectedErrString:     "sink is not valid cluster local svc, failed with error",
		},
		{
			name:                  "test with no existing svc in the cluster, service has the wrong name",
			givenSubscriptionSink: "https://eventing-nats.test.svc.cluster.local:8080",
			svcNameToCreate:       "test", // wrong name
			expectedErrString:     "sink is not valid cluster local svc, failed with error",
		},
		{
			name:                  "test with no existing svc in the cluster, service has the wrong name",
			givenSubscriptionSink: "https://eventing-nats.test.svc.cluster.local:8080",
			svcNameToCreate:       "eventing-nats",
			expectedErrString:     "",
		},
	}
	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			sub := getSubscriptionWithConditionsAndStatus([]eventingv1alpha1.Condition{}, true, testCase.givenSubscriptionSink)

			// create the service if required for test
			if testCase.svcNameToCreate != "" {
				svc := &corev1.Service{
					ObjectMeta: v1.ObjectMeta{
						Name:      testCase.svcNameToCreate,
						Namespace: namespaceName,
					},
				}

				err := r.Client.Create(ctx, svc)
				g.Expect(err).To(BeNil())
			}

			// call the defaultSinkValidator function
			err := r.sinkValidator(ctx, r, sub)
			log.WithContext().Infof("Result of defaultSinkValidatorCall: %+v", err)

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

// mocking the required resources for tests
type TestEnvironment struct {
	Context    context.Context
	Client     *client.WithWatch
	Backend    *mocks.MessagingBackend
	Reconciler *Reconciler
	Logger     *logger.Logger
}

func setupTestEnvironment(t *testing.T) *TestEnvironment {
	g := NewGomegaWithT(t)
	mockedBackend := &mocks.MessagingBackend{}
	ctx := context.Background()
	fakeClient := createFakeClient(g)

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	r := Reconciler{
		Backend:       mockedBackend,
		Client:        fakeClient,
		logger:        defaultLogger,
		recorder:      &record.FakeRecorder{},
		sinkValidator: defaultSinkValidator,
	}

	return &TestEnvironment{
		Context:    ctx,
		Client:     &fakeClient,
		Backend:    mockedBackend,
		Reconciler: &r,
		Logger:     defaultLogger,
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

func getSubscriptionWithConditionsAndStatus(conditions []eventingv1alpha1.Condition, status bool, sink string) *eventingv1alpha1.Subscription {
	return &eventingv1alpha1.Subscription{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: namespaceName,
		},
		Spec: eventingv1alpha1.SubscriptionSpec{
			Sink: sink,
		},
		Status: eventingv1alpha1.SubscriptionStatus{
			Conditions: conditions,
			Ready:      status,
		},
	}
}
