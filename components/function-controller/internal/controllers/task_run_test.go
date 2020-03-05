package controllers

import (
	"testing"

	"github.com/onsi/gomega"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck/v1beta1"
)

func TestGetTaskRunCondition(t *testing.T) {
	testCases := []struct {
		tr       *tektonv1alpha1.TaskRun
		expected TaskRunCondition
	}{
		{
			expected: TaskRunConditionSucceeded,
			tr: &tektonv1alpha1.TaskRun{
				Status: tektonv1alpha1.TaskRunStatus{
					Status: v1beta1.Status{
						Conditions: v1beta1.Conditions{
							apis.Condition{
								Type:   apis.ConditionSucceeded,
								Status: v1.ConditionTrue,
							},
							apis.Condition{
								Type:   apis.ConditionReady,
								Status: v1.ConditionTrue,
							},
						},
					},
				},
			},
		},
		{
			expected: TaskRunConditionRunning,
			tr:       &tektonv1alpha1.TaskRun{},
		},
		{
			expected: TaskRunConditionCanceled,
			tr: &tektonv1alpha1.TaskRun{
				Spec: tektonv1alpha1.TaskRunSpec{
					Status: "TaskRunCancelled",
				},
			},
		},
		{
			expected: TaskRunConditionFailed,
			tr: &tektonv1alpha1.TaskRun{
				Status: tektonv1alpha1.TaskRunStatus{
					Status: v1beta1.Status{
						Conditions: v1beta1.Conditions{
							apis.Condition{
								Type:   apis.ConditionSucceeded,
								Status: v1.ConditionFalse,
							},
							apis.Condition{
								Type:   apis.ConditionReady,
								Status: v1.ConditionTrue,
							},
						},
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.expected.String(), func(t *testing.T) {
			g := gomega.NewWithT(t)

			actual := getTaskRunCondition(tC.tr)
			g.Expect(actual).Should(gomega.Equal(tC.expected))
		})
	}
}
