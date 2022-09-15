//go:build unit
// +build unit

package v1alpha1_test

import (
	"reflect"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

func Test_InitializeSubscriptionConditions(t *testing.T) {
	var tests = []struct {
		name            string
		givenConditions []api.Condition
	}{
		{
			name: "Conditions empty",
			givenConditions: func() []api.Condition {
				return api.MakeSubscriptionConditions()
			}(),
		},
		{
			name: "Conditions partially initialized",
			givenConditions: func() []api.Condition {
				// on purpose, we only set one condition
				return []api.Condition{
					{
						Type:               api.ConditionSubscribed,
						LastTransitionTime: metav1.Now(),
						Status:             corev1.ConditionUnknown,
					},
				}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			g := NewGomegaWithT(t)
			s := api.SubscriptionStatus{}
			s.Conditions = tt.givenConditions
			wantConditionTypes := []api.ConditionType{api.ConditionSubscribed, api.ConditionSubscriptionActive, api.ConditionAPIRuleStatus, api.ConditionWebhookCallStatus}

			// when
			s.InitializeConditions()

			// then
			g.Expect(s.Conditions).To(HaveLen(len(wantConditionTypes)))
			foundConditionTypes := make([]api.ConditionType, 0)
			for _, condition := range s.Conditions {
				g.Expect(condition.Status).To(BeEquivalentTo(corev1.ConditionUnknown))
				foundConditionTypes = append(foundConditionTypes, condition.Type)
			}
			g.Expect(wantConditionTypes).To(ConsistOf(foundConditionTypes))
		})
	}
}

func Test_IsReady(t *testing.T) {
	testCases := []struct {
		name            string
		givenConditions []api.Condition
		wantReadyStatus bool
	}{
		{
			name:            "should not be ready if conditions are nil",
			givenConditions: nil,
			wantReadyStatus: false,
		},
		{
			name:            "should not be ready if conditions are empty",
			givenConditions: []api.Condition{{}},
			wantReadyStatus: false,
		},
		{
			name:            "should not be ready if only ConditionSubscribed is available and true",
			givenConditions: []api.Condition{{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue}},
			wantReadyStatus: false,
		},
		{
			name:            "should not be ready if only ConditionSubscriptionActive is available and true",
			givenConditions: []api.Condition{{Type: api.ConditionSubscriptionActive, Status: corev1.ConditionTrue}},
			wantReadyStatus: false,
		},
		{
			name:            "should not be ready if only ConditionAPIRuleStatus is available and true",
			givenConditions: []api.Condition{{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionTrue}},
			wantReadyStatus: false,
		},
		{
			name: "should not be ready if all conditions are unknown",
			givenConditions: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionUnknown},
				{Type: api.ConditionSubscriptionActive, Status: corev1.ConditionUnknown},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionUnknown},
			},
			wantReadyStatus: false,
		},
		{
			name: "should not be ready if all conditions are false",
			givenConditions: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionFalse},
				{Type: api.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionFalse},
			},
			wantReadyStatus: false,
		},
		{
			name: "should be ready if all conditions are true",
			givenConditions: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: api.ConditionSubscriptionActive, Status: corev1.ConditionTrue},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
				{Type: api.ConditionWebhookCallStatus, Status: corev1.ConditionTrue},
			},
			wantReadyStatus: true,
		},
	}

	status := api.SubscriptionStatus{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status.Conditions = tc.givenConditions
			if gotReadyStatus := status.IsReady(); tc.wantReadyStatus != gotReadyStatus {
				t.Errorf("Subscription status is not valid, want: %v but got: %v", tc.wantReadyStatus, gotReadyStatus)
			}
		})
	}
}

