package jetstream

import (
	"context"
	"testing"

	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstream"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstream/mocks"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
// and if so with how much initial delay.
// Returning error or a `Result{Requeue: true}` would cause the reconciliation to be requeued.
// Everything else is mocked since we are only interested in the logic of the Reconcile method
// and not the reconciler dependencies.
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
	missingSubSyncErr := jetstream.ErrMissingSubscription
	backendDeleteErr := errors.New("backend delete error")
	validatorErr := errors.New("invalid sink")
	happyValidator := sink.ValidatorFunc(func(s *eventingv1alpha2.Subscription) error { return nil })
	unhappyValidator := sink.ValidatorFunc(func(s *eventingv1alpha2.Subscription) error { return validatorErr })

	var testCases = []struct {
		name                 string
		givenSubscription    *eventingv1alpha2.Subscription
		givenReconcilerSetup func() (*Reconciler, *mocks.Backend)
		wantReconcileResult  ctrl.Result
		wantReconcileError   error
	}{
		{
			name:              "Return nil and default Result{} when there is no error from the reconciler dependencies",
			givenSubscription: testSub,
			givenReconcilerSetup: func() (*Reconciler, *mocks.Backend) {
				te := setupTestEnvironment(t, testSub)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				te.Backend.On("SyncSubscription", mock.Anything).Return(nil)
				te.Backend.On("GetJetStreamSubjects", mock.Anything, mock.Anything, mock.Anything).Return(
					[]string{controllertesting.JetStreamSubject})
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, te.Cleaner, happyValidator), te.Backend
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  nil,
		},
		{
			name:              "Return nil and default Result{} when the subscription does not exist on the cluster",
			givenSubscription: testSub,
			givenReconcilerSetup: func() (*Reconciler, *mocks.Backend) {
				te := setupTestEnvironment(t)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, te.Cleaner, happyValidator), te.Backend
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  nil,
		},
		{
			name:              "Return nil and default Result{} when the subscription has no finalizer",
			givenSubscription: controllertesting.NewSubscription(subscriptionName, namespaceName),
			givenReconcilerSetup: func() (*Reconciler, *mocks.Backend) {
				te := setupTestEnvironment(t, controllertesting.NewSubscription(subscriptionName, namespaceName))
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, te.Cleaner, happyValidator), te.Backend
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  nil,
		},
		{
			name:              "Return error and default Result{} when backend sync returns error",
			givenSubscription: testSub,
			givenReconcilerSetup: func() (*Reconciler, *mocks.Backend) {
				te := setupTestEnvironment(t, testSub)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				te.Backend.On("SyncSubscription", mock.Anything).Return(backendSyncErr)
				te.Backend.On("GetJetStreamSubjects", mock.Anything, mock.Anything, mock.Anything).Return(
					[]string{controllertesting.JetStreamSubject})
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, te.Cleaner, happyValidator), te.Backend
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  backendSyncErr,
		},
		{
			name: "Return nil and RequeueAfter with requeue duration when " +
				"backend sync returns missingSubscriptionErr",
			givenSubscription: testSub,
			givenReconcilerSetup: func() (*Reconciler, *mocks.Backend) {
				te := setupTestEnvironment(t, testSub)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				te.Backend.On("SyncSubscription", mock.Anything).Return(missingSubSyncErr)
				te.Backend.On("GetJetStreamSubjects", mock.Anything, mock.Anything, mock.Anything).Return(
					[]string{controllertesting.JetStreamSubject})
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, te.Cleaner, happyValidator), te.Backend
			},
			wantReconcileResult: ctrl.Result{RequeueAfter: requeueDuration},
			wantReconcileError:  nil,
		},
		{
			name:              "Return error and default Result{} when backend delete returns error",
			givenSubscription: testSubUnderDeletion,
			givenReconcilerSetup: func() (*Reconciler, *mocks.Backend) {
				te := setupTestEnvironment(t, testSubUnderDeletion)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				te.Backend.On("DeleteSubscription", mock.Anything).Return(backendDeleteErr)
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, te.Cleaner, happyValidator), te.Backend
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  errFailedToDeleteSub,
		},
		{
			name:              "Return error and default Result{} when validator returns error",
			givenSubscription: testSub,
			givenReconcilerSetup: func() (*Reconciler, *mocks.Backend) {
				te := setupTestEnvironment(t, testSub)
				te.Backend.On("Initialize", mock.Anything).Return(nil)
				te.Backend.On("DeleteSubscriptionsOnly", mock.Anything).Return(nil)
				te.Backend.On("GetJetStreamSubjects", mock.Anything, mock.Anything, mock.Anything).Return(
					[]string{controllertesting.JetStreamSubject})
				return NewReconciler(ctx, te.Client, te.Backend, te.Logger, te.Recorder, te.Cleaner, unhappyValidator), te.Backend
			},
			wantReconcileResult: ctrl.Result{},
			wantReconcileError:  validatorErr,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// given
			reconciler, mockedBackend := tc.givenReconcilerSetup()
			r := ctrl.Request{NamespacedName: types.NamespacedName{
				Namespace: tc.givenSubscription.Namespace,
				Name:      tc.givenSubscription.Name,
			}}

			// when
			res, err := reconciler.Reconcile(context.Background(), r)

			// then
			req.Equal(res, tc.wantReconcileResult)
			req.ErrorIs(err, tc.wantReconcileError)
			mockedBackend.AssertExpectations(t)
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
		wantError       error
	}{
		{
			name:            "With no finalizers the NATS subscription should not be deleted",
			givenFinalizers: []string{},
			wantDeleteCall:  false,
			wantFinalizers:  []string{},
			wantError:       nil,
		},
		{
			name: "With eventing finalizer the NATS subscription should be deleted" +
				" and the finalizer should be cleared",
			givenFinalizers: []string{eventingv1alpha2.Finalizer},
			wantDeleteCall:  true,
			wantFinalizers:  []string{},
			wantError:       nil,
		},
		{
			name:            "With wrong finalizer the NATS subscription should not be deleted",
			givenFinalizers: []string{"eventing2.kyma-project.io"},
			wantDeleteCall:  false,
			wantFinalizers:  []string{"eventing2.kyma-project.io"},
			wantError:       nil,
		},
		{
			name:            "With Delete Subscription returning an error, the finalizer still remains",
			givenFinalizers: []string{eventingv1alpha2.Finalizer},
			wantDeleteCall:  true,
			wantFinalizers:  []string{eventingv1alpha2.Finalizer},
			wantError:       errFailedToDeleteSub,
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

			if testCase.wantDeleteCall {
				if errors.Is(testCase.wantError, errFailedToDeleteSub) {
					mockedBackend.On("DeleteSubscription", sub).Return(errors.New("deletion error"))
				} else {
					mockedBackend.On("DeleteSubscription", sub).Return(nil)
				}
			}

			// when
			_, err = r.handleSubscriptionDeletion(ctx, sub, r.namedLogger())
			require.ErrorIs(t, err, testCase.wantError)

			// then
			mockedBackend.AssertExpectations(t)

			ensureFinalizerMatch(t, sub, testCase.wantFinalizers)

			// check the changes were made on the kubernetes server
			fetchedSub, err := fetchTestSubscription(ctx, r)
			require.NoError(t, err)
			ensureFinalizerMatch(t, &fetchedSub, testCase.wantFinalizers)

			// clean up finalizers first before deleting sub
			fetchedSub.ObjectMeta.Finalizers = nil
			err = r.Client.Update(ctx, &fetchedSub)
			require.NoError(t, err)
			err = r.Client.Delete(ctx, &fetchedSub)
			require.NoError(t, err)
		})
	}
}

