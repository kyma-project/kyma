package serverless

import (
	"testing"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/onsi/gomega"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_equalInt32Pointer(t *testing.T) {
	one := int32(1)
	two := int32(2)

	type args struct {
		first  *int32
		second *int32
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "two nils",
			args: args{
				first:  nil,
				second: nil,
			},
			want: true,
		},
		{
			name: "one nil, one value",
			args: args{
				first:  &one,
				second: nil,
			},
			want: false,
		},
		{
			name: "two same values",
			args: args{
				first:  &one,
				second: &one,
			},
			want: true,
		},
		{
			name: "two different values",
			args: args{
				first:  &one,
				second: &two,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := equalInt32Pointer(tt.args.first, tt.args.second); got != tt.want {
				t.Errorf("equalInt32Pointer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionReconciler_equalHorizontalPodAutoscalers(t *testing.T) {
	fifty := int32(50)
	two := int32(2)

	type args struct {
		existing autoscalingv1.HorizontalPodAutoscaler
		expected autoscalingv1.HorizontalPodAutoscaler
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should be equal in simple case",
			args: args{
				existing: autoscalingv1.HorizontalPodAutoscaler{
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &two,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &fifty,
					},
				},
				expected: autoscalingv1.HorizontalPodAutoscaler{
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &two,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &fifty,
					},
				},
			},
			want: true,
		},
		{
			name: "should be equal when labels are provided",
			args: args{
				existing: autoscalingv1.HorizontalPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"label1": "value",
						},
					},
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &two,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &fifty,
					},
				},
				expected: autoscalingv1.HorizontalPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"label1": "value",
						},
					},
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &two,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &fifty,
					},
				},
			},
			want: true,
		},
		{
			name: "should return false if labels are different",
			args: args{
				existing: autoscalingv1.HorizontalPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"label1": "value",
						},
					},
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &two,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &fifty,
					},
				},
				expected: autoscalingv1.HorizontalPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"label2": "value",
						},
					},
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &two,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &fifty,
					},
				},
			},
			want: false,
		},
		{
			name: "should be false if minReplicas are different",
			args: args{
				existing: autoscalingv1.HorizontalPodAutoscaler{
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &fifty,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &fifty,
					},
				},
				expected: autoscalingv1.HorizontalPodAutoscaler{
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &two,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &fifty,
					},
				},
			},
			want: false,
		},
		{
			name: "should be false if cpuUtil is different",
			args: args{
				existing: autoscalingv1.HorizontalPodAutoscaler{
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &two,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &fifty,
					},
				},
				expected: autoscalingv1.HorizontalPodAutoscaler{
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &two,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &two,
					},
				},
			},
			want: false,
		},
		{
			name: "should be false if ref name is different",
			args: args{
				existing: autoscalingv1.HorizontalPodAutoscaler{
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &two,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &fifty,
						ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
							Kind:       "Deployment",
							Name:       "deploy2",
							APIVersion: "apps/v1",
						},
					},
				},
				expected: autoscalingv1.HorizontalPodAutoscaler{
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &two,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &fifty,
						ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
							Kind:       "Deployment",
							Name:       "deploy1",
							APIVersion: "apps/v1",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "should be true if ref name is the same",
			args: args{
				existing: autoscalingv1.HorizontalPodAutoscaler{
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &two,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &fifty,
						ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
							Kind:       "Deployment",
							Name:       "deploy-name",
							APIVersion: "apps/v1",
						},
					},
				},
				expected: autoscalingv1.HorizontalPodAutoscaler{
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MinReplicas:                    &two,
						MaxReplicas:                    10,
						TargetCPUUtilizationPercentage: &fifty,
						ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
							Kind:       "Deployment",
							Name:       "deploy-name",
							APIVersion: "apps/v1",
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.equalHorizontalPodAutoscalers(tt.args.existing, tt.args.expected)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func Test_isScalingEnabled(t *testing.T) {
	type args struct {
		minReplicas int32
		maxReplicas int32
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "scaling enabled",
			args: args{
				minReplicas: 1,
				maxReplicas: 2,
			},
			want: true,
		},
		{
			name: "scaling disabled",
			args: args{
				minReplicas: 1,
				maxReplicas: 1,
			},
			want: false,
		},
		{
			name: "scaling disabled with multiple replicas",
			args: args{
				minReplicas: 5,
				maxReplicas: 5,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := &serverlessv1alpha1.Function{
				Spec: serverlessv1alpha1.FunctionSpec{
					MinReplicas: &tt.args.minReplicas,
					MaxReplicas: &tt.args.maxReplicas,
				},
			}

			if got := isScalingEnabled(instance); got != tt.want {
				t.Errorf("isScalingEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionReconciler_isOnHorizontalPodAutoscalerChange(t *testing.T) {
	testName := "test"
	deploys := []appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: testName,
			},
		},
	}
	equalFunction := newFixFunction(testName, testName, 1, 2)
	equalHPA := (&FunctionReconciler{}).buildHorizontalPodAutoscaler(equalFunction, testName)

	type args struct {
		instance    *serverlessv1alpha1.Function
		hpas        []autoscalingv1.HorizontalPodAutoscaler
		deployments []appsv1.Deployment
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "scaling enabled and equal HPA",
			args: args{
				deployments: deploys,
				instance:    equalFunction,
				hpas: []autoscalingv1.HorizontalPodAutoscaler{
					equalHPA,
				},
			},
			want: false,
		},
		{
			name: "scaling disabled and no HPA",
			args: args{
				deployments: deploys,
				instance:    newFixFunction(testName, testName, 2, 2),
				hpas:        []autoscalingv1.HorizontalPodAutoscaler{},
			},
			want: false,
		},
		{
			name: "no deployments",
			args: args{
				deployments: []appsv1.Deployment{},
			},
			want: false,
		},
		{
			name: "scaling enabled and no HPA",
			args: args{
				deployments: deploys,
				instance:    newFixFunction(testName, testName, 1, 2),
				hpas:        []autoscalingv1.HorizontalPodAutoscaler{},
			},
			want: true,
		},
		{
			name: "scaling enabled and more than one HPA",
			args: args{
				deployments: deploys,
				instance:    newFixFunction(testName, testName, 1, 2),
				hpas: []autoscalingv1.HorizontalPodAutoscaler{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "hpa-1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "hpa-2",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "scaling enabled and unequal HPA",
			args: args{
				deployments: deploys,
				instance:    newFixFunction(testName, testName, 1, 2),
				hpas: []autoscalingv1.HorizontalPodAutoscaler{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "hpa-1",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "scaling disabled and HPA exists",
			args: args{
				deployments: deploys,
				instance:    newFixFunction(testName, testName, 2, 2),
				hpas: []autoscalingv1.HorizontalPodAutoscaler{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "hpa-1",
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &FunctionReconciler{}
			if got := r.isOnHorizontalPodAutoscalerChange(tt.args.instance, tt.args.hpas, tt.args.deployments); got != tt.want {
				t.Errorf("isOnHorizontalPodAutoscalerChange() = %v, want %v", got, tt.want)
			}
		})
	}
}
