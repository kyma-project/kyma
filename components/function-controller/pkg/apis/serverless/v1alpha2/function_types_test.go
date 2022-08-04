package v1alpha2

import (
	"testing"
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
