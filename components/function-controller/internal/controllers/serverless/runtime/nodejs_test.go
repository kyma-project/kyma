package runtime_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/onsi/gomega"
)

func TestNodejs_SanitizeDependencies(t *testing.T) {
	tests := []struct {
		name string
		deps string
		want string
	}{
		{
			name: "Should not touch empty deps - {}",
			deps: "{}",
			want: "{}",
		},
		{
			name: "Should not touch empty deps",
			deps: "",
			want: "{}",
		},
		{
			name: "Should not touch empty deps - empty string",
			deps: "random-string",
			want: "random-string",
		},
		{
			name: "Should not touch empty deps - empty string",
			deps: "     ",
			want: "{}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := runtime.GetRuntime(v1alpha1.Nodejs12)
			got := r.SanitizeDependencies(tt.deps)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}