func Test_addFinalizer(t *testing.T) {
	// given
	ctx := context.Background()
	eventingFinalizer := []string{eventingv1alpha2.Finalizer}
	var emptyFinalizer []string
	sub := controllertesting.NewSubscription(subscriptionName, namespaceName)
	testEnvironment := setupTestEnvironment(t, sub)
	r := testEnvironment.Reconciler
	fakeSub := controllertesting.NewSubscription("fake", namespaceName)

	testCases := []struct {
		name             string
		givenFinalizers  []string
		wantErrorMessage string
	}{
		{
			name:            "A Subscription should be updated with the eventing finalizer if it is missing",
			givenFinalizers: emptyFinalizer,
		},
		{
			name:            "addFinalizer() should propagate the error returned by the update operation",
			givenFinalizers: eventingFinalizer,
			wantErrorMessage: apierr.NewNotFound(
				schema.GroupResource{
					Group:    "eventing.kyma-project.io",
					Resource: "subscriptions",
				},
				fakeSub.Name).Error(),
		},
	}

	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			// in case syncSubscriptionStatus throws an error
			if tC.wantErrorMessage != "" {
				_, err := r.addFinalizer(ctx, fakeSub)
				require.ErrorContains(t, err, tC.wantErrorMessage)
				return
			}
			sub.Finalizers = tC.givenFinalizers
			_, err := r.addFinalizer(ctx, sub)
			require.NoError(t, err)

			fetchedSub, err := fetchTestSubscription(testEnvironment.Context, r)
			require.NoError(t, err)

			if tC.wantErrorMessage != "" {
				ensureFinalizerMatch(t, &fetchedSub, emptyFinalizer)
			} else {
				ensureFinalizerMatch(t, &fetchedSub, eventingFinalizer)
			}
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
		name           string
		givenSub       *eventingv1alpha2.Subscription
		givenError     error
		wantConditions []eventingv1alpha2.Condition
		wantStatus     bool
	}{
		{
			name: "Subscription with no conditions should stay not ready with false condition",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha2.Condition{}),
				controllertesting.WithStatus(true),
			),
			givenError:     jetStreamError,
			wantConditions: []eventingv1alpha2.Condition{falseNatsSubActiveCondition},
			wantStatus:     false,
		},
		{
			name: "Subscription with false ready condition should stay not ready with false condition",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha2.Condition{falseNatsSubActiveCondition}),
				controllertesting.WithStatus(false),
			),
			givenError:     jetStreamError,
			wantConditions: []eventingv1alpha2.Condition{falseNatsSubActiveCondition},
			wantStatus:     false,
		},
		{
			name: "Subscription should become ready because because, there is no reconciliation error",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha2.Condition{falseNatsSubActiveCondition}),
				controllertesting.WithStatus(false),
			),
			givenError:     nil,
			wantConditions: []eventingv1alpha2.Condition{trueNatsSubActiveCondition},
			wantStatus:     true,
		},
		{
			name: "Subscription should stay with ready condition and status",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha2.Condition{trueNatsSubActiveCondition}),
				controllertesting.WithStatus(true),
			),
			givenError:     nil,
			wantConditions: []eventingv1alpha2.Condition{trueNatsSubActiveCondition},
			wantStatus:     true,
		},
		{
			name: "Subscription should become not ready because of the occurred error",
			givenSub: controllertesting.NewSubscription(subscriptionName, namespaceName,
				controllertesting.WithConditions([]eventingv1alpha2.Condition{trueNatsSubActiveCondition}),
				controllertesting.WithStatus(true),
			),
			givenError:     jetStreamError,
			wantConditions: []eventingv1alpha2.Condition{falseNatsSubActiveCondition},
			wantStatus:     false,
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
			err = r.syncSubscriptionStatus(ctx, sub, testCase.givenError, r.namedLogger())
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

func Test_syncEventTypes(t *testing.T) {
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
		},
	}
	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			// given
			sub := testCase.givenSub

			// when
			r.syncEventTypes(sub)

			// then
			require.Equal(t, testCase.wantSubStatus.Types, sub.Status.Types)
			require.Equal(t, testCase.wantSubStatus.Backend, sub.Status.Backend)
			require.Equal(t, testCase.wantSubStatus.Ready, sub.Status.Ready)
		})
	}
}

