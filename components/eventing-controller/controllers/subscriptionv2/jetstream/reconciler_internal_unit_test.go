package jetstream

import (
	"context"
	"testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2/mocks"
	sinkv2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink/v2"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
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
)

const (
	subscriptionName = "testSubscription"
	namespaceName    = "test"
)

// Test_Reconcile tests the return values of the Reconcile() method of the reconciler.
// This is important, as it dictates whether the reconciliation should be requeued by Controller Runtime,
// and if so with how much initial delay. Returning error or a `Result{Requeue: true}` would cause the reconciliation to be requeued.
// Everything else is mocked since we are only interested in the logic of the Reconcile method and not the reconciler dependencies.
func Test_Reconcile(t *testing.T) {
	ctx := context.Background()
	req := require.New(t)

	// A subscription with the correct Finalizer, ready for reconciliation with the backend.
	testSub := controllertesting.NewSubscription("sub1", namespaceName,
		controllertesting.WithFinalizers([]string{eventingv1alpha2.Finalizer}),
		controllertesting.WithSource(controllertesting.EventSourceClean),
		controllertesting.WithEventType(controllertesting.OrderCreatedV1Event),
	)
	// A subscription marked for deletion.
	testSubUnderDeletion := controllertesting.NewSubscription("sub2", namespaceName,
		controllertesting.WithNonZeroDeletionTimestamp(),
		controllertesting.WithFinalizers([]string{eventingv1alpha2.Finalizer}),
		controllertesting.WithSource(controllertesting.EventSourceClean),
		controllertesting.WithEventType(controllertesting.OrderCreatedV1Event),
	)

	backendSyncErr := errors.New("backend sync error")
	missingSubSyncErr := errors.New(jetstreamv2.MissingNATSSubscriptionMsg)
	backendDeleteErr := errors.New("backend delete error")
	validatorErr := errors.New("invalid sink")
	happyValidator := sinkv2.ValidatorFunc(func(s *eventingv1alpha2.Subscription) error { return nil })
	unhappyValidator := sinkv2.ValidatorFunc(func(s *eventingv1alpha2.Subscription) error { return validatorErr })

	var testCases = []struct {
		name                 string
		givenSubscription    *eventingv1alpha2.Subscription
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
				te.Backend.On("GetJetStreamSubjects", mock.Anything, mock.Anything, mock.Anything).Return([]string{controllertesting.JetStreamSubject})
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, te.Cleaner, happyValidator)
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
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, te.Cleaner, happyValidator)
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
				te.Backend.On("GetJetStreamSubjects", mock.Anything, mock.Anything, mock.Anything).Return([]string{controllertesting.JetStreamSubject})
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, te.Cleaner, happyValidator)
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  backendSyncErr,
		},
		{
			name:              "Return nil and RequeueAfter with requeue duration when backend sync returns missingSubscriptionErr",
			givenSubscription: testSub,
			givenReconcilerSetup: func() *Reconciler {
				te := setupTestEnvironment(t, testSub)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				te.Backend.On("SyncSubscription", mock.Anything).Return(missingSubSyncErr)
				te.Backend.On("GetJetStreamSubjects", mock.Anything, mock.Anything, mock.Anything).Return([]string{controllertesting.JetStreamSubject})
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, te.Cleaner, happyValidator)
			},
			wantReconcileResult: ctrl.Result{RequeueAfter: requeueDuration},
			wantReconcileError:  nil,
		},
		{
			name:              "Return error and default Result{} when backend delete returns error",
			givenSubscription: testSubUnderDeletion,
			givenReconcilerSetup: func() *Reconciler {
				te := setupTestEnvironment(t, testSubUnderDeletion)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				te.Backend.On("DeleteSubscription", mock.Anything).Return(backendDeleteErr)
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, te.Cleaner, happyValidator)
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
				te.Backend.On("GetJetStreamSubjects", mock.Anything, mock.Anything, mock.Anything).Return([]string{controllertesting.JetStreamSubject})
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, te.Cleaner, unhappyValidator)
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  validatorErr,
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
			givenFinalizers: []string{eventingv1alpha2.Finalizer},
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
	falseNatsSubActiveCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscriptionActive,
		eventingv1alpha2.ConditionReasonNATSSubscriptionNotActive,
		corev1.ConditionFalse, jetStreamError.Error())
	trueNatsSubActiveCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscriptionActive,
		eventingv1alpha2.ConditionReasonNATSSubscriptionActive,
		corev1.ConditionTrue, "")

	testCases := []struct {
		name              string
		givenSub          *eventingv1alpha2.Subscription
		givenNatsSubReady bool
		givenUpdateStatus bool
		givenError        error
		wantConditions    []eventingv1alpha2.Condition
		wantStatus        bool
	}{
		{
			name: "Subscription with no conditions should stay not ready with false condition",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha2.Condition{}),
				controllertesting.WithStatus(true),
			),
			givenNatsSubReady: false,
			givenUpdateStatus: false,
			givenError:        jetStreamError,
			wantConditions:    []eventingv1alpha2.Condition{falseNatsSubActiveCondition},
			wantStatus:        false,
		},
		{
			name: "Subscription with false ready condition should stay not ready with false condition",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha2.Condition{falseNatsSubActiveCondition}),
				controllertesting.WithStatus(false),
			),
			givenNatsSubReady: false,
			givenUpdateStatus: false,
			givenError:        jetStreamError,
			wantConditions:    []eventingv1alpha2.Condition{falseNatsSubActiveCondition},
			wantStatus:        false,
		},
		{
			name: "Subscription should become ready because of isNatsSubReady flag",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha2.Condition{falseNatsSubActiveCondition}),
				controllertesting.WithStatus(false),
			),
			givenNatsSubReady: true, // isNatsSubReady
			givenUpdateStatus: false,
			givenError:        nil,
			wantConditions:    []eventingv1alpha2.Condition{trueNatsSubActiveCondition},
			wantStatus:        true,
		},
		{
			name: "Subscription should stay with ready condition and status",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha2.Condition{trueNatsSubActiveCondition}),
				controllertesting.WithStatus(true),
			),
			givenNatsSubReady: true,
			givenUpdateStatus: false,
			givenError:        nil,
			wantConditions:    []eventingv1alpha2.Condition{trueNatsSubActiveCondition},
			wantStatus:        true,
		},
		{
			name: "Subscription should become not ready because of false isNatsSubReady flag",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha2.Condition{trueNatsSubActiveCondition}),
				controllertesting.WithStatus(true),
			),
			givenNatsSubReady: false, // isNatsSubReady
			givenUpdateStatus: false,
			givenError:        jetStreamError,
			wantConditions:    []eventingv1alpha2.Condition{falseNatsSubActiveCondition},
			wantStatus:        false,
		},
		{
			name: "Subscription should stay with the same condition, but still updated because of the forceUpdateStatus flag",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha2.Condition{trueNatsSubActiveCondition}),
				controllertesting.WithStatus(true),
			),
			givenNatsSubReady: true,
			givenUpdateStatus: true,
			givenError:        nil,
			wantConditions:    []eventingv1alpha2.Condition{trueNatsSubActiveCondition},
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

	jsSubjects := []string{controllertesting.JetStreamSubjectV2}
	eventTypes := []eventingv1alpha2.EventType{
		{
			OriginalType: controllertesting.OrderCreatedUncleanEvent,
			CleanType:    controllertesting.OrderCreatedCleanEvent,
		},
	}
	jsTypes := []eventingv1alpha2.JetStreamTypes{
		{
			OriginalType: controllertesting.OrderCreatedUncleanEvent,
			ConsumerName: "a59e97ceb4883938c193bc0abf6e8bca",
		},
	}
	backendStatus := eventingv1alpha2.Backend{
		Types: jsTypes,
	}
	testEnvironment.Backend.On("GetJetStreamSubjects", mock.Anything, mock.Anything, mock.Anything).Return(jsSubjects)

	testCases := []struct {
		name          string
		givenSub      *eventingv1alpha2.Subscription
		wantSubStatus eventingv1alpha2.SubscriptionStatus
		wantStatus    bool
	}{
		{
			name: "A new Subscription must be updated with cleanEventTypes and backend jstypes and return true",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithStatus(false),
				controllertesting.WithEventType(controllertesting.OrderCreatedUncleanEvent),
			),
			wantSubStatus: eventingv1alpha2.SubscriptionStatus{
				Ready:   false,
				Types:   eventTypes,
				Backend: backendStatus,
			},
			wantStatus: true,
		},
		{
			name: "A subscription with the same cleanEventTypes and jsTypes must return false",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithStatus(true),
				controllertesting.WithEventType(controllertesting.OrderCreatedUncleanEvent),
				controllertesting.WithStatusTypes(eventTypes),
				controllertesting.WithStatusJSBackendTypes(jsTypes),
			),
			wantSubStatus: eventingv1alpha2.SubscriptionStatus{
				Ready:   true,
				Types:   eventTypes,
				Backend: backendStatus,
			},
			wantStatus: false,
		},
		{
			name: "A subscription with the same eventTypes and new jsTypes must return true",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithStatus(true),
				controllertesting.WithEventType(controllertesting.OrderCreatedUncleanEvent),
				controllertesting.WithStatusTypes(eventTypes),
			),
			wantSubStatus: eventingv1alpha2.SubscriptionStatus{
				Ready:   true,
				Types:   eventTypes,
				Backend: backendStatus,
			},
			wantStatus: true,
		},
		{
			name: "A subscription with changed eventTypes and the same jsTypes must return true",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithStatus(true),
				controllertesting.WithEventType(controllertesting.OrderCreatedUncleanEvent),
				controllertesting.WithStatusJSBackendTypes(jsTypes),
			),
			wantSubStatus: eventingv1alpha2.SubscriptionStatus{
				Ready:   true,
				Types:   eventTypes,
				Backend: backendStatus,
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
			require.Equal(t, testCase.wantSubStatus.Types, sub.Status.Types)
			require.Equal(t, testCase.wantSubStatus.Backend, sub.Status.Backend)
			require.Equal(t, testCase.wantSubStatus.Ready, sub.Status.Ready)
			require.Equal(t, testCase.wantStatus, gotStatus)
		})
	}
}

