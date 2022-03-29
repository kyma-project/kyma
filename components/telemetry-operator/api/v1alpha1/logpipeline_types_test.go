package v1alpha1

import (
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/apis/apps"
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
	tests := []struct {
		name           string
		status         LogPipelineStatus
		cond           LogPipelineCondition
		expectedStatus LogPipelineStatus
	}{
		{
			name:           "set for the first time",
			status:         &apps.DeploymentStatus{},
			cond:           condAvailable(),
			expectedStatus: LogPipelineStatus{Conditions: []apps.DeploymentCondition{condAvailable()}},
		},
		{
			name:           "simple set",
			status:         &apps.DeploymentStatus{Conditions: []apps.DeploymentCondition{condProgressing()}},
			cond:           LogPipelineCondition{Type: LogPipelinePending},
			expectedStatus: LogPipelineStatus{Conditions: []LogPipelineCondition{{Type: LogPipelinePending}}},
		},
		{
			name:           "overwrite",
			status:         &apps.DeploymentStatus{Conditions: []apps.DeploymentCondition{condProgressing()}},
			cond:           condProgressing2(),
			expectedStatus: LogPipelineStatus{Conditions: []apps.DeploymentCondition{condProgressing2()}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.status.SetCondition(test.status, test.cond)
			if !reflect.DeepEqual(test.status, test.expectedStatus) {
				t.Errorf("%s: expected status: %v, got: %v", test.name, test.expectedStatus, test.status)
			}
		})
	}
}