func Test_updateStatus(t *testing.T) {
	sub := controllertesting.NewSubscription(subscriptionName, namespaceName, controllertesting.WithStatus(true))

	ctx := context.Background()
	testEnvironment := setupTestEnvironment(t, sub)
	r := testEnvironment.Reconciler

	testCases := []struct {
		name       string
		oldStatus  eventingv1alpha2.SubscriptionStatus
		newStatus  eventingv1alpha2.SubscriptionStatus
		wantError  error
		wantChange bool
	}{
		{
			name:       "No update should happen in case when the statuses are equal",
			oldStatus:  eventingv1alpha2.SubscriptionStatus{Ready: true},
			newStatus:  eventingv1alpha2.SubscriptionStatus{Ready: true},
			wantChange: false,
		},
		{
			name:       "An Update should happen in case when the statuses are equal",
			oldStatus:  eventingv1alpha2.SubscriptionStatus{Ready: true},
			newStatus:  eventingv1alpha2.SubscriptionStatus{Ready: false},
			wantChange: true,
		},
		{
			name:       "The error message from the update operation should fail the whole function",
			oldStatus:  eventingv1alpha2.SubscriptionStatus{Ready: true},
			newStatus:  eventingv1alpha2.SubscriptionStatus{Ready: false},
			wantChange: true,
			wantError:  errFailedToUpdateStatus,
		},
	}

	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			// given
			resourceVersionBefore := sub.ResourceVersion
			sub.Status = tC.oldStatus
			newSub := sub.DeepCopy()
			newSub.Status = tC.newStatus

			// simulate the update error
			if tC.wantError != nil {
				require.NoError(t, r.Client.Delete(ctx, sub))
			}

			// when
			updateStatusErr := r.updateStatus(ctx, sub, newSub, r.namedLogger())

			// then
			// in case the function should fail
			if tC.wantError != nil {
				require.ErrorIs(t, updateStatusErr, tC.wantError)
				return
			}

			// then
			require.NoError(t, updateStatusErr)
			fetchedSub, err := fetchTestSubscription(testEnvironment.Context, r)
			require.NoError(t, err)

			require.Equal(t, tC.wantChange, fetchedSub.ResourceVersion != resourceVersionBefore)
			require.Equal(t, newSub.Status.Ready, fetchedSub.Status.Ready)
			require.Equal(t, newSub.Status.Conditions, fetchedSub.Status.Conditions)
		})
	}
}

