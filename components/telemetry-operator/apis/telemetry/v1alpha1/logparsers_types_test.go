package v1alpha1

import (
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLogParserGetCondition(t *testing.T) {
	exampleStatus := LogParserStatus{Conditions: []LogParserCondition{{Type: LogParserPending}}}

	tests := []struct {
		name     string
		status   LogParserStatus
		condType LogParserConditionType
		expected bool
	}{
		{
			name:     "condition exists",
			status:   exampleStatus,
			condType: LogParserPending,
			expected: true,
		},
		{
			name:     "condition does not exist",
			status:   exampleStatus,
			condType: LogParserRunning,
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

func TestLogParserSetCondition(t *testing.T) {
	condPending := LogParserCondition{Type: LogParserPending, Reason: "ForSomeReason"}
	condRunning := LogParserCondition{Type: LogParserRunning, Reason: "ForSomeOtherReason"}
	condRunningOtherReason := LogParserCondition{Type: LogParserRunning, Reason: "BecauseItIs"}

	ts := metav1.Now()
	tsLater := metav1.NewTime(ts.Add(1 * time.Minute))

	tests := []struct {
		name           string
		status         LogParserStatus
		cond           LogParserCondition
		expectedStatus LogParserStatus
	}{
		{
			name:           "set for the first time",
			status:         LogParserStatus{},
			cond:           condPending,
			expectedStatus: LogParserStatus{Conditions: []LogParserCondition{condPending}},
		},
		{
			name:           "simple set",
			status:         LogParserStatus{Conditions: []LogParserCondition{condPending}},
			cond:           condRunning,
			expectedStatus: LogParserStatus{Conditions: []LogParserCondition{condPending, condRunning}},
		},
		{
			name:           "overwrite",
			status:         LogParserStatus{Conditions: []LogParserCondition{condRunning}},
			cond:           condRunningOtherReason,
			expectedStatus: LogParserStatus{Conditions: []LogParserCondition{condRunningOtherReason}},
		},
		{
			name:           "overwrite",
			status:         LogParserStatus{Conditions: []LogParserCondition{condRunning}},
			cond:           condRunningOtherReason,
			expectedStatus: LogParserStatus{Conditions: []LogParserCondition{condRunningOtherReason}},
		},
		{
			name:           "not overwrite last transition time",
			status:         LogParserStatus{Conditions: []LogParserCondition{{Type: LogParserPending, LastTransitionTime: ts}}},
			cond:           LogParserCondition{Type: LogParserPending, LastTransitionTime: tsLater},
			expectedStatus: LogParserStatus{Conditions: []LogParserCondition{{Type: LogParserPending, LastTransitionTime: ts}}},
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
