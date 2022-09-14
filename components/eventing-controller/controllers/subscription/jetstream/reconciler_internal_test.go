package jetstream

import (
	"context"
	"testing"
	"time"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/mocks"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

const (
	subscriptionName = "testSubscription"
	namespaceName    = "test"
)

var defaultSubsConfig = env.DefaultSubscriptionConfig{MaxInFlightMessages: 1, DispatcherRetryPeriod: time.Second, DispatcherMaxRetries: 1}

// TestReconciler_Reconcile tests the return values of the Reconcile() method of the reconciler.
// This is important, as it dictates whether the reconciliation should be requeued by Controller Runtime,
// and if so with how much initial delay. Returning error or a `Result{Requeue: true}` would cause the reconciliation to be requeued.
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
				te := setupTestEnvironment(t, testSub)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				te.Backend.On("SyncSubscription", mock.Anything).Return(nil)
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, happyCleaner, defaultSubConfig, happyValidator)
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  nil,
		},
		{
			name:              "Return nil and default Result{} when the subscription does not exist on the cluster",
			givenSubscription: testSub,
			givenReconcilerSetup: func() *Reconciler {
				te := setupTestEnvironment(t)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, happyCleaner, defaultSubConfig, happyValidator)
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  nil,
		},
		{
			name:              "Return error and default Result{} when backend sync returns error",
			givenSubscription: testSub,
			givenReconcilerSetup: func() *Reconciler {
				te := setupTestEnvironment(t, testSub)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				te.Backend.On("SyncSubscription", mock.Anything).Return(backendSyncErr)
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, happyCleaner, defaultSubConfig, happyValidator)
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  backendSyncErr,
		},
		{
			name:              "Return error and default Result{} when backend delete returns error",
			givenSubscription: testSubUnderDeletion,
			givenReconcilerSetup: func() *Reconciler {
				te := setupTestEnvironment(t, testSubUnderDeletion)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				te.Backend.On("DeleteSubscription", mock.Anything).Return(backendDeleteErr)
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, happyCleaner, defaultSubConfig, happyValidator)
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  xerrors.Errorf("failed to delete JetStream subscription: %v", backendDeleteErr),
		},
		{
			name:              "Return error and default Result{} when validator returns error",
			givenSubscription: testSub,
			givenReconcilerSetup: func() *Reconciler {
				te := setupTestEnvironment(t, testSub)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				te.Backend.On("SyncSubscription", mock.Anything).Return(nil)
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, happyCleaner, defaultSubConfig, unhappyValidator)
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  validatorErr,
		},
		{
			name:              "Return error and default Result{} when event type cleaner returns error",
			givenSubscription: testSub,
			givenReconcilerSetup: func() *Reconciler {
				te := setupTestEnvironment(t, testSub)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				te.Backend.On("SyncSubscription", mock.Anything).Return(nil)
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, unhappyCleaner, defaultSubConfig, happyValidator)
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  xerrors.Errorf("failed to get clean subjects: %v", cleanerErr),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			reconciler := tc.givenReconcilerSetup()
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
			sub := controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithFinalizers(testCase.givenFinalizers),
			)
			err := r.Client.Create(testEnvironment.Context, sub)
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
	ctx, r := testEnvironment.Context, testEnvironment.Reconciler

	jetStreamError := errors.New("JetStream is not ready")
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
			err = r.syncSubscriptionStatus(ctx, sub, testCase.givenUpdateStatus, testCase.givenError)
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

// helper functions and structs

// TestEnvironment provides mocked resources for tests.
type TestEnvironment struct {
	Context    context.Context
	Client     client.Client
	Backend    *mocks.JetStreamBackend
	Reconciler *Reconciler
	Logger     *logger.Logger
	Recorder   *record.FakeRecorder
}

// setupTestEnvironment is a TestEnvironment constructor.
func setupTestEnvironment(t *testing.T, objs ...client.Object) *TestEnvironment {
	mockedBackend := &mocks.JetStreamBackend{}
	ctx := context.Background()
	fakeClient := createFakeClientBuilder(t).WithObjects(objs...).Build()
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
		Client:     fakeClient,
		Backend:    mockedBackend,
		Reconciler: &r,
		Logger:     defaultLogger,
		Recorder:   recorder,
	}
}

func createFakeClientBuilder(t *testing.T) *fake.ClientBuilder {
	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)
	return fake.NewClientBuilder().WithScheme(scheme.Scheme)
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
