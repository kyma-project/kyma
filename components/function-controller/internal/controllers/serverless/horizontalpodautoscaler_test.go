package serverless

import (
	"testing"

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
