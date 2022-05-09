package jetstream

import (
	"context"
	"testing"
	"time"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/mocks"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/sink"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	subscriptionName = "testSubscription"
	namespaceName    = "test"
)

var defaultSubsConfig = env.DefaultSubscriptionConfig{MaxInFlightMessages: 1, DispatcherRetryPeriod: time.Second, DispatcherMaxRetries: 1}

func Test_handleSubscriptionDeletion(t *testing.T) {
	testEnvironment := setupTestEnvironment(t)
	ctx, r, mockedBackend := testEnvironment.Context, testEnvironment.Reconciler, testEnvironment.Backend

	testCases := []struct {
		name            string
		givenFinalizers []string
		wantDeleteCall  bool
		wantFinalizers  []string
	}{
		{
			name:            "With no finalizers the NATS subscription should not be deleted",
			givenFinalizers: []string{},
			wantDeleteCall:  false,
			wantFinalizers:  []string{},
		},
		{
			name:            "With eventing finalizer the NATS subscription should be deleted and the finalizer should be cleared",
			givenFinalizers: []string{Finalizer},
			wantDeleteCall:  true,
			wantFinalizers:  []string{},
		},
		{
			name:            "With wrong finalizer the NATS subscription should not be deleted",
			givenFinalizers: []string{"eventing2.kyma-project.io"},
			wantDeleteCall:  false,
			wantFinalizers:  []string{"eventing2.kyma-project.io"},
		},
	}

	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			// given
			subscription := controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithFinalizers(testCase.givenFinalizers),
			)
			err := r.Client.Create(testEnvironment.Context, subscription)
			require.NoError(t, err)

			mockedBackend.On("DeleteSubscription", subscription).Return(nil)

			// when
			err = r.handleSubscriptionDeletion(ctx, subscription, r.namedLogger())
			require.NoError(t, err)

			// then
			if testCase.wantDeleteCall {
				mockedBackend.AssertCalled(t, "DeleteSubscription", subscription)
			} else {
				mockedBackend.AssertNotCalled(t, "DeleteSubscription", subscription)
			}

			ensureFinalizerMatch(t, subscription, testCase.wantFinalizers)

			// check the changes were made on the kubernetes server
			fetchedSub, err := fetchTestSubscription(ctx, r)
			require.NoError(t, err)
			ensureFinalizerMatch(t, &fetchedSub, testCase.wantFinalizers)

			// clean up
			err = r.Client.Delete(ctx, subscription)
			require.NoError(t, err)
		})
	}
}

