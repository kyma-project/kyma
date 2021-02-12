package serverless

import (
	"testing"

	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFunctionReconciler_equalDeployments(t *testing.T) {
	type args struct {
		existing       appsv1.Deployment
		expected       appsv1.Deployment
		scalingEnabled bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "simple case - false on empty structs",
			args: args{
				existing:       appsv1.Deployment{},
				expected:       appsv1.Deployment{},
				scalingEnabled: true,
			},
			want: false, // yes, false, as we can't compare services without spec.template.containers, it makes no sense
		},
		{
			name: "simple case",
			args: args{
				existing: appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"label-key": "label-value",
						},
					},
					Spec: appsv1.DeploymentSpec{

						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"some-template-label-key": "some-template-label-val",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: "container-image1",
										Env: []corev1.EnvVar{{
											Name:  "env-name1",
											Value: "env-value1",
										}},
										Resources: corev1.ResourceRequirements{
											Limits: map[corev1.ResourceName]k8sresource.Quantity{
												corev1.ResourceCPU:    k8sresource.MustParse("50m"),
												corev1.ResourceMemory: k8sresource.MustParse("50Mi"),
											},
											Requests: map[corev1.ResourceName]k8sresource.Quantity{
												corev1.ResourceCPU:    k8sresource.MustParse("20m"),
												corev1.ResourceMemory: k8sresource.MustParse("20Mi"),
											},
										},
									},
								},
							},
						},
					},
				},
				expected: appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"label-key": "label-value",
						},
					},
					Spec: appsv1.DeploymentSpec{

						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"some-template-label-key": "some-template-label-val",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: "container-image1",
										Env: []corev1.EnvVar{{
											Name:  "env-name1",
											Value: "env-value1",
										}},
										Resources: corev1.ResourceRequirements{
											Limits: map[corev1.ResourceName]k8sresource.Quantity{
												corev1.ResourceCPU:    k8sresource.MustParse("50m"),
												corev1.ResourceMemory: k8sresource.MustParse("50Mi"),
											},
											Requests: map[corev1.ResourceName]k8sresource.Quantity{
												corev1.ResourceCPU:    k8sresource.MustParse("20m"),
												corev1.ResourceMemory: k8sresource.MustParse("20Mi"),
											},
										},
									},
								},
							},
						},
					},
				},
				scalingEnabled: true,
			},
			want: true,
		},
		{
			name: "different labels on pods",
			args: args{
				existing: appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"label-key": "label-value",
						},
					},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"some-template-label-key": "some-template-label-val",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: "container-image1",
										Env: []corev1.EnvVar{{
											Name:  "env-name1",
											Value: "env-value1",
										}},
										Resources: corev1.ResourceRequirements{
											Limits: map[corev1.ResourceName]k8sresource.Quantity{
												corev1.ResourceCPU:    k8sresource.MustParse("50m"),
												corev1.ResourceMemory: k8sresource.MustParse("50Mi"),
											},
											Requests: map[corev1.ResourceName]k8sresource.Quantity{
												corev1.ResourceCPU:    k8sresource.MustParse("20m"),
												corev1.ResourceMemory: k8sresource.MustParse("20Mi"),
											},
										},
									},
								},
							},
						},
					},
				},
				expected: appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"label-key": "label-value",
						},
					},
					Spec: appsv1.DeploymentSpec{

						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"some-template-label-key": "different-value", // that's different
								},
							},
							Spec: corev1.PodSpec{

								Containers: []corev1.Container{
									{
										Image: "container-image1",
										Env: []corev1.EnvVar{{
											Name:  "env-name1",
											Value: "env-value1",
										}},
										Resources: corev1.ResourceRequirements{
											Limits: map[corev1.ResourceName]k8sresource.Quantity{
												corev1.ResourceCPU:    k8sresource.MustParse("50m"),
												corev1.ResourceMemory: k8sresource.MustParse("50Mi"),
											},
											Requests: map[corev1.ResourceName]k8sresource.Quantity{
												corev1.ResourceCPU:    k8sresource.MustParse("20m"),
												corev1.ResourceMemory: k8sresource.MustParse("20Mi"),
											},
										},
									},
								},
							},
						},
					},
				},
				scalingEnabled: true,
			},
			want: false,
		},
		{
			name: "different pod annotations",
			args: args{
				existing: appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"label-key": "label-value",
						},
					},
					Spec: appsv1.DeploymentSpec{

						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "container-image1"}},
							},
						},
					},
				},
				expected: appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"label-key": "label-value",
						},
					},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{
									"here's something": "that should be different than in 'existing'",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "container-image1"}}},
						},
					},
				},
				scalingEnabled: true,
			},
			want: false,
		},
		{
			name: "different resources",
			args: args{
				existing: appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{

								Containers: []corev1.Container{
									{
										Image: "container-image1",
										Resources: corev1.ResourceRequirements{
											Limits: map[corev1.ResourceName]k8sresource.Quantity{
												corev1.ResourceCPU:    k8sresource.MustParse("50m"),
												corev1.ResourceMemory: k8sresource.MustParse("50Mi"),
											},
											Requests: map[corev1.ResourceName]k8sresource.Quantity{
												corev1.ResourceCPU:    k8sresource.MustParse("20m"),
												corev1.ResourceMemory: k8sresource.MustParse("20Mi"),
											},
										},
									},
								},
							},
						},
					},
				},
				expected: appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{

						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{

								Containers: []corev1.Container{
									{
										Image: "container-image1",
										Resources: corev1.ResourceRequirements{
											Limits: map[corev1.ResourceName]k8sresource.Quantity{
												corev1.ResourceCPU:    k8sresource.MustParse("50m"),
												corev1.ResourceMemory: k8sresource.MustParse("50Mi"),
											},
											Requests: map[corev1.ResourceName]k8sresource.Quantity{
												corev1.ResourceCPU:    k8sresource.MustParse("400m"),
												corev1.ResourceMemory: k8sresource.MustParse("40Mi"),
											},
										},
									},
								},
							},
						},
					},
				},
				scalingEnabled: true,
			},
			want: false,
		},
		{
			name: "different image",
			args: args{
				existing: appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: "container-image1",
									},
								},
							},
						},
					},
				},
				expected: appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: "different-image",
									},
								},
							},
						},
					},
				},
				scalingEnabled: true,
			},
			want: false,
		},
		{
			name: "scaling enabled and replicas differ",
			args: args{
				existing:       fixDeploymentWithReplicas(1),
				expected:       fixDeploymentWithReplicas(2),
				scalingEnabled: true,
			},
			want: true,
		},
		{
			name: "scaling enabled and replicas match",
			args: args{
				existing:       fixDeploymentWithReplicas(3),
				expected:       fixDeploymentWithReplicas(3),
				scalingEnabled: true,
			},
			want: true,
		},
		{
			name: "scaling disabled and replicas differ",
			args: args{
				existing:       fixDeploymentWithReplicas(1),
				expected:       fixDeploymentWithReplicas(2),
				scalingEnabled: false,
			},
			want: false,
		},
		{
			name: "scaling disabled and replicas match",
			args: args{
				existing:       fixDeploymentWithReplicas(3),
				expected:       fixDeploymentWithReplicas(3),
				scalingEnabled: false,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.equalDeployments(tt.args.existing, tt.args.expected, tt.args.scalingEnabled)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func Test_equalResources(t *testing.T) {
	type args struct {
		existing corev1.ResourceRequirements
		expected corev1.ResourceRequirements
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should work in easy case",
			args: args{
				existing: corev1.ResourceRequirements{
					Limits: map[corev1.ResourceName]k8sresource.Quantity{
						corev1.ResourceMemory: k8sresource.MustParse("51Mi"),
						corev1.ResourceCPU:    k8sresource.MustParse("51m"),
					},
					Requests: map[corev1.ResourceName]k8sresource.Quantity{
						corev1.ResourceMemory: k8sresource.MustParse("50Mi"),
						corev1.ResourceCPU:    k8sresource.MustParse("50m"),
					},
				},
				expected: corev1.ResourceRequirements{
					Limits: map[corev1.ResourceName]k8sresource.Quantity{
						corev1.ResourceMemory: k8sresource.MustParse("51Mi"),
						corev1.ResourceCPU:    k8sresource.MustParse("51m"),
					},
					Requests: map[corev1.ResourceName]k8sresource.Quantity{
						corev1.ResourceMemory: k8sresource.MustParse("50Mi"),
						corev1.ResourceCPU:    k8sresource.MustParse("50m"),
					},
				}},
			want: true,
		},
		{
			name: "should return false if cpu values do not match ",
			args: args{
				existing: corev1.ResourceRequirements{
					Limits: map[corev1.ResourceName]k8sresource.Quantity{
						corev1.ResourceMemory: k8sresource.MustParse("51Mi"),
						corev1.ResourceCPU:    k8sresource.MustParse("51m"),
					},
					Requests: map[corev1.ResourceName]k8sresource.Quantity{
						corev1.ResourceMemory: k8sresource.MustParse("50Mi"),
						corev1.ResourceCPU:    k8sresource.MustParse("50m"),
					},
				},
				expected: corev1.ResourceRequirements{
					Limits: map[corev1.ResourceName]k8sresource.Quantity{
						corev1.ResourceMemory: k8sresource.MustParse("51Mi"),
						corev1.ResourceCPU:    k8sresource.MustParse("52m"), // this one is different
					},
					Requests: map[corev1.ResourceName]k8sresource.Quantity{
						corev1.ResourceMemory: k8sresource.MustParse("50Mi"),
						corev1.ResourceCPU:    k8sresource.MustParse("50m"),
					},
				}},
			want: false,
		},
		{
			name: "should return false if no values are provided for existing",
			args: args{
				existing: corev1.ResourceRequirements{},
				expected: corev1.ResourceRequirements{
					Limits: map[corev1.ResourceName]k8sresource.Quantity{
						corev1.ResourceMemory: k8sresource.MustParse("50Mi"),
						corev1.ResourceCPU:    k8sresource.MustParse("50m"),
					},
					Requests: map[corev1.ResourceName]k8sresource.Quantity{
						corev1.ResourceMemory: k8sresource.MustParse("51Mi"),
						corev1.ResourceCPU:    k8sresource.MustParse("51m"),
					},
				}},
			want: false,
		},
		{
			name: "should return truefor two empty structs",
			args: args{
				existing: corev1.ResourceRequirements{},
				expected: corev1.ResourceRequirements{}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			got := equalResources(tt.args.expected, tt.args.existing)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_hasDeploymentConditionTrueStatus(t *testing.T) {
	type args struct {
		conditions    []appsv1.DeploymentCondition
		conditionType appsv1.DeploymentConditionType
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "simple case",
			args: args{conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentProgressing,
				Status: corev1.ConditionTrue,
			}}, conditionType: appsv1.DeploymentProgressing},
			want: true,
		},
		{
			name: "simple case - 2 conditions",
			args: args{conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentReplicaFailure,
				Status: corev1.ConditionFalse,
			}, {
				Type:   appsv1.DeploymentProgressing,
				Status: corev1.ConditionTrue,
			}},
				conditionType: appsv1.DeploymentProgressing},
			want: true,
		},
		{
			name: "fails on empty condition",
			args: args{conditions: []appsv1.DeploymentCondition{}, conditionType: appsv1.DeploymentProgressing},
			want: false,
		},
		{
			name: "fails if there is no proper condition",
			args: args{conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentReplicaFailure,
				Status: corev1.ConditionFalse,
			}, {
				Type:   appsv1.DeploymentProgressing,
				Status: corev1.ConditionFalse,
			}},
				conditionType: appsv1.DeploymentAvailable,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.hasDeploymentConditionTrueStatus(appsv1.Deployment{
				Status: appsv1.DeploymentStatus{Conditions: tt.args.conditions},
			}, tt.args.conditionType)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_hasDeploymentConditionTrueStatusWithReason(t *testing.T) {
	type args struct {
		conditions      []appsv1.DeploymentCondition
		conditionType   appsv1.DeploymentConditionType
		conditionReason string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "simple case",
			args: args{conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentProgressing,
				Status: corev1.ConditionTrue,
				Reason: "SomeReason",
			}}, conditionType: appsv1.DeploymentProgressing, conditionReason: "SomeReason"},
			want: true,
		},
		{
			name: "simple case - 2 conditions",
			args: args{conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentReplicaFailure,
				Status: corev1.ConditionFalse,
			}, {
				Type:   appsv1.DeploymentProgressing,
				Status: corev1.ConditionTrue,
				Reason: "SomeReason",
			}},
				conditionType: appsv1.DeploymentProgressing, conditionReason: "SomeReason"},
			want: true,
		},
		{
			name: "fails on empty condition",
			args: args{conditions: []appsv1.DeploymentCondition{}, conditionType: appsv1.DeploymentProgressing},
			want: false,
		},
		{
			name: "fails if there is no proper condition",
			args: args{conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentReplicaFailure,
				Status: corev1.ConditionFalse,
			}, {
				Type:   appsv1.DeploymentProgressing,
				Status: corev1.ConditionTrue,
				Reason: "SomeReason",
			}},
				conditionType: appsv1.DeploymentAvailable, conditionReason: "SomeReason",
			},
			want: false,
		},
		{
			name: "fails if there is proper condition with wrong reason",
			args: args{conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentReplicaFailure,
				Status: corev1.ConditionFalse,
			}, {
				Type:   appsv1.DeploymentProgressing,
				Status: corev1.ConditionTrue,
				Reason: "SomeReason",
			}},
				conditionType: appsv1.DeploymentProgressing, conditionReason: "AnotherReason",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.hasDeploymentConditionTrueStatusWithReason(appsv1.Deployment{
				Status: appsv1.DeploymentStatus{Conditions: tt.args.conditions},
			}, tt.args.conditionType, tt.args.conditionReason)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_isDeploymentReady(t *testing.T) {
	type args struct {
		conditions []appsv1.DeploymentCondition
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "fail on 1 good condition",
			args: args{conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentProgressing,
				Status: corev1.ConditionTrue,
				Reason: NewRSAvailableReason,
			}}},
			want: false,
		},
		{
			name: "2 good conditions",
			args: args{conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentAvailable,
				Status: corev1.ConditionTrue,
				Reason: MinimumReplicasAvailable,
			}, {
				Type:   appsv1.DeploymentProgressing,
				Status: corev1.ConditionTrue,
				Reason: NewRSAvailableReason,
			}}},
			want: true,
		},
		{
			name: "Fails on empty condition",
			args: args{conditions: []appsv1.DeploymentCondition{}},
			want: false,
		},
		{
			name: "fails if there is one condition with wrong reason",
			args: args{conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentAvailable,
				Status: corev1.ConditionTrue,
				Reason: "WrongReason",
			}, {
				Type:   appsv1.DeploymentProgressing,
				Status: corev1.ConditionTrue,
				Reason: NewRSAvailableReason,
			}}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.isDeploymentReady(appsv1.Deployment{
				Status: appsv1.DeploymentStatus{Conditions: tt.args.conditions},
			})
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func fixDeploymentWithReplicas(replicas int32) appsv1.Deployment {
	return appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{}},
				},
			},
		},
	}
}
