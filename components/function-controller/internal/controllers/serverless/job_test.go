package serverless

import (
	"testing"

	"github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestFunctionReconciler_equalJobs(t *testing.T) {
	type args struct {
		existing batchv1.Job
		expected batchv1.Job
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "two jobs with same container args are same",
			args: args{
				existing: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "--destination=123"}}},
						},
					},
				}},
				expected: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "--destination=123"}}},
						},
					},
				}},
			},
			want: true,
		},
		{
			name: "two jobs with same, multiple container args are same",
			args: args{
				existing: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "arg2", "--destination=123"}}},
						},
					},
				}},
				expected: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "arg2", "--destination=123"}}},
						},
					},
				}},
			},
			want: true,
		},
		{
			name: "two jobs with different length of args are same when destination is same",
			args: args{
				existing: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "--destination=123"}}},
						},
					},
				}},
				expected: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "args2", "--destination=123"}}},
						},
					},
				}},
			},
			want: true,
		},
		{
			name: "two jobs with different arg are the same when destination is same",
			args: args{
				existing: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "--destination=123"}}},
						},
					},
				}},
				expected: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"argument-number-1", "--destination=123"}}},
						},
					},
				}},
			},
			want: true,
		},
		{
			name: "two jobs with different destination arg are the not same",
			args: args{
				existing: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "--destination=1223"}}},
						},
					},
				}},
				expected: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "--destination=123"}}},
						},
					},
				}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			got := equalJobs(tt.args.existing, tt.args.expected)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_getArg(t *testing.T) {
	tests := []struct {
		name string
		args []string
		arg  string
		want string
	}{
		{
			name: "return argument when exist",
			args: []string{"--arg1=123", "--arg2", "--destination=1234"},
			arg:  "--destination",
			want: "--destination=1234",
		},
		{
			name: "return empty when not exist",
			args: []string{"--arg1=123", "--arg2"},
			arg:  "--destination",
			want: "",
		},
		{
			name: "return empty when no arguments passed",
			args: nil,
			arg:  "--destination",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			got := getArg(tt.args, tt.arg)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}
