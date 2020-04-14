package controllers

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/onsi/gomega"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	imgName        = "testImgName"
	secret         = "testSecret"
	serviceAccount = "serviceAccount"
	key            = "FUNC_TIMEOUT"
	value          = "190"
)

func envs() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  key,
			Value: value,
		},
	}
}

func TestNewPodSpec(t *testing.T) {
	g := gomega.NewWithT(t)

	envs := envs()

	podSpec := newPodSpec(
		imgName,
		secret,
		serviceAccount,
		envs...,
	)

	g.Expect(podSpec.Containers).Should(gomega.HaveLen(1))

	for _, envar := range append(envVarsForRevision, envs...) {
		desc := fmt.Sprintf("env var [%s=%s] should be found", envar.Name, envar.Value)
		t.Run(desc, func(t *testing.T) {
			g := gomega.NewWithT(t)
			g.Expect(podSpec.Containers[0].Env).Should(gomega.ContainElement(envar))
		})
	}
}

func TestEqual(t *testing.T) {
	testCases := []struct {
		desc     string
		l, r     []corev1.EnvVar
		expected bool
	}{
		{
			l:        []corev1.EnvVar{},
			r:        []corev1.EnvVar{},
			expected: true,
		},
		{ // 1
			l: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "v2",
					Value: "v2",
				},
				corev1.EnvVar{
					Name:  "v1",
					Value: "v1",
				},
			},
			r: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "v1",
					Value: "v1",
				},
				corev1.EnvVar{
					Name:  "v2",
					Value: "v2",
				},
			},
			expected: true,
		},
		{ // 2
			l: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "v1",
					Value: "v1",
				},
			},
			r: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "v2",
					Value: "v2",
				},
			},
			expected: false,
		},
		{ // 3
			l: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "v1",
					Value: "diff",
				},
			},
			r: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "v2",
					Value: "v2",
				},
			},
			expected: false,
		},
		{ // 4
			l: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "v1",
					Value: "v1",
				},
			},
			r: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "v1",
					Value: "v1",
				},
				corev1.EnvVar{
					Name:  "v2",
					Value: "v2",
				},
			},
			expected: false,
		},
		{ // 5
			l: []corev1.EnvVar{},
			r: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "v1",
					Value: "v1",
				},
				corev1.EnvVar{
					Name:  "v2",
					Value: "v2",
				},
			},
			expected: false,
		},
		{ // 6
			l: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "v1",
					Value: "v1",
				},
				corev1.EnvVar{
					Name:  "v2",
					Value: "v2",
				},
			},
			r:        []corev1.EnvVar{},
			expected: false,
		},
		{ // 7
			l: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "v1",
					Value: "v1",
				},
				corev1.EnvVar{
					Name:  "v2",
					Value: "v2",
				},
			},
			r: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "v1",
					Value: "v1",
				},
				corev1.EnvVar{
					Name:  "v1",
					Value: "change",
				},
				corev1.EnvVar{
					Name:  "v2",
					Value: "me plz",
				},
				corev1.EnvVar{
					Name:  "v1",
					Value: "v1",
				},
				corev1.EnvVar{
					Name:  "v2",
					Value: "v2",
				},
			},
			expected: true,
		},
	}
	for i, tC := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := gomega.NewWithT(t)
			actual := equal(tC.l, tC.r)
			g.Expect(actual).To(gomega.Equal(tC.expected))
		})
	}
}

// returns knative serving that does not have lambda container
func invalidServing(imageTag string) *servingv1.Service {
	return &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"imageTag": imageTag,
			},
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{
								corev1.Container{
									Name: "test",
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestShouldUpdateServingContainerNameNotFound(t *testing.T) {
	// given
	imageTag := "123"
	svc := invalidServing(imageTag)

	g := gomega.NewWithT(t)

	// when
	_, err := shouldUpdateServing(testLog, svc, make([]corev1.EnvVar, 0), imageTag)

	// then
	g.Expect(err).Should(gomega.HaveOccurred())
}
