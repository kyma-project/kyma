//go:build integration
// +build integration

// This file contains unit tests for the NATS subscription reconciler.
// It uses the testing.T and stretchr/testify libraries to perform assertions.
// testEnvironment struct mocks the required resources.
package nats

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/xerrors"

	"github.com/stretchr/testify/mock"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink"

	"github.com/stretchr/testify/require"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/mocks"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

const (
	subscriptionName = "testSubscription"
	namespaceName    = "test"
)

var defaultSubsConfig = env.DefaultSubscriptionConfig{MaxInFlightMessages: 1, DispatcherRetryPeriod: time.Second, DispatcherMaxRetries: 1}

func Test_handleSubscriptionDeletion(t *testing.T) {
	testEnvironment := setupTestEnvironment(t)
	ctx, r, mockedBackend := context.Background(), testEnvironment.reconciler, testEnvironment.backend

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
			givenFinalizers: []string{eventingv1alpha1.Finalizer},
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
			sub := NewTestSubscription(
				controllertesting.WithFinalizers(testCase.givenFinalizers),
			)
			err := r.Client.Create(context.Background(), sub)
			require.NoError(t, err)

			mockedBackend.On("DeleteSubscription", sub).Return(nil)

			// when
			err = r.handleSubscriptionDeletion(ctx, sub, r.namedLogger())
			require.NoError(t, err)

			// then
			if testCase.wantDeleteCall {
				mockedBackend.AssertCalled(t, "DeleteSubscription", sub)
			} else {
				mockedBackend.AssertNotCalled(t, "DeleteSubscription", sub)
			}

			ensureFinalizerMatch(t, sub, testCase.wantFinalizers)

			// check the changes were made on the kubernetes server
			fetchedSub, err := fetchTestSubscription(ctx, r)
			require.NoError(t, err)
			ensureFinalizerMatch(t, &fetchedSub, testCase.wantFinalizers)

			// clean up
			err = r.Client.Delete(ctx, sub)
			require.NoError(t, err)
		})
	}
}

