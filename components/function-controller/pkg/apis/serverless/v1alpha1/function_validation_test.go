package v1alpha1

//func TestFunctionSpec_validateResources(t *testing.T) {
//	tests := []struct {
//		name      string
//		resources corev1.ResourceRequirements
//		want      *apis.FieldError
//	}{
//		{
//			name: "does not error on requests=limits",
//			want: nil,
//			resources: corev1.ResourceRequirements{
//				Limits: corev1.ResourceList{
//					corev1.ResourceCPU:    resource.MustParse("50m"),
//					corev1.ResourceMemory: resource.MustParse("50Mi"),
//				},
//				Requests: corev1.ResourceList{
//					corev1.ResourceCPU:    resource.MustParse("50m"),
//					corev1.ResourceMemory: resource.MustParse("50Mi"),
//				},
//			},
//		},
//		{
//			name: "returns both errors when requests' cpu and memory are higher",
//			want: nil,
//			resources: corev1.ResourceRequirements{
//				Limits: corev1.ResourceList{
//					corev1.ResourceCPU:    resource.MustParse("50m"),
//					corev1.ResourceMemory: resource.MustParse("50Mi"),
//				},
//				Requests: corev1.ResourceList{
//					corev1.ResourceCPU:    resource.MustParse("51m"),
//					corev1.ResourceMemory: resource.MustParse("51Mi"),
//				},
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			spec := &FunctionSpec{
//				Resources: tt.resources,
//			}
//			got := spec.validateResources()
//
//			gm := gomega.NewGomegaWithT(t)
//			gm.Expect(got).To(gomega.Equal(tt.want))
//			// TODO check if there's mention of both errors in `got`
//		})
//	}
//}