func Test_addFinalizerToSubscription(t *testing.T) {
	// given
	sub := controllertesting.NewSubscription(subscriptionName, namespaceName)
	fakeSub := controllertesting.NewSubscription("fake", namespaceName)
	testEnvironment := setupTestEnvironment(t, sub)
	r := testEnvironment.Reconciler
	fetchedSub, err := fetchTestSubscription(testEnvironment.Context, r)
	require.NoError(t, err)
	ensureFinalizerMatch(t, &fetchedSub, []string{})

	testCases := []struct {
		name      string
		givenSub  *eventingv1alpha2.Subscription
		wantError bool
	}{
		{
			name:      "A new Subscription must be updated with cleanEventTypes and backend jstypes and return true",
			givenSub:  sub,
			wantError: false,
		},
		{
			name:      "A new Subscription must be updated with cleanEventTypes and backend jstypes and return true",
			givenSub:  fakeSub,
			wantError: true,
		},
	}

	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			// when
			err = r.addFinalizerToSubscription(testCase.givenSub, r.namedLogger())

			// then
			if testCase.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				fetchedSub, err = fetchTestSubscription(testEnvironment.Context, r)
				require.NoError(t, err)
				ensureFinalizerMatch(t, &fetchedSub, []string{eventingv1alpha2.Finalizer})
			}
		})
	}
}

