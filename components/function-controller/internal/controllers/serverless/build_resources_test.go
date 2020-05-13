package serverless

import (
	"testing"

	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/serving/pkg/apis/autoscaling"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func TestFunctionReconciler_buildConfigMap(t *testing.T) {
	tests := []struct {
		name string
		fn   *serverlessv1alpha1.Function
		want v1.ConfigMap
	}{
		{
			name: "should properly build configmap",
			fn: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "fn-ns",
					UID:       "fn-uuid",
					Name:      "function-name",
				},
				Spec: serverlessv1alpha1.FunctionSpec{Source: "fn-source", Deps: ""},
			},
			want: v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    "fn-ns",
					GenerateName: "function-name-",
					Labels: map[string]string{
						serverlessv1alpha1.FunctionManagedByLabel: "function-controller",
						serverlessv1alpha1.FunctionNameLabel:      "function-name",
						serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
					},
				},
				Data: map[string]string{
					"handler.main": "handler.main",
					"handler.js":   "fn-source",
					"package.json": "{}",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.buildConfigMap(tt.fn)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_sanitizeDependencies(t *testing.T) {
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
			r := &FunctionReconciler{}
			got := r.sanitizeDependencies(tt.deps)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

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
			r := &FunctionReconciler{}
			got := r.mergeLabels(tt.args.labelsCollection...)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_internalFunctionLabels(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha1.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "should return only 3 correct labels",
			args: args{instance: &serverlessv1alpha1.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			}}},
			want: map[string]string{
				serverlessv1alpha1.FunctionManagedByLabel: "function-controller",
				serverlessv1alpha1.FunctionNameLabel:      "fn-name",
				serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.internalFunctionLabels(tt.args.instance)
			g.Expect(got).To(gomega.Equal(tt.want))
			g.Expect(got).To(gomega.HaveLen(3))
		})
	}
}

func TestFunctionReconciler_servicePodLabels(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha1.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Should work on function with no labels",
			args: args{instance: &serverlessv1alpha1.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			}}},
			want: map[string]string{
				serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
				serverlessv1alpha1.FunctionManagedByLabel: "function-controller",
				serverlessv1alpha1.FunctionNameLabel:      "fn-name",
			},
		},
		{
			name: "Should work with function with some labels",
			args: args{instance: &serverlessv1alpha1.Function{ObjectMeta: metav1.ObjectMeta{
				Name: "fn-name",
				UID:  "fn-uuid",
			},
				Spec: serverlessv1alpha1.FunctionSpec{
					Labels: map[string]string{
						"test-some": "test-label",
					},
				}}},
			want: map[string]string{
				serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
				serverlessv1alpha1.FunctionManagedByLabel: "function-controller",
				serverlessv1alpha1.FunctionNameLabel:      "fn-name",
				"test-some":                               "test-label",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.servicePodLabels(tt.args.instance)
			g.Expect(got).To(gomega.Equal(tt.want))
			g.Expect(got).To(gomega.HaveLen(len(tt.args.instance.Spec.Labels) + 3))
		})
	}
}

func TestFunctionReconciler_servicePodAnnotations(t *testing.T) {
	two := int32(2)
	three := int32(3)

	type args struct {
		instance *serverlessv1alpha1.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "should return correct annotations with function that has no set replicas",
			args: args{instance: &serverlessv1alpha1.Function{}},
			want: map[string]string{
				autoscaling.MinScaleAnnotationKey: "1",
				autoscaling.MaxScaleAnnotationKey: "1",
			},
		},
		{
			name: "should return correct annotations with function that has set replicas",
			args: args{instance: &serverlessv1alpha1.Function{Spec: serverlessv1alpha1.FunctionSpec{
				MinReplicas: &two,
				MaxReplicas: &three,
			}}},
			want: map[string]string{
				autoscaling.MinScaleAnnotationKey: "2",
				autoscaling.MaxScaleAnnotationKey: "3",
			},
		},
		{
			name: "should return correct annotations with function that has no minreplicas set",
			args: args{instance: &serverlessv1alpha1.Function{Spec: serverlessv1alpha1.FunctionSpec{
				MaxReplicas: &three,
			}}},
			want: map[string]string{
				autoscaling.MinScaleAnnotationKey: "1",
				autoscaling.MaxScaleAnnotationKey: "3",
			},
		},
		{
			name: "should return correct annotations with function that has no maxreplicas set",
			args: args{instance: &serverlessv1alpha1.Function{Spec: serverlessv1alpha1.FunctionSpec{
				MinReplicas: &two,
			}}},
			want: map[string]string{
				autoscaling.MinScaleAnnotationKey: "2",
				autoscaling.MaxScaleAnnotationKey: "1", // TODO yes, here we have a nasty bug that will be handled when we'll create webhooks and it'll be validated there
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.servicePodAnnotations(tt.args.instance)
			g.Expect(got).To(gomega.Equal(tt.want))
			g.Expect(got).To(gomega.HaveLen(2))
		})
	}
}

func TestFunctionReconciler_serviceLabels(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha1.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.serviceLabels(tt.args.instance)
			g.Expect(got).To(gomega.Equal(tt.want))
			g.Expect(got).To(gomega.HaveLen(len(tt.args.instance.Labels) + 4))
		})
	}
}

func TestFunctionReconciler_functionLabels(t *testing.T) {
	type args struct {
		instance *serverlessv1alpha1.Function
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "should return fn labels + 3 internal ones",
			args: args{
				instance: &serverlessv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fn-name",
						UID:  "fn-uuid",
						Labels: map[string]string{
							"some-key": "whatever-value",
						}},
				},
			},
			want: map[string]string{
				serverlessv1alpha1.FunctionManagedByLabel: "function-controller",
				serverlessv1alpha1.FunctionNameLabel:      "fn-name",
				serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
				"some-key":                                "whatever-value",
			},
		}, {
			name: "should return 3 internal ones if there's no labels on fn",
			args: args{
				instance: &serverlessv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fn-name",
						UID:  "fn-uuid",
					},
				}},
			want: map[string]string{
				serverlessv1alpha1.FunctionManagedByLabel: "function-controller",
				serverlessv1alpha1.FunctionNameLabel:      "fn-name",
				serverlessv1alpha1.FunctionUUIDLabel:      "fn-uuid",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.functionLabels(tt.args.instance)
			g.Expect(got).To(gomega.Equal(tt.want))
			g.Expect(got).To(gomega.HaveLen(len(tt.args.instance.Labels) + 3))
		})
	}
}
