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
			name: "Should create internal labels",
			args: args{instance: &v1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fn-name",
					UID:  "fn-uuid",
				}},
			},
			want: map[string]string{
				v1alpha2.FunctionUUIDLabel:      "fn-uuid",
				v1alpha2.FunctionManagedByLabel: v1alpha2.FunctionControllerValue,
				v1alpha2.FunctionNameLabel:      "fn-name",
				v1alpha2.FunctionResourceLabel:  v1alpha2.FunctionResourceLabelDeploymentValue,
			},
		},
		{
			name: "Should create internal and additional labels",
			args: args{instance: &v1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
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
			name: "Should create internal and from `spec.template.labels` labels",
			args: args{instance: &v1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
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
			name: "Should create internal and from `spec.labels` labels",
			args: args{instance: &v1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
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
			args: args{instance: &v1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
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
			//GIVEN
			g := gomega.NewGomegaWithT(t)
			s := &systemState{
				instance: *tt.args.instance,
			}
			//WHEN
			got := s.podLabels()
			//THEN
			g.Expect(tt.want).To(gomega.Equal(got))
		})
	}
}

func Test_systemState_podAnnotations(t *testing.T) {
	type args struct {
		instance *v1alpha2.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Should create internal annotations",
			args: args{instance: &v1alpha2.Function{}},
			want: map[string]string{
				istioConfigLabelKey: istioEnableHoldUntilProxyStartLabelValue,
			},
		},
		{
			name: "Should create internal and from `.spec.annotations` annotations",
			args: args{instance: &v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Annotations: map[string]string{
						"test-some": "test-annotation",
					},
				}}},
			want: map[string]string{
				istioConfigLabelKey: istioEnableHoldUntilProxyStartLabelValue,
				"test-some":         "test-annotation",
			},
		},
		{
			name: "Should not overwrite internal annotations",
			args: args{instance: &v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Annotations: map[string]string{
						"test-some":             "test-annotation",
						"proxy.istio.io/config": "another-config",
					},
				}}},
			want: map[string]string{
				istioConfigLabelKey: istioEnableHoldUntilProxyStartLabelValue,
				"test-some":         "test-annotation",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//GIVEN
			g := gomega.NewGomegaWithT(t)
			s := &systemState{
				instance: *tt.args.instance,
			}
			//WHEN
			got := s.podAnnotations()
			//THEN
			g.Expect(tt.want).To(gomega.Equal(got))
		})
	}
}