// helper functions and structs

// TestEnvironment provides mocked resources for tests.
type TestEnvironment struct {
	Context    context.Context
	Client     client.Client
	Backend    *mocks.Backend
	Reconciler *Reconciler
	Logger     *logger.Logger
	Recorder   *record.FakeRecorder
	Cleaner    cleaner.Cleaner
}

// setupTestEnvironment is a TestEnvironment constructor.
func setupTestEnvironment(t *testing.T, objs ...client.Object) *TestEnvironment {
	mockedBackend := &mocks.Backend{}
	ctx := context.Background()
	fakeClient := createFakeClientBuilder(t).WithObjects(objs...).Build()
	recorder := &record.FakeRecorder{}

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}
	jsCleaner := cleaner.NewJetStreamCleaner(defaultLogger)
	defaultSinkValidator := sinkv2.NewValidator(ctx, fakeClient, recorder, defaultLogger)

	r := Reconciler{
		Backend:       mockedBackend,
		Client:        fakeClient,
		logger:        defaultLogger,
		recorder:      recorder,
		sinkValidator: defaultSinkValidator,
		cleaner:       jsCleaner,
	}

	return &TestEnvironment{
		Context:    ctx,
		Client:     fakeClient,
		Backend:    mockedBackend,
		Reconciler: &r,
		Logger:     defaultLogger,
		Recorder:   recorder,
		Cleaner:    jsCleaner,
	}
}

func createFakeClientBuilder(t *testing.T) *fake.ClientBuilder {
	err := eventingv1alpha2.AddToScheme(scheme.Scheme)
	require.NoError(t, err)
	return fake.NewClientBuilder().WithScheme(scheme.Scheme)
}

func ensureFinalizerMatch(t *testing.T, subscription *eventingv1alpha2.Subscription, wantFinalizers []string) {
	if len(wantFinalizers) == 0 {
		require.Empty(t, subscription.ObjectMeta.Finalizers)
	} else {
		require.Equal(t, wantFinalizers, subscription.ObjectMeta.Finalizers)
	}
}

func fetchTestSubscription(ctx context.Context, r *Reconciler) (eventingv1alpha2.Subscription, error) {
	var fetchedSub eventingv1alpha2.Subscription
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      subscriptionName,
		Namespace: namespaceName,
	}, &fetchedSub)
	return fetchedSub, err
}

func ensureSubscriptionMatchesConditionsAndStatus(t *testing.T, subscription eventingv1alpha2.Subscription, wantConditions []eventingv1alpha2.Condition, wantStatus bool) {
	require.Equal(t, len(wantConditions), len(subscription.Status.Conditions))
	comparisonResult := eventingv1alpha2.ConditionsEquals(wantConditions, subscription.Status.Conditions)
	require.True(t, comparisonResult)
	require.Equal(t, wantStatus, subscription.Status.Ready)
}
