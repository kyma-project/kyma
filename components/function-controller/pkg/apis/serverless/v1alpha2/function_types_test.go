package v1alpha2

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestFunction_TypeOf(t *testing.T) {
	testCases := map[string]struct {
		function *Function
		args     FunctionType
		want     bool
	}{
		"Have Inline Function want Inline Function": {
			args: FunctionTypeInline,
			function: &Function{Spec: FunctionSpec{Source: Source{Inline: &InlineSource{
				Source:       "aaa",
				Dependencies: "bbb",
			}}}},
			want: true,
		},
		"Have Git function want Git function": {
			args: FunctionTypeGit,
			function: &Function{Spec: FunctionSpec{Source: Source{GitRepository: &GitRepositorySource{
				URL: "bbb",
			}}}},
			want: true,
		},
		"Have Inline Function want Git Function": {
			args: FunctionTypeGit,
			function: &Function{Spec: FunctionSpec{Source: Source{Inline: &InlineSource{
				Source:       "aaa",
				Dependencies: "bbb",
			}}}},
			want: false,
		},
		"Have Git Function want Inline Function": {
			args: FunctionTypeInline,
			function: &Function{Spec: FunctionSpec{Source: Source{GitRepository: &GitRepositorySource{
				URL: "bbb",
			}}}},
			want: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {

			if got := testCase.function.TypeOf(testCase.args); got != testCase.want {
				t.Errorf("TypeOf() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestFunction_IsUpdating(t *testing.T) {

	tests := []struct {
		name     string
		function *Function
		want     bool
	}{
		{
			name: "Function is updating - running",
			function: &Function{
				Status: FunctionStatus{
					Conditions: []Condition{
						{
							Type:   ConditionBuildReady,
							Status: v1.ConditionTrue,
						},
						{
							Type:   ConditionConfigurationReady,
							Status: v1.ConditionTrue,
						},
						{
							Type:   ConditionRunning,
							Status: v1.ConditionFalse,
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Function is updating - building",
			function: &Function{
				Status: FunctionStatus{
					Conditions: []Condition{
						{
							Type:   ConditionBuildReady,
							Status: v1.ConditionFalse,
						},
						{
							Type:   ConditionConfigurationReady,
							Status: v1.ConditionTrue,
						},
						{
							Type:   ConditionRunning,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Function is not updating",
			function: &Function{
				Status: FunctionStatus{
					Conditions: []Condition{
						{
							Type:   ConditionBuildReady,
							Status: v1.ConditionTrue,
						},
						{
							Type:   ConditionConfigurationReady,
							Status: v1.ConditionTrue,
						},
						{
							Type:   ConditionRunning,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.function.IsUpdating(); got != tt.want {
				t.Errorf("Function.IsUpdating() = %v, want %v", got, tt.want)
			}
		})
	}
}
