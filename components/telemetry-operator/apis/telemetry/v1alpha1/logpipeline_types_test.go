package v1alpha1

import (
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetCondition(t *testing.T) {
	exampleStatus := LogPipelineStatus{Conditions: []LogPipelineCondition{{Type: LogPipelinePending}}}

	tests := []struct {
		name     string
		status   LogPipelineStatus
		condType LogPipelineConditionType
		expected bool
	}{
		{
			name:     "condition exists",
			status:   exampleStatus,
			condType: LogPipelinePending,
			expected: true,
		},
		{
			name:     "condition does not exist",
			status:   exampleStatus,
			condType: LogPipelineRunning,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cond := test.status.GetCondition(test.condType)
			exists := cond != nil
			if exists != test.expected {
				t.Errorf("%s: expected condition to exist: %t, got: %t", test.name, test.expected, exists)
			}
		})
	}
}

func TestSetCondition(t *testing.T) {
	condPending := LogPipelineCondition{Type: LogPipelinePending, Reason: "ForSomeReason"}
	condRunning := LogPipelineCondition{Type: LogPipelineRunning, Reason: "ForSomeOtherReason"}
	condRunningOtherReason := LogPipelineCondition{Type: LogPipelineRunning, Reason: "BecauseItIs"}

	ts := metav1.Now()
	tsLater := metav1.NewTime(ts.Add(1 * time.Minute))

	tests := []struct {
		name           string
		status         LogPipelineStatus
		cond           LogPipelineCondition
		expectedStatus LogPipelineStatus
	}{
		{
			name:           "set for the first time",
			status:         LogPipelineStatus{},
			cond:           condPending,
			expectedStatus: LogPipelineStatus{Conditions: []LogPipelineCondition{condPending}},
		},
		{
			name:           "simple set",
			status:         LogPipelineStatus{Conditions: []LogPipelineCondition{condPending}},
			cond:           condRunning,
			expectedStatus: LogPipelineStatus{Conditions: []LogPipelineCondition{condPending, condRunning}},
		},
		{
			name:           "overwrite",
			status:         LogPipelineStatus{Conditions: []LogPipelineCondition{condRunning}},
			cond:           condRunningOtherReason,
			expectedStatus: LogPipelineStatus{Conditions: []LogPipelineCondition{condRunningOtherReason}},
		},
		{
			name:           "overwrite",
			status:         LogPipelineStatus{Conditions: []LogPipelineCondition{condRunning}},
			cond:           condRunningOtherReason,
			expectedStatus: LogPipelineStatus{Conditions: []LogPipelineCondition{condRunningOtherReason}},
		},
		{
			name:           "not overwrite last transition time",
			status:         LogPipelineStatus{Conditions: []LogPipelineCondition{{Type: LogPipelinePending, LastTransitionTime: ts}}},
			cond:           LogPipelineCondition{Type: LogPipelinePending, LastTransitionTime: tsLater},
			expectedStatus: LogPipelineStatus{Conditions: []LogPipelineCondition{{Type: LogPipelinePending, LastTransitionTime: ts}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.status.SetCondition(test.cond)
			if !reflect.DeepEqual(test.status, test.expectedStatus) {
				t.Errorf("%s: expected status: %v, got: %v", test.name, test.expectedStatus, test.status)
			}
		})
	}
}