func Test_syncSubscriptionStatus(t *testing.T) {
	testEnvironment := setupTestEnvironment(t)
	ctx, r := context.Background(), testEnvironment.reconciler

	message := "message is not required for tests"
	falseNatsSubActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionNotActive,
		corev1.ConditionFalse, message)
	trueNatsSubActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
		corev1.ConditionTrue, message)

	testCases := []struct {
		name              string
		givenSub          *eventingv1alpha1.Subscription
		givenNatsSubReady bool
		givenForceStatus  bool
		wantConditions    []eventingv1alpha1.Condition
		wantStatus        bool
	}{
		{
			name: "Subscription with no conditions should stay not ready with false condition",
			givenSub: NewTestSubscription(
				controllertesting.WithConditions([]eventingv1alpha1.Condition{}),
				controllertesting.WithStatus(true),
			),
			givenNatsSubReady: false,
			givenForceStatus:  false,
			wantConditions:    []eventingv1alpha1.Condition{falseNatsSubActiveCondition},
			wantStatus:        false,
		},
		{
			name: "Subscription with false ready condition should stay not ready with false condition",
			givenSub: NewTestSubscription(
				controllertesting.WithConditions([]eventingv1alpha1.Condition{falseNatsSubActiveCondition}),
				controllertesting.WithStatus(false),
			),
			givenNatsSubReady: false,
			givenForceStatus:  false,
			wantConditions:    []eventingv1alpha1.Condition{falseNatsSubActiveCondition},
			wantStatus:        false,
		},
		{
			name: "Subscription should become ready because of isNatsSubReady flag",
			givenSub: NewTestSubscription(
				controllertesting.WithConditions([]eventingv1alpha1.Condition{falseNatsSubActiveCondition}),
				controllertesting.WithStatus(false),
			),
			givenNatsSubReady: true, // isNatsSubReady
			givenForceStatus:  false,
			wantConditions:    []eventingv1alpha1.Condition{trueNatsSubActiveCondition},
			wantStatus:        true,
		},
		{
			name: "Subscription should stay with ready condition and status",
			givenSub: NewTestSubscription(
				controllertesting.WithConditions([]eventingv1alpha1.Condition{trueNatsSubActiveCondition}),
				controllertesting.WithStatus(true),
			),
			givenNatsSubReady: true,
			givenForceStatus:  false,
			wantConditions:    []eventingv1alpha1.Condition{trueNatsSubActiveCondition},
			wantStatus:        true,
		},
		{
			name: "Subscription should become not ready because of false isNatsSubReady flag",
			givenSub: NewTestSubscription(
				controllertesting.WithConditions([]eventingv1alpha1.Condition{trueNatsSubActiveCondition}),
				controllertesting.WithStatus(true),
			),
			givenNatsSubReady: false, // isNatsSubReady
			givenForceStatus:  false,
			wantConditions:    []eventingv1alpha1.Condition{falseNatsSubActiveCondition},
			wantStatus:        false,
		},
		{
			name: "Subscription should stay with the same condition, but still updated because of the forceUpdateStatus flag",
			givenSub: NewTestSubscription(
				controllertesting.WithConditions([]eventingv1alpha1.Condition{trueNatsSubActiveCondition}),
				controllertesting.WithStatus(true),
			),
			givenNatsSubReady: true,
			givenForceStatus:  true,
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
			err = r.syncSubscriptionStatus(ctx, sub, testCase.givenNatsSubReady, testCase.givenForceStatus, message)
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
	r := testEnvironment.reconciler

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
			givenSub: NewTestSubscription(
				controllertesting.WithStatus(false),
				controllertesting.WithFilter("", controllertesting.OrderCreatedEventType),
			),
			wantSubStatus: eventingv1alpha1.SubscriptionStatus{
				Ready:           false,
				CleanEventTypes: []string{controllertesting.OrderCreatedEventType},
				Config:          wantSubConfig,
			},
			wantStatus: true,
		},
		{
			name: "A subscription with the same cleanEventTypes and config must return false",
			givenSub: NewTestSubscription(
				controllertesting.WithStatus(true),
				controllertesting.WithFilter("", controllertesting.OrderCreatedEventType),
				controllertesting.WithStatusCleanEventTypes([]string{controllertesting.OrderCreatedEventType}),
				controllertesting.WithStatusConfig(defaultSubsConfig),
			),
			wantSubStatus: eventingv1alpha1.SubscriptionStatus{
				Ready:           true,
				CleanEventTypes: []string{controllertesting.OrderCreatedEventType},
				Config:          wantSubConfig,
			},
			wantStatus: false,
		},
		{
			name: "A subscription with the same cleanEventTypes and new config must return true",
			givenSub: NewTestSubscription(
				controllertesting.WithStatus(true),
				controllertesting.WithFilter("", controllertesting.OrderCreatedEventType),
				controllertesting.WithStatusCleanEventTypes([]string{controllertesting.OrderCreatedEventType}),
				controllertesting.WithSpecConfig(newSubsConfig),
			),
			wantSubStatus: eventingv1alpha1.SubscriptionStatus{
				Ready:           true,
				CleanEventTypes: []string{controllertesting.OrderCreatedEventType},
				Config:          updatedSubConfig,
			},
			wantStatus: true,
		},
		{
			name: "A subscription with changed cleanEventTypes and the same config must return true",
			givenSub: NewTestSubscription(
				controllertesting.WithStatus(true),
				controllertesting.WithFilter("", controllertesting.OrderCreatedEventTypeNotClean),
				controllertesting.WithStatusCleanEventTypes([]string{controllertesting.OrderCreatedEventType}),
				controllertesting.WithStatusConfig(defaultSubsConfig),
			),
			wantSubStatus: eventingv1alpha1.SubscriptionStatus{
				Ready:           true,
				CleanEventTypes: []string{controllertesting.OrderCreatedEventTypeNotClean},
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
			gotStatus, err := r.syncInitialStatus(sub)
			require.NoError(t, err)

			// then
			require.Equal(t, testCase.wantSubStatus.CleanEventTypes, sub.Status.CleanEventTypes)
			require.Equal(t, testCase.wantSubStatus.Config, sub.Status.Config)
			require.Equal(t, testCase.wantSubStatus.Ready, sub.Status.Ready)
			require.Equal(t, gotStatus, testCase.wantStatus)
		})
	}
}