func Test_FindCondition(t *testing.T) {
	currentTime := metav1.NewTime(time.Now())

	testCases := []struct {
		name              string
		givenConditions   []api.Condition
		findConditionType api.ConditionType
		wantCondition     *api.Condition
	}{
		{
			name: "should be able to find the present condition",
			givenConditions: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue, LastTransitionTime: currentTime},
				{Type: api.ConditionSubscriptionActive, Status: corev1.ConditionTrue, LastTransitionTime: currentTime},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionTrue, LastTransitionTime: currentTime},
				{Type: api.ConditionWebhookCallStatus, Status: corev1.ConditionTrue, LastTransitionTime: currentTime},
			},
			findConditionType: api.ConditionSubscriptionActive,
			wantCondition:     &api.Condition{Type: api.ConditionSubscriptionActive, Status: corev1.ConditionTrue, LastTransitionTime: currentTime},
		},
		{
			name: "should not be able to find the non-present condition",
			givenConditions: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue, LastTransitionTime: currentTime},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionTrue, LastTransitionTime: currentTime},
				{Type: api.ConditionWebhookCallStatus, Status: corev1.ConditionTrue, LastTransitionTime: currentTime},
			},
			findConditionType: api.ConditionSubscriptionActive,
			wantCondition:     nil,
		},
	}

	status := api.SubscriptionStatus{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status.Conditions = tc.givenConditions

			if gotCondition := status.FindCondition(tc.findConditionType); !reflect.DeepEqual(tc.wantCondition, gotCondition) {
				t.Errorf("Subscription FindCondition failed, want: %v but got: %v", tc.wantCondition, gotCondition)
			}
		})
	}
}

func Test_ShouldUpdateReadyStatus(t *testing.T) {
	testCases := []struct {
		name                   string
		subscriptionReady      bool
		subscriptionConditions []api.Condition
		wantStatus             bool
	}{
		{
			name:              "should not update if the subscription is ready and the conditions are ready",
			subscriptionReady: true,
			subscriptionConditions: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: api.ConditionSubscriptionActive, Status: corev1.ConditionTrue},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
				{Type: api.ConditionWebhookCallStatus, Status: corev1.ConditionTrue},
			},
			wantStatus: false,
		},
		{
			name:              "should not update if the subscription is not ready and the conditions are not ready",
			subscriptionReady: false,
			subscriptionConditions: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionFalse},
				{Type: api.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionFalse},
				{Type: api.ConditionWebhookCallStatus, Status: corev1.ConditionFalse},
			},
			wantStatus: false,
		},
		{
			name:              "should update if the subscription is not ready and the conditions are ready",
			subscriptionReady: false,
			subscriptionConditions: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: api.ConditionSubscriptionActive, Status: corev1.ConditionTrue},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
				{Type: api.ConditionWebhookCallStatus, Status: corev1.ConditionTrue},
			},
			wantStatus: true,
		},
		{
			name:              "should update if the subscription is ready and the conditions are not ready",
			subscriptionReady: true,
			subscriptionConditions: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionFalse},
				{Type: api.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionFalse},
				{Type: api.ConditionWebhookCallStatus, Status: corev1.ConditionFalse},
			},
			wantStatus: true,
		},
		{
			name:              "should update if the subscription is ready and some of the conditions are missing",
			subscriptionReady: true,
			subscriptionConditions: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionUnknown},
			},
			wantStatus: true,
		},
		{
			name:              "should not update if the subscription is not ready and some of the conditions are missing",
			subscriptionReady: false,
			subscriptionConditions: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionUnknown},
			},
			wantStatus: false,
		},
		{
			name:              "should update if the subscription is ready and the status of the conditions are unknown",
			subscriptionReady: true,
			subscriptionConditions: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionUnknown},
				{Type: api.ConditionSubscriptionActive, Status: corev1.ConditionUnknown},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionUnknown},
				{Type: api.ConditionWebhookCallStatus, Status: corev1.ConditionUnknown},
			},
			wantStatus: true,
		},
	}

	status := api.SubscriptionStatus{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status.Conditions = tc.subscriptionConditions
			status.Ready = tc.subscriptionReady
			if gotStatus := status.ShouldUpdateReadyStatus(); tc.wantStatus != gotStatus {
				t.Errorf("ShouldUpdateReadyStatus is not valid, want: %v but got: %v", tc.wantStatus, gotStatus)
			}
		})
	}
}