func Test_syncSubscriptionStatus(t *testing.T) {
	testEnvironment := setupTestEnvironment(t)
	ctx, r := testEnvironment.Context, testEnvironment.Reconciler

	jetStreamError := errors.New("Jetstream is not ready")
	falseNatsSubActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionNotActive,
		corev1.ConditionFalse, jetStreamError.Error())
	trueNatsSubActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
		corev1.ConditionTrue, "")

	testCases := []struct {
		name              string
		givenSub          *eventingv1alpha1.Subscription
		givenNatsSubReady bool
		givenUpdateStatus bool
		givenError        error
		wantConditions    []eventingv1alpha1.Condition
		wantStatus        bool
	}{
		{
			name: "Subscription with no conditions should stay not ready with false condition",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha1.Condition{}),
				controllertesting.WithStatus(true),
			),
			givenNatsSubReady: false,
			givenUpdateStatus: false,
			givenError:        jetStreamError,
			wantConditions:    []eventingv1alpha1.Condition{falseNatsSubActiveCondition},
			wantStatus:        false,
		},
		{
			name: "Subscription with false ready condition should stay not ready with false condition",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha1.Condition{falseNatsSubActiveCondition}),
				controllertesting.WithStatus(false),
			),
			givenNatsSubReady: false,
			givenUpdateStatus: false,
			givenError:        jetStreamError,
			wantConditions:    []eventingv1alpha1.Condition{falseNatsSubActiveCondition},
			wantStatus:        false,
		},
		{
			name: "Subscription should become ready because of isNatsSubReady flag",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha1.Condition{falseNatsSubActiveCondition}),
				controllertesting.WithStatus(false),
			),
			givenNatsSubReady: true, // isNatsSubReady
			givenUpdateStatus: false,
			givenError:        nil,
			wantConditions:    []eventingv1alpha1.Condition{trueNatsSubActiveCondition},
			wantStatus:        true,
		},
		{
			name: "Subscription should stay with ready condition and status",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha1.Condition{trueNatsSubActiveCondition}),
				controllertesting.WithStatus(true),
			),
			givenNatsSubReady: true,
			givenUpdateStatus: false,
			givenError:        nil,
			wantConditions:    []eventingv1alpha1.Condition{trueNatsSubActiveCondition},
			wantStatus:        true,
		},
		{
			name: "Subscription should become not ready because of false isNatsSubReady flag",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha1.Condition{trueNatsSubActiveCondition}),
				controllertesting.WithStatus(true),
			),
			givenNatsSubReady: false, // isNatsSubReady
			givenUpdateStatus: false,
			givenError:        jetStreamError,
			wantConditions:    []eventingv1alpha1.Condition{falseNatsSubActiveCondition},
			wantStatus:        false,
		},
		{
			name: "Subscription should stay with the same condition, but still updated because of the forceUpdateStatus flag",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha1.Condition{trueNatsSubActiveCondition}),
				controllertesting.WithStatus(true),
			),
			givenNatsSubReady: true,
			givenUpdateStatus: true,
			givenError:        nil,
			wantConditions:    []eventingv1alpha1.Condition{trueNatsSubActiveCondition},
			wantStatus:        true,
		},
	}
	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			// given
			sub := testCase.givenSub
			err := r.Client.Create(ctx, sub)
			require.NoError(t, err)

			// when
			err = r.syncSubscriptionStatus(ctx, sub, sub, testCase.givenUpdateStatus, testCase.givenError)
			require.NoError(t, err)

			// then
			ensureSubscriptionMatchesConditionsAndStatus(t, *sub, testCase.wantConditions, testCase.wantStatus)

			// fetch the sub also from k8s server in order to check whether changes were done both in-memory and on k8s server
			fetchedSub, err := fetchTestSubscription(ctx, r)
			require.NoError(t, err)
			ensureSubscriptionMatchesConditionsAndStatus(t, fetchedSub, testCase.wantConditions, testCase.wantStatus)

			// clean up
			err = r.Client.Delete(ctx, sub)
			require.NoError(t, err)
		})
	}
}

func Test_syncInitialStatus(t *testing.T) {
	testEnvironment := setupTestEnvironment(t)
	r := testEnvironment.Reconciler

	cleanSubjects := []string{controllertesting.OrderCreatedEventType}
	newSubjects := []string{controllertesting.OrderCreatedEventTypeNotClean}
	testEnvironment.Backend.On("GetJetStreamSubjects", cleanSubjects).Return(cleanSubjects)
	testEnvironment.Backend.On("GetJetStreamSubjects", newSubjects).Return(newSubjects)

	wantSubConfig := eventingv1alpha1.MergeSubsConfigs(nil, &defaultSubsConfig)
	newSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	updatedSubConfig := eventingv1alpha1.MergeSubsConfigs(nil, &newSubsConfig)

	testCases := []struct {
		name          string
		givenSub      *eventingv1alpha1.Subscription
		wantSubStatus eventingv1alpha1.SubscriptionStatus
		wantStatus    bool
	}{
		{
			name: "A new Subscription must be updated with cleanEventTypes and config and return true",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithStatus(false),
				controllertesting.WithFilter("", controllertesting.OrderCreatedEventType),
			),
			wantSubStatus: eventingv1alpha1.SubscriptionStatus{
				Ready:           false,
				CleanEventTypes: cleanSubjects,
				Config:          wantSubConfig,
			},
			wantStatus: true,
		},
		{
			name: "A subscription with the same cleanEventTypes and config must return false",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithStatus(true),
				controllertesting.WithFilter("", controllertesting.OrderCreatedEventType),
				controllertesting.WithStatusCleanEventTypes([]string{controllertesting.OrderCreatedEventType}),
				controllertesting.WithStatusConfig(defaultSubsConfig),
			),
			wantSubStatus: eventingv1alpha1.SubscriptionStatus{
				Ready:           true,
				CleanEventTypes: cleanSubjects,
				Config:          wantSubConfig,
			},
			wantStatus: false,
		},
		{
			name: "A subscription with the same cleanEventTypes and new config must return true",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithStatus(true),
				controllertesting.WithFilter("", controllertesting.OrderCreatedEventType),
				controllertesting.WithStatusCleanEventTypes([]string{controllertesting.OrderCreatedEventType}),
				controllertesting.WithSpecConfig(newSubsConfig),
			),
			wantSubStatus: eventingv1alpha1.SubscriptionStatus{
				Ready:           true,
				CleanEventTypes: cleanSubjects,
				Config:          updatedSubConfig,
			},
			wantStatus: true,
		},
		{
			name: "A subscription with changed cleanEventTypes and the same config must return true",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithStatus(true),
				controllertesting.WithFilter("", controllertesting.OrderCreatedEventTypeNotClean),
				controllertesting.WithStatusCleanEventTypes(cleanSubjects),
				controllertesting.WithStatusConfig(defaultSubsConfig),
			),
			wantSubStatus: eventingv1alpha1.SubscriptionStatus{
				Ready:           true,
				CleanEventTypes: newSubjects,
				Config:          wantSubConfig,
			},
			wantStatus: true,
		},
	}
	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			// given
			sub := testCase.givenSub

			// when
			gotStatus, err := r.syncInitialStatus(sub, r.namedLogger())
			require.NoError(t, err)

			// then
			require.Equal(t, testCase.wantSubStatus.CleanEventTypes, sub.Status.CleanEventTypes)
			require.Equal(t, testCase.wantSubStatus.Config, sub.Status.Config)
			require.Equal(t, testCase.wantSubStatus.Ready, sub.Status.Ready)
			require.Equal(t, gotStatus, testCase.wantStatus)
		})
	}
}

