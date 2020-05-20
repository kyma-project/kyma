package v1alpha1

import (
	"testing"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFunctionSpec_validateResources(t *testing.T) {
	minusOne := int32(-1)
	one := int32(1)
	for testName, testData := range map[string]struct {
		givenFunc     *Function
		expectedError gomega.OmegaMatcher
	}{
		"Should be ok": {
			givenFunc: &Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: "test-source",
				},
			},
			expectedError: gomega.BeNil(),
		},
		"Should validate all fields without error": {
			givenFunc: &Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: "test-source",
					Deps:   " { test }     \t\n",
					Env: []corev1.EnvVar{
						{
							Name:  "test",
							Value: "test",
						},
						{
							Name:  "config",
							Value: "test",
						},
					},
					Labels: map[string]string{
						"shoul-be-ok": "test",
						"test":        "test",
					},
					MinReplicas: &one,
					MaxReplicas: &one,
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("50m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
					},
				},
			},
			expectedError: gomega.BeNil(),
		},
		"Should return errors on empty function": {
			givenFunc:     &Function{},
			expectedError: gomega.HaveOccurred(),
		},
		"Should return error on deps validation": {
			givenFunc: &Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: "test-source",
					Deps:   "{",
				},
			},
			expectedError: gomega.HaveOccurred(),
		},
		"Should return error on env validation": {
			givenFunc: &Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: "test-source",
					Env: []corev1.EnvVar{
						{
							Name:  "test",
							Value: "test",
						},
						{
							Name:  "K_CONFIGURATION",
							Value: "should reject this",
						},
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
		},
		"Should return error on labels validation": {
			givenFunc: &Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: "test-source",
					Labels: map[string]string{
						"shoul-be-ok":      "test",
						"should BE not OK": "test",
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
		},
		"Should return error on replicas validation": {
			givenFunc: &Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source:      "test-source",
					MinReplicas: &one,
					MaxReplicas: &minusOne,
				},
			},
			expectedError: gomega.HaveOccurred(),
		},
		"Should return error on resources validation": {
			givenFunc: &Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: "test-source",
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("50m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// given
			g := gomega.NewWithT(t)

			// when
			errs := testData.givenFunc.Validate(nil)

			//then
			g.Expect(errs).To(testData.expectedError)
		})
	}
}