func Test_conditionsEquals(t *testing.T) {
	testCases := []struct {
		name            string
		conditionsSet1  []api.Condition
		conditionsSet2  []api.Condition
		wantEqualStatus bool
	}{
		{
			name: "should not be equal if the number of conditions are not equal",
			conditionsSet1: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue},
			},
			conditionsSet2:  []api.Condition{},
			wantEqualStatus: false,
		},
		{
			name: "should be equal if the conditions are the same",
			conditionsSet1: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
			},
			conditionsSet2: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
			},
			wantEqualStatus: true,
		},
		{
			name: "should not be equal if the condition types are different",
			conditionsSet1: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
			},
			conditionsSet2: []api.Condition{
				{Type: api.ConditionWebhookCallStatus, Status: corev1.ConditionTrue},
				{Type: api.ConditionSubscriptionActive, Status: corev1.ConditionTrue},
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the condition types are the same but the status is different",
			conditionsSet1: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue},
			},
			conditionsSet2: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionFalse},
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the condition types are different but the status is the same",
			conditionsSet1: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionFalse},
			},
			conditionsSet2: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the condition types are different and an empty key is referenced",
			conditionsSet1: []api.Condition{
				{Type: api.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
			},
			conditionsSet2: []api.Condition{
				{Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
				{Type: api.ConditionControllerReady, Status: corev1.ConditionTrue},
			},
			wantEqualStatus: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if gotEqualStatus := api.ConditionsEquals(tc.conditionsSet1, tc.conditionsSet2); tc.wantEqualStatus != gotEqualStatus {
				t.Errorf("The list of conditions are not equal, want: %v but got: %v", tc.wantEqualStatus, gotEqualStatus)
			}
		})
	}
}

func Test_conditionEquals(t *testing.T) {
	testCases := []struct {
		name            string
		condition1      api.Condition
		condition2      api.Condition
		wantEqualStatus bool
	}{
		{
			name: "should not be equal if the types are the same but the status is different",
			condition1: api.Condition{
				Type: api.ConditionSubscribed, Status: corev1.ConditionTrue,
			},

			condition2: api.Condition{
				Type: api.ConditionSubscribed, Status: corev1.ConditionUnknown,
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the types are different but the status is the same",
			condition1: api.Condition{
				Type: api.ConditionSubscribed, Status: corev1.ConditionTrue,
			},

			condition2: api.Condition{
				Type: api.ConditionAPIRuleStatus, Status: corev1.ConditionTrue,
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the message fields are different",
			condition1: api.Condition{
				Type: api.ConditionSubscribed, Status: corev1.ConditionTrue, Message: "",
			},

			condition2: api.Condition{
				Type: api.ConditionSubscribed, Status: corev1.ConditionTrue, Message: "some message",
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the reason fields are different",
			condition1: api.Condition{
				Type: api.ConditionSubscribed, Status: corev1.ConditionTrue, Reason: api.ConditionReasonSubscriptionDeleted,
			},

			condition2: api.Condition{
				Type: api.ConditionSubscribed, Status: corev1.ConditionTrue, Reason: api.ConditionReasonSubscriptionActive,
			},
			wantEqualStatus: false,
		},
		{
			name: "should be equal if all the fields are the same",
			condition1: api.Condition{
				Type:    api.ConditionAPIRuleStatus,
				Status:  corev1.ConditionFalse,
				Reason:  api.ConditionReasonAPIRuleStatusNotReady,
				Message: "API Rule is not ready",
			},
			condition2: api.Condition{
				Type:    api.ConditionAPIRuleStatus,
				Status:  corev1.ConditionFalse,
				Reason:  api.ConditionReasonAPIRuleStatusNotReady,
				Message: "API Rule is not ready",
			},
			wantEqualStatus: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if gotEqualStatus := api.ConditionEquals(tc.condition1, tc.condition2); tc.wantEqualStatus != gotEqualStatus {
				t.Errorf("The conditions are not equal, want: %v but got: %v", tc.wantEqualStatus, gotEqualStatus)
			}
		})
	}
}