// helper functions and structs

// TestEnvironment provides mocked resources for tests.
type TestEnvironment struct {
	Context    context.Context
	Client     *client.WithWatch
	Backend    *mocks.JetStreamBackend
	Reconciler *Reconciler
	Logger     *logger.Logger
	Recorder   *record.FakeRecorder
}

// setupTestEnvironment is a TestEnvironment constructor
func setupTestEnvironment(t *testing.T) *TestEnvironment {
	mockedBackend := &mocks.JetStreamBackend{}
	ctx := context.Background()
	fakeClient := createFakeClient(t)
	recorder := &record.FakeRecorder{}

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}
	cleaner := func(et string) (string, error) {
		return et, nil
	}
	defaultSinkValidator := sink.NewValidator(ctx, fakeClient, recorder, defaultLogger)

	r := Reconciler{
		Backend:          mockedBackend,
		Client:           fakeClient,
		logger:           defaultLogger,
		subsConfig:       defaultSubsConfig,
		recorder:         recorder,
		sinkValidator:    defaultSinkValidator,
		eventTypeCleaner: eventtype.CleanerFunc(cleaner),
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

func createFakeClient(t *testing.T) client.WithWatch {
	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)
	return fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
}

func ensureSubscriptionMatchesConditionsAndStatus(t *testing.T, subscription eventingv1alpha1.Subscription, wantConditions []eventingv1alpha1.Condition, wantStatus bool) {
	require.Equal(t, len(wantConditions), len(subscription.Status.Conditions))
	comparisonResult := eventingv1alpha1.ConditionsEquals(wantConditions, subscription.Status.Conditions)
	require.True(t, comparisonResult)
	require.Equal(t, wantStatus, subscription.Status.Ready)
}

func fetchTestSubscription(ctx context.Context, r *Reconciler) (eventingv1alpha1.Subscription, error) {
	var fetchedSub eventingv1alpha1.Subscription
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      subscriptionName,
		Namespace: namespaceName,
	}, &fetchedSub)
	return fetchedSub, err
}

func ensureFinalizerMatch(t *testing.T, subscription *eventingv1alpha1.Subscription, wantFinalizers []string) {
	if len(wantFinalizers) == 0 {
		require.Empty(t, subscription.ObjectMeta.Finalizers)
	} else {
		require.Equal(t, wantFinalizers, subscription.ObjectMeta.Finalizers)
	}
}
