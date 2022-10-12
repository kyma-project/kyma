package serverless

import (
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestFunctionReconciler_mergeLabels(t *testing.T) {
	type args struct {
		labelsCollection []map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "should work with empty slice",
			args: args{labelsCollection: []map[string]string{}},
			want: map[string]string{},
		},
		{
			name: "should work with 1 map as argument",
			args: args{labelsCollection: []map[string]string{{"key": "value"}}},
			want: map[string]string{"key": "value"},
		},
		{
			name: "should work with multiple maps",
			args: args{labelsCollection: []map[string]string{{"key": "value"}, {"key1": "value1"}, {"key2": "value2"}}},
			want: map[string]string{
				"key":  "value",
				"key1": "value1",
				"key2": "value2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			got := mergeLabels(tt.args.labelsCollection...)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_servicePodLabels(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha2.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Should work on function with no labels",
			args: args{instance: &serverlessv1alpha2.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			}}},
			want: map[string]string{
				serverlessv1alpha2.FunctionUUIDLabel:      "fn-uuid",
				serverlessv1alpha2.FunctionManagedByLabel: serverlessv1alpha2.FunctionControllerValue,
				serverlessv1alpha2.FunctionNameLabel:      "fn-name",
				serverlessv1alpha2.FunctionResourceLabel:  serverlessv1alpha2.FunctionResourceLabelDeploymentValue,
			},
		},
		{
			name: "Should work with function with some labels",
			args: args{instance: &serverlessv1alpha2.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			},
				Spec: serverlessv1alpha2.FunctionSpec{
					Templates: &serverlessv1alpha2.Templates{
						FunctionPod: &serverlessv1alpha2.PodTemplate{
							Metadata: &serverlessv1alpha2.MetadataTemplate{
								Labels: map[string]string{
									"test-some": "test-label",
								},
							},
						},
					}}}},
			want: map[string]string{
				serverlessv1alpha2.FunctionUUIDLabel:      "fn-uuid",
				serverlessv1alpha2.FunctionManagedByLabel: serverlessv1alpha2.FunctionControllerValue,
				serverlessv1alpha2.FunctionNameLabel:      "fn-name",
				serverlessv1alpha2.FunctionResourceLabel:  serverlessv1alpha2.FunctionResourceLabelDeploymentValue,
				"test-some":                               "test-label",
			},
		},
		{
			name: "Should not overwrite internal labels",
			args: args{instance: &serverlessv1alpha2.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			},
				Spec: serverlessv1alpha2.FunctionSpec{
					Templates: &serverlessv1alpha2.Templates{
						FunctionPod: &serverlessv1alpha2.PodTemplate{
							Metadata: &serverlessv1alpha2.MetadataTemplate{
								Labels: map[string]string{
									"test-some":                              "test-label",
									serverlessv1alpha2.FunctionResourceLabel: "job",
									serverlessv1alpha2.FunctionNameLabel:     "some-other-name",
								},
							},
						},
					}}}},
			want: map[string]string{
				serverlessv1alpha2.FunctionUUIDLabel:      "fn-uuid",
				serverlessv1alpha2.FunctionManagedByLabel: serverlessv1alpha2.FunctionControllerValue,
				serverlessv1alpha2.FunctionNameLabel:      "fn-name",
				serverlessv1alpha2.FunctionResourceLabel:  serverlessv1alpha2.FunctionResourceLabelDeploymentValue,
				"test-some":                               "test-label",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			s := &systemState{}
			got := s.podLabels()
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}
