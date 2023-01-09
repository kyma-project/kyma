package jetstream

import (
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	subtesting "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func Test_isInDeletion(t *testing.T) {
	testCases := []struct {
		name       string
		givenSub   *v1alpha2.Subscription
		wantResult bool
	}{
		{
			name:       "subscription with no deletion timestamp",
			givenSub:   subtesting.NewSubscription("test", "test"),
			wantResult: false,
		},
		{
			name: "subscription with deletion timestamp",
			givenSub: subtesting.NewSubscription("test", "test",
				subtesting.WithNonZeroDeletionTimestamp()),
			wantResult: true,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// when
			result := isInDeletion(tc.givenSub)

			// then
			require.Equal(t, tc.wantResult, result)
		})
	}
}

func Test_containsFinalizer(t *testing.T) {
	testCases := []struct {
		name       string
		givenSub   *v1alpha2.Subscription
		wantResult bool
	}{
		{
			name: "subscription containing finalizer",
			givenSub: subtesting.NewSubscription("test", "test",
				subtesting.WithFinalizers([]string{v1alpha2.Finalizer})),
			wantResult: true,
		},
		{
			name: "subscription containing finalizer",
			givenSub: subtesting.NewSubscription("test", "test",
				subtesting.WithFinalizers([]string{"invalid"})),
			wantResult: false,
		},
		{
			name:       "subscription not containing finalizer",
			givenSub:   subtesting.NewSubscription("test", "test"),
			wantResult: false,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// when
			result := containsFinalizer(tc.givenSub)

			// then
			require.Equal(t, tc.wantResult, result)
		})
	}
}

func Test_setSubReadyStatus(t *testing.T) {
	testCases := []struct {
		name         string
		givenSub     *v1alpha2.Subscription
		givenIsReady bool
		wantResult   bool
	}{
		{
			name: "subscription ready status not changed",
			givenSub: subtesting.NewSubscription("test", "test",
				subtesting.WithStatus(true)),
			givenIsReady: true,
			wantResult:   false,
		},
		{
			name: "subscription ready status not changed",
			givenSub: subtesting.NewSubscription("test", "test",
				subtesting.WithStatus(false)),
			givenIsReady: false,
			wantResult:   false,
		},
		{
			name: "subscription ready status changed",
			givenSub: subtesting.NewSubscription("test", "test",
				subtesting.WithStatus(true)),
			givenIsReady: false,
			wantResult:   true,
		},
		{
			name: "subscription ready status changed",
			givenSub: subtesting.NewSubscription("test", "test",
				subtesting.WithStatus(false)),
			givenIsReady: true,
			wantResult:   true,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// when
			result := setSubReadyStatus(&tc.givenSub.Status, tc.givenIsReady)

			// then
			require.Equal(t, tc.wantResult, result)
			require.Equal(t, tc.givenSub.Status.Ready, tc.givenIsReady)
		})
	}
}

func Test_initializeDesiredConditions(t *testing.T) {
	// given
	expectedConditions := []v1alpha2.Condition{v1alpha2.MakeCondition(
		v1alpha2.ConditionSubscriptionActive,
		v1alpha2.ConditionReasonNATSSubscriptionNotActive,
		corev1.ConditionFalse, "")}

	// when
	conditions := initializeDesiredConditions()

	// then
	require.Equal(t, len(conditions), 1)
	require.True(t, v1alpha2.ConditionsEquals(conditions, expectedConditions))
}

func Test_setConditionSubscriptionActive(t *testing.T) {
	err := errors.New("some error")
	condition1 := v1alpha2.MakeCondition(
		v1alpha2.ConditionSubscriptionActive,
		v1alpha2.ConditionReasonNATSSubscriptionNotActive,
		corev1.ConditionFalse, "")
	condition2 := v1alpha2.MakeCondition(
		v1alpha2.ConditionSubscribed,
		v1alpha2.ConditionReasonSubscriptionDeleted,
		corev1.ConditionFalse, "")
	conditionReady := v1alpha2.MakeCondition(
		v1alpha2.ConditionSubscriptionActive,
		v1alpha2.ConditionReasonNATSSubscriptionActive,
		corev1.ConditionTrue, "")
	conditionNotReady := v1alpha2.MakeCondition(
		v1alpha2.ConditionSubscriptionActive,
		v1alpha2.ConditionReasonNATSSubscriptionNotActive,
		corev1.ConditionFalse, err.Error())

	testCases := []struct {
		name            string
		givenConditions []v1alpha2.Condition
		givenError      error
		wantConditions  []v1alpha2.Condition
	}{
		{
			name:            "no error should set the condition to ready",
			givenConditions: []v1alpha2.Condition{condition1, condition2},
			givenError:      nil,
			wantConditions:  []v1alpha2.Condition{conditionReady, condition2},
		},
		{
			name:            "error should set the condition to ready",
			givenConditions: []v1alpha2.Condition{condition1, condition2},
			givenError:      err,
			wantConditions:  []v1alpha2.Condition{conditionNotReady, condition2},
		},
		{
			name:            "if condition is not present, do nothing",
			givenConditions: []v1alpha2.Condition{condition2},
			givenError:      err,
			wantConditions:  []v1alpha2.Condition{condition2},
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// when
			setConditionSubscriptionActive(tc.givenConditions, tc.givenError)

			// then
			require.True(t, v1alpha2.ConditionsEquals(tc.givenConditions, tc.wantConditions))
		})
	}
}
