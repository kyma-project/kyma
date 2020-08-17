package object

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	tPodLabelKey   = "podlabelkey"
	tPodLabelValue = "podlabelvalue"
	tContainerName = "testname"
	tPortName1     = "first"
	tPortName2     = "second"
)

func TestNewDeployment(t *testing.T) {
	const img = "registry/image:tag"
	var replicas int32 = 3

	deployment := NewDeployment(tNs, tName,
		WithPort(8080, tPortName1),
		WithImage(img),
		WithEnvVar("TEST_ENV1", "val1"),
		WithPort(8081, tPortName2),
		WithProbe("/are/you/alive", 8080),
		WithEnvVar("TEST_ENV2", "val2"),
		WithPodLabel(tPodLabelKey, tPodLabelValue),
		WithName(tContainerName),
		WithMatchLabelsSelector(tPodLabelKey, tPodLabelValue),
		WithReplicas(3),
	)

	expectDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{tPodLabelKey: tPodLabelValue}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{tPodLabelKey: tPodLabelValue}},
				Spec: corev1.PodSpec{

					Containers: []corev1.Container{{
						Name:  tContainerName,
						Image: img,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          tPortName1,
						}, {
							ContainerPort: 8081,
							Name:          tPortName2,
						}},
						Env: []corev1.EnvVar{{
							Name:  "TEST_ENV1",
							Value: "val1",
						}, {
							Name:  "TEST_ENV2",
							Value: "val2",
						}},
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/are/you/alive",
									Port: intstr.FromInt(8080),
								},
							},
						},
					}},
				},
			},
		},
	}

	if d := cmp.Diff(expectDeployment, deployment); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}

func TestAppApplyExistingServiceAttributes(t *testing.T) {
	existingDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       tNs,
			Name:            tName,
			ResourceVersion: "1",
			Annotations: map[string]string{
				deploymentAnnotations[0]: "1",
				"another-annotation":     "some-value",
			},
		},
		Status: appsv1.DeploymentStatus{},
	}

	// Service with empty spec, status, annotations, ...
	deployment := NewDeployment(tNs, tName)
	ApplyExistingDeploymentAttributes(existingDeployment, deployment)

	expectDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       tNs,
			Name:            tName,
			ResourceVersion: "1",
			Annotations: map[string]string{
				deploymentAnnotations[0]: "1",
			},
		},
		Status: appsv1.DeploymentStatus{},
	}

	if d := cmp.Diff(expectDeployment, deployment); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}