// Test the return values of the Reconcile() method of the reconciler. This is important, as it dictates whether the
// reconciliation should be requeued by Controller Runtime, and if so with how much initial delay.
// Returning error or a `Result{Requeue: true}` would cause the reconciliation to be requeued.
// Everything else is mocked since we are only interested in the logic of the Reconcile method and not the reconciler dependencies.
func TestReconciler_Reconcile(t *testing.T) {
	ctx := context.Background()
	req := require.New(t)

	defaultSubConfig := env.DefaultSubscriptionConfig{}
	// A subscription with the correct Finalizer, ready for reconciliation with the backend.
	testSub := controllertesting.NewSubscription("sub1", "test",
		controllertesting.WithFinalizers([]string{eventingv1alpha1.Finalizer}),
		controllertesting.WithFilter(controllertesting.EventSource, controllertesting.OrderCreatedEventType),
	)
	// A subscription marked for deletion.
	testSubUnderDeletion := controllertesting.NewSubscription("sub2", "test",
		controllertesting.WithNonZeroDeletionTimestamp(),
		controllertesting.WithFinalizers([]string{eventingv1alpha1.Finalizer}),
		controllertesting.WithFilter(controllertesting.EventSource, controllertesting.OrderCreatedEventType),
	)

	backendSyncErr := errors.New("backend sync error")
	backendDeleteErr := errors.New("backend delete error")
	validatorErr := errors.New("invalid sink")
	cleanerErr := errors.New("invalid event type format")
	happyCleaner := eventtype.CleanerFunc(func(et string) (string, error) { return et, nil })
	unhappyCleaner := eventtype.CleanerFunc(func(et string) (string, error) { return et, cleanerErr })
	happyValidator := sink.ValidatorFunc(func(s *eventingv1alpha1.Subscription) error { return nil })
	unhappyValidator := sink.ValidatorFunc(func(s *eventingv1alpha1.Subscription) error { return validatorErr })

	var testCases = []struct {
		name                 string
		givenSubscription    *eventingv1alpha1.Subscription
		givenReconcilerSetup func() *Reconciler
		wantReconcileResult  ctrl.Result
		wantReconcileError   error
	}{
		{
			name:              "Return nil and default Result{} when there is no error from the reconciler dependencies",
			givenSubscription: testSub,
			givenReconcilerSetup: func() *Reconciler {
				te := setupTestEnvironment(t)
				fakeClient := te.clientBuilder.WithObjects(testSub).Build()
				te.backend.On("Initialize", mock.Anything).Return(nil)
				te.backend.On("SyncSubscription", mock.Anything).Return(nil)
				return NewReconciler(ctx, fakeClient, te.backend, happyCleaner, te.logger, te.recorder, defaultSubConfig, happyValidator)
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  nil,
		},
		{
			name:              "Return error and default Result{} when backend sync returns error",
			givenSubscription: testSub,
			givenReconcilerSetup: func() *Reconciler {
				te := setupTestEnvironment(t)
				fakeClient := te.clientBuilder.WithObjects(testSub).Build()
				te.backend.On("Initialize", mock.Anything).Return(nil)
				te.backend.On("SyncSubscription", mock.Anything).Return(backendSyncErr)
				return NewReconciler(ctx, fakeClient, te.backend, happyCleaner, te.logger, te.recorder, defaultSubConfig, happyValidator)
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  xerrors.Errorf("failed to sync subscription to NATS backend: %v", backendSyncErr),
		},
		{
			name:              "Return error and default Result{} when backend delete returns error",
			givenSubscription: testSubUnderDeletion,
			givenReconcilerSetup: func() *Reconciler {
				te := setupTestEnvironment(t)
				fakeClient := te.clientBuilder.WithObjects(testSubUnderDeletion).Build()
				te.backend.On("Initialize", mock.Anything).Return(nil)
				te.backend.On("DeleteSubscription", mock.Anything).Return(backendDeleteErr)
				return NewReconciler(ctx, fakeClient, te.backend, happyCleaner, te.logger, te.recorder, defaultSubConfig, happyValidator)
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  xerrors.Errorf("failed to delete NATS subscription: %v", backendDeleteErr),
		},
		{
			name:              "Return error and default Result{} when validator returns error",
			givenSubscription: testSub,
			givenReconcilerSetup: func() *Reconciler {
				te := setupTestEnvironment(t)
				fakeClient := te.clientBuilder.WithObjects(testSub).Build()
				te.backend.On("Initialize", mock.Anything).Return(nil)
				te.backend.On("SyncSubscription", mock.Anything).Return(nil)
				return NewReconciler(ctx, fakeClient, te.backend, happyCleaner, te.logger, te.recorder, defaultSubConfig, unhappyValidator)
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  xerrors.Errorf("failed to validate subscription sink URL: %v", validatorErr),
		},
		{
			name:              "Return error and default Result{} when event type cleaner returns error",
			givenSubscription: testSub,
			givenReconcilerSetup: func() *Reconciler {
				te := setupTestEnvironment(t)
				fakeClient := te.clientBuilder.WithObjects(testSub).Build()
				te.backend.On("Initialize", mock.Anything).Return(nil)
				te.backend.On("SyncSubscription", mock.Anything).Return(nil)
				return NewReconciler(ctx, fakeClient, te.backend, unhappyCleaner, te.logger, te.recorder, defaultSubConfig, happyValidator)
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  xerrors.Errorf("failed to sync subscription initial status: %v", xerrors.Errorf("failed to get clean subjects: %v", cleanerErr)),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			reconciler := testCase.givenReconcilerSetup()
			r := ctrl.Request{NamespacedName: types.NamespacedName{
				Namespace: tc.givenSubscription.Namespace,
				Name:      tc.givenSubscription.Name,
			}}
			res, err := reconciler.Reconcile(context.Background(), r)
			req.Equal(res, tc.wantReconcileResult)
			if tc.wantReconcileError == nil {
				req.Equal(err, nil)
			} else {
				req.Equal(err.Error(), tc.wantReconcileError.Error())
			}
		})
	}
}

// helper functions and structs

// testEnvironment provides mocked resources for tests.
type testEnvironment struct {
	clientBuilder *fake.ClientBuilder
	backend       *mocks.NatsBackend
	reconciler    *Reconciler
	logger        *logger.Logger
	recorder      *record.FakeRecorder
}

// setupTestEnvironment is a testEnvironment constructor.
func setupTestEnvironment(t *testing.T) *testEnvironment {
	mockedBackend := &mocks.NatsBackend{}
	ctx := context.Background()
	fakeClientBuilder := createFakeClientBuilder(t)
	fakeClient := fakeClientBuilder.Build()
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

	return &testEnvironment{
		clientBuilder: fakeClientBuilder,
		backend:       mockedBackend,
		reconciler:    &r,
		logger:        defaultLogger,
		recorder:      recorder,
	}
}

func createFakeClientBuilder(t *testing.T) *fake.ClientBuilder {
	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)
	return fake.NewClientBuilder().WithScheme(scheme.Scheme)
}

// NewTestSubscription creates a test subscription.
func NewTestSubscription(opts ...controllertesting.SubscriptionOpt) *eventingv1alpha1.Subscription {
	return controllertesting.NewSubscription(subscriptionName, namespaceName, opts...)
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

func ensureSubscriptionMatchesConditionsAndStatus(t *testing.T, subscription eventingv1alpha1.Subscription, wantConditions []eventingv1alpha1.Condition, wantStatus bool) {
	require.Equal(t, len(wantConditions), len(subscription.Status.Conditions))
	for i := range wantConditions {
		comparisonResult := equalsConditionsIgnoringTime(wantConditions[i], subscription.Status.Conditions[i])
		require.True(t, comparisonResult)
	}
	require.Equal(t, wantStatus, subscription.Status.Ready)
}
