package serverless

import (
	"testing"

	"k8s.io/utils/pointer"

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
			name: "equal deployments",
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
								Annotations: map[string]string{
									"some-template-annotation-key": "some-template-annotation-val",
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
								Annotations: map[string]string{
									"some-template-annotation-key": "some-template-annotation-val",
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
			name: "different env",
			args: args{
				existing: appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{

								Containers: []corev1.Container{
									{
										Image: "container-image1",
										Env:   []corev1.EnvVar{{Name: "AAA", Value: "BBB"}},
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
										Env:   []corev1.EnvVar{{Name: "CCC", Value: "DDD"}},
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
				scalingEnabled: false,
			},
			want: false,
		},
		{
			name: "different fn-image",
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
										Image: "different-fn-image",
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
			want: false,
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
		{
			name: "different secret volumes",
			args: args{
				existing: appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: podSpecWithSecretVolume(),
						},
					},
				},
				expected: appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: func() corev1.PodSpec {
								volumeName := "another-volume-name"
								podSpec := podSpecWithSecretVolume()
								podSpec.Volumes[0].Name = volumeName
								podSpec.Volumes[0].Secret.SecretName = "another-secret-name"
								podSpec.Containers[0].VolumeMounts[0].Name = volumeName
								podSpec.Containers[0].VolumeMounts[0].MountPath = "/another/mount/path"
								return podSpec
							}(),
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			got := equalDeployments(tt.args.existing, tt.args.expected)
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
			name: "should return true for two empty structs",
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

func Test_equalSecretMounts(t *testing.T) {
	type args struct {
		existing corev1.PodSpec
		expected corev1.PodSpec
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should work in easy equal case",
			args: args{
				existing: podSpecWithSecretVolume(),
				expected: podSpecWithSecretVolume(),
			},
			want: true,
		},
		{
			name: "should return true for empty structs",
			args: args{
				existing: corev1.PodSpec{
					Volumes: []corev1.Volume{},
					Containers: []corev1.Container{
						{
							VolumeMounts: []corev1.VolumeMount{},
						},
					},
				},
				expected: corev1.PodSpec{
					Volumes: []corev1.Volume{},
					Containers: []corev1.Container{
						{
							VolumeMounts: []corev1.VolumeMount{},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "should detect difference between secret names",
			args: args{
				existing: func() corev1.PodSpec {
					podSpec := podSpecWithSecretVolume()
					podSpec.Volumes[0].Secret.SecretName = "secret-name-1"
					return podSpec
				}(),
				expected: func() corev1.PodSpec {
					podSpec := podSpecWithSecretVolume()
					podSpec.Volumes[0].Secret.SecretName = "secret-name-2"
					return podSpec
				}(),
			},
			want: false,
		},
		{
			name: "should detect difference between mount path",
			args: args{
				existing: func() corev1.PodSpec {
					podSpec := podSpecWithSecretVolume()
					podSpec.Containers[0].VolumeMounts[0].MountPath = "/mount/path/1"
					return podSpec
				}(),
				expected: func() corev1.PodSpec {
					podSpec := podSpecWithSecretVolume()
					podSpec.Containers[0].VolumeMounts[0].MountPath = "/mount/path/2"
					return podSpec
				}(),
			},
			want: false,
		},
		{
			name: "should ignore volumes without secret",
			args: args{
				existing: func() corev1.PodSpec {
					podSpec := podSpecWithSecretVolume()
					podSpec.Volumes = append(podSpec.Volumes, notSecretVolume())
					return podSpec
				}(),
				expected: func() corev1.PodSpec {
					podSpec := podSpecWithSecretVolume()
					anotherNotSecretVolume := notSecretVolume()
					anotherNotSecretVolume.Name = "another-not-secret-volume"
					podSpec.Volumes = append(podSpec.Volumes, anotherNotSecretVolume)
					return podSpec
				}(),
			},
			want: true,
		},
		{
			name: "should detect difference for new secret volume",
			args: args{
				existing: func() corev1.PodSpec {
					podSpec := podSpecWithSecretVolume()
					newVolume := secretVolume()
					newVolume.Name = "new-volume"
					podSpec.Volumes = append(podSpec.Volumes, newVolume)
					return podSpec
				}(),
				expected: func() corev1.PodSpec {
					podSpec := podSpecWithSecretVolume()
					podSpec.Volumes[0].Secret.SecretName = "secret-name-2"
					return podSpec
				}(),
			},
			want: false,
		},
		{
			name: "should ignore volume mounts not connected with secret volumes",
			args: args{
				existing: func() corev1.PodSpec {
					podSpec := podSpecWithSecretVolume()
					notSecretVolume := notSecretVolume()
					podSpec.Volumes = append(podSpec.Volumes, notSecretVolume)
					notSecretVolumeMount := corev1.VolumeMount{
						Name:      notSecretVolume.Name,
						MountPath: "/not/secret/volume/mount/path",
					}
					podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, notSecretVolumeMount)
					return podSpec
				}(),
				expected: func() corev1.PodSpec {
					podSpec := podSpecWithSecretVolume()
					sizeLimit := k8sresource.MustParse("350Mi")
					podSpec.Volumes = append(podSpec.Volumes,
						corev1.Volume{
							Name: "another-not-secret-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									SizeLimit: &sizeLimit,
								},
							},
						})
					return podSpec
				}(),
			},
			want: true,
		},
		{
			name: "should ignore different order of volumes",
			args: func() args {
				firstVolume := secretVolume()
				firstVolume.Name = "first"
				secondVolume := secretVolume()
				secondVolume.Name = "second"

				existing := podSpecWithSecretVolume()
				existing.Volumes = []corev1.Volume{
					firstVolume,
					secondVolume,
				}

				expected := podSpecWithSecretVolume()
				expected.Volumes = []corev1.Volume{
					secondVolume,
					firstVolume,
				}

				return args{
					existing: existing,
					expected: expected,
				}
			}(),
			want: true,
		},
		{
			name: "should ignore different order of volumes",
			args: func() args {
				firstVolume := secretVolume()
				firstVolume.Name = "first"
				secondVolume := secretVolume()
				secondVolume.Name = "second"
				volumes := []corev1.Volume{
					firstVolume,
					secondVolume,
				}

				firstMount := secretVolumeMount()
				firstMount.Name = firstVolume.Name
				secondMount := secretVolumeMount()
				secondMount.Name = secondVolume.Name

				existing := podSpecWithSecretVolume()
				existing.Volumes = volumes
				existing.Containers[0].VolumeMounts = []corev1.VolumeMount{
					firstMount,
					secondMount,
				}

				expected := podSpecWithSecretVolume()
				expected.Volumes = volumes
				expected.Containers[0].VolumeMounts = []corev1.VolumeMount{
					secondMount,
					firstMount,
				}

				return args{
					existing: existing,
					expected: expected,
				}
			}(),
			want: true,
		},
		{
			name: "should ignore different volume mounts in not first container",
			// now we works only with single container
			args: args{
				existing: func() corev1.PodSpec {
					someSecretVolumeMount := secretVolumeMount()
					someSecretVolumeMount.MountPath = "/some/secret/volume/mount"
					someSecondContainer := corev1.Container{
						VolumeMounts: []corev1.VolumeMount{
							someSecretVolumeMount,
						},
					}
					podSpec := podSpecWithSecretVolume()
					podSpec.Containers = append(podSpec.Containers, someSecondContainer)
					return podSpec
				}(),
				expected: func() corev1.PodSpec {
					anotherSecretVolumeMount := secretVolumeMount()
					anotherSecretVolumeMount.MountPath = "/another/secret/volume/mount"
					anotherSecondContainer := corev1.Container{
						VolumeMounts: []corev1.VolumeMount{
							anotherSecretVolumeMount,
						},
					}
					podSpec := podSpecWithSecretVolume()
					podSpec.Containers = append(podSpec.Containers, anotherSecondContainer)
					return podSpec
				}(),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			got := equalSecretMounts(tt.args.expected, tt.args.existing)
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
			s := systemState{
				deployments: appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						{Status: appsv1.DeploymentStatus{Conditions: tt.args.conditions}},
					},
				},
			}

			got := s.hasDeploymentConditionTrueStatus(tt.args.conditionType)
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
			s := systemState{
				deployments: appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						{
							Status: appsv1.DeploymentStatus{Conditions: tt.args.conditions},
						},
					},
				},
			}
			got := s.hasDeploymentConditionTrueStatusWithReason(tt.args.conditionType, tt.args.conditionReason)
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
			s := systemState{
				deployments: appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						{
							Status: appsv1.DeploymentStatus{Conditions: tt.args.conditions},
						},
					},
				}}
			got := s.isDeploymentReady()
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

func secretVolume() corev1.Volume {
	return corev1.Volume{
		Name: "volume-name",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  "secret-name",
				DefaultMode: pointer.Int32(0644),
				Optional:    pointer.Bool(false),
			},
		},
	}
}

func notSecretVolume() corev1.Volume {
	sizeLimit := k8sresource.MustParse("50Mi")
	return corev1.Volume{
		Name: "not-secret-volume",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				SizeLimit: &sizeLimit,
			},
		},
	}
}

func secretVolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      "volume-name",
		ReadOnly:  false,
		MountPath: "/mount/path",
	}
}

func podSpecWithSecretVolume() corev1.PodSpec {
	return corev1.PodSpec{
		Volumes: []corev1.Volume{
			secretVolume(),
		},
		Containers: []corev1.Container{
			{
				VolumeMounts: []corev1.VolumeMount{
					secretVolumeMount(),
				},
			},
		},
	}
}
