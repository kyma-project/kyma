package controllers

import (
	"fmt"
	"time"

	serverless "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"

	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	taskrunAnnotations = map[string]string{
		"sidecar.istio.io/inject": "false",
	}
	defaultBuildTimeout = 30 * time.Minute
)

//go:generate stringer -type=TaskRunCondition -trimprefix=TaskRunCondition

type TaskRunCondition int8

const (
	TaskRunConditionCanceled TaskRunCondition = iota + 1
	TaskRunConditionFailed
	TaskRunConditionRunning
	TaskRunConditionSucceeded
)

func getTaskRunCondition(tr *tektonv1alpha1.TaskRun) TaskRunCondition {
	if tr.Spec.Status == "TaskRunCancelled" {
		return TaskRunConditionCanceled
	}

	conditionStatus := getConditionStatus(tr.Status.Conditions)

	if ConditionStatusSucceeded == conditionStatus {
		return TaskRunConditionSucceeded
	}

	if ConditionStatusFailed == conditionStatus {
		return TaskRunConditionFailed
	}

	return TaskRunConditionRunning
}

func applyLabels(uid, imageTag string, labels map[string]string) map[string]string {
	trLabels := make(map[string]string, len(labels)+2)
	for k, v := range labels {
		trLabels[k] = v
	}
	trLabels[serverless.FnUUID] = uid
	trLabels[serverless.ImageTag] = imageTag
	return trLabels
}

func newTascRunSpec(
	srcConfigmap, imageName, runtimeConfigmap string,
	limits, requests *corev1.ResourceList) *tektonv1alpha1.TaskSpec {
	sources := configmapVolumeSpec{
		name: sourceVolName,
		path: "/src",
		cmap: srcConfigmap,
	}

	workspace := configmapVolumeSpec{
		name: dockerfileVolName,
		path: "/workspace",
		cmap: runtimeConfigmap,
	}

	return &tektonv1alpha1.TaskSpec{
		Steps: []tektonv1alpha1.Step{
			{
				Container: corev1.Container{
					Name:  "executor",
					Image: kanikoExecutorImage,
					Env:   taskrunEnvs,
					VolumeMounts: []corev1.VolumeMount{
						*sources.volumeMount(),
						*workspace.volumeMount(),
					},
					Args: []string{
						fmt.Sprintf("--destination=%s", imageName),
						"--insecure",
						"--skip-tls-verify",
					},
					Resources: corev1.ResourceRequirements{
						Limits:   *limits,
						Requests: *requests,
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			*sources.volume(),
			*workspace.volume(),
		},
	}
}

func newTaskRun(
	prefix, namespace, serviceAccount string,
	spec *tektonv1alpha1.TaskSpec,
	labels map[string]string) *tektonv1alpha1.TaskRun {
	generateName := fmt.Sprintf("%s-", prefix)
	return &tektonv1alpha1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: generateName,
			Namespace:    namespace,
			Labels:       labels,
			Annotations:  taskrunAnnotations,
		},
		Spec: tektonv1alpha1.TaskRunSpec{
			ServiceAccountName: serviceAccount,
			Timeout: &metav1.Duration{
				Duration: defaultBuildTimeout,
			},
			TaskSpec: spec,
		},
	}
}

func newTaskRunFromFn(
	fn *serverless.Function,
	imageName, imageTag, serviceAccount string,
	srcConfigmap, runtimeConfigmap string,
	limits, requests *corev1.ResourceList) *tektonv1alpha1.TaskRun {
	trSpec := newTascRunSpec(srcConfigmap, imageName, runtimeConfigmap, limits, requests)
	labels := applyLabels(string(fn.UID), imageTag, fn.Labels)
	return newTaskRun(fn.Name, fn.Namespace, serviceAccount, trSpec, labels)
}