func Test_updateSubscription(t *testing.T) {
	// given
	sub := controllertesting.NewSubscription(subscriptionName, namespaceName,
		controllertesting.WithStatusTypes([]eventingv1alpha2.EventType{
			{OriginalType: "order.created.v1", CleanType: "order.created.v1"},
		}),
	)

	subWithDifferentStatus := sub.DeepCopy()
	subWithDifferentStatus.Status.Ready = !sub.Status.Ready

	ctx := context.Background()
	testEnvironment := setupTestEnvironment(t)
	r := testEnvironment.Reconciler

	testCases := []struct {
		name             string
		actualSub        *eventingv1alpha2.Subscription
		desiredSub       *eventingv1alpha2.Subscription
		wantErrorMessage string
		wantChange       bool
		wantSubStatus    *eventingv1alpha2.SubscriptionStatus
	}{
		{
			name:       "When Subscription cannot be fetched from the api server the function should return an error",
			actualSub:  sub,
			desiredSub: sub,
			wantErrorMessage: apierr.NewNotFound(
				eventingv1alpha2.SubscriptionGroupVersionResource().GroupResource(),
				subscriptionName).Error(),
		},
		{
			name:       "updateSubscriptionStatus() should be idempotent",
			actualSub:  sub,
			desiredSub: sub,
			wantChange: false,
		},
		{
			name:       "Subscription's status should be updated on difference",
			actualSub:  sub,
			desiredSub: subWithDifferentStatus,
			wantChange: true,
		},
	}

	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			if tC.wantErrorMessage == "" {
				sub.ResourceVersion = ""
				require.NoError(t, r.Client.Create(ctx, sub))
			}
			resourceVersionBefore := sub.ResourceVersion

			// when
			err := r.updateSubscriptionStatus(ctx, tC.desiredSub, r.namedLogger())

			// then
			// in case the subscription doesn't exist
			if tC.wantErrorMessage != "" {
				require.ErrorContains(t, err, testCase.wantErrorMessage)
				return
			}

			// then
			fetchedSub, err := fetchTestSubscription(testEnvironment.Context, r)
			require.NoError(t, err)

			require.Equal(t, tC.wantChange, fetchedSub.ResourceVersion != resourceVersionBefore)
			ensureFinalizerMatch(t, &fetchedSub, testCase.desiredSub.Finalizers)
			require.Equal(t, testCase.desiredSub.Status.Ready, fetchedSub.Status.Ready)
			require.Equal(t, testCase.desiredSub.Status.Conditions, fetchedSub.Status.Conditions)

			// clean up
			require.NoError(t, r.Client.Delete(ctx, sub))
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
	defaultSinkValidator := sink.NewValidator(ctx, fakeClient, recorder)

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

func ensureSubscriptionMatchesConditionsAndStatus(t *testing.T,
	subscription eventingv1alpha2.Subscription, wantConditions []eventingv1alpha2.Condition, wantStatus bool) {
	require.Equal(t, len(wantConditions), len(subscription.Status.Conditions))
	comparisonResult := eventingv1alpha2.ConditionsEquals(wantConditions, subscription.Status.Conditions)
	require.True(t, comparisonResult)
	require.Equal(t, wantStatus, subscription.Status.Ready)
}
