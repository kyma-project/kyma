package serverless

import (
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_systemState_podLabels(t *testing.T) {
	type args struct {
		instance *v1alpha2.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Should work on function with no labels",
			args: args{instance: &v1alpha2.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			}}},
			want: map[string]string{
				v1alpha2.FunctionUUIDLabel:      "fn-uuid",
				v1alpha2.FunctionManagedByLabel: v1alpha2.FunctionControllerValue,
				v1alpha2.FunctionNameLabel:      "fn-name",
				v1alpha2.FunctionResourceLabel:  v1alpha2.FunctionResourceLabelDeploymentValue,
			},
		},
		{
			name: "Should work with function with some labels",
			args: args{instance: &v1alpha2.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			},
				Spec: v1alpha2.FunctionSpec{
					Template: &v1alpha2.Template{
						Labels: map[string]string{
							"test-some": "test-label",
						},
					},
					Labels: map[string]string{
						"test-another": "test-another-label",
					},
				}}},
			want: map[string]string{
				v1alpha2.FunctionUUIDLabel:      "fn-uuid",
				v1alpha2.FunctionManagedByLabel: v1alpha2.FunctionControllerValue,
				v1alpha2.FunctionNameLabel:      "fn-name",
				v1alpha2.FunctionResourceLabel:  v1alpha2.FunctionResourceLabelDeploymentValue,
				"test-some":                     "test-label",
				"test-another":                  "test-another-label",
			},
		},
		{
			name: "Should work with function with some labels from spec.template.labels",
			args: args{instance: &v1alpha2.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			},
				Spec: v1alpha2.FunctionSpec{
					Template: &v1alpha2.Template{
						Labels: map[string]string{
							"test-some": "test-label",
						},
					},
				}}},
			want: map[string]string{
				v1alpha2.FunctionUUIDLabel:      "fn-uuid",
				v1alpha2.FunctionManagedByLabel: v1alpha2.FunctionControllerValue,
				v1alpha2.FunctionNameLabel:      "fn-name",
				v1alpha2.FunctionResourceLabel:  v1alpha2.FunctionResourceLabelDeploymentValue,
				"test-some":                     "test-label",
			},
		},
		{
			name: "Should work with function with some labels from spec.labels",
			args: args{instance: &v1alpha2.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			},
				Spec: v1alpha2.FunctionSpec{
					Labels: map[string]string{
						"test-some": "test-label",
					},
				}}},
			want: map[string]string{
				v1alpha2.FunctionUUIDLabel:      "fn-uuid",
				v1alpha2.FunctionManagedByLabel: v1alpha2.FunctionControllerValue,
				v1alpha2.FunctionNameLabel:      "fn-name",
				v1alpha2.FunctionResourceLabel:  v1alpha2.FunctionResourceLabelDeploymentValue,
				"test-some":                     "test-label",
			},
		},
		{
			name: "Should not overwrite internal labels",
			args: args{instance: &v1alpha2.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			},
				Spec: v1alpha2.FunctionSpec{
					Template: &v1alpha2.Template{
						Labels: map[string]string{
							"test-some":                    "test-label",
							v1alpha2.FunctionResourceLabel: "job",
							v1alpha2.FunctionNameLabel:     "some-other-name",
						},
					},
					Labels: map[string]string{
						"test-another":                 "test-label",
						v1alpha2.FunctionResourceLabel: "another-job",
						v1alpha2.FunctionNameLabel:     "another-name",
					},
				}}},
			want: map[string]string{
				v1alpha2.FunctionUUIDLabel:      "fn-uuid",
				v1alpha2.FunctionManagedByLabel: v1alpha2.FunctionControllerValue,
				v1alpha2.FunctionNameLabel:      "fn-name",
				v1alpha2.FunctionResourceLabel:  v1alpha2.FunctionResourceLabelDeploymentValue,
				"test-some":                     "test-label",
				"test-another":                  "test-label",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			s := &systemState{
				instance: *tt.args.instance,
			}
			got := s.podLabels()
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}
