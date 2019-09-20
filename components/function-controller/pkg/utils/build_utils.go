/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"os"
	"time"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var buildTimeout = os.Getenv("BUILD_TIMEOUT")

var defaultMode = int32(420)

func GetBuildResource(rnInfo *RuntimeInfo, fn *serverlessv1alpha1.Function, imageName string, buildName string) *buildv1alpha1.Build {
	args := []buildv1alpha1.ArgumentSpec{}
	args = append(args, buildv1alpha1.ArgumentSpec{Name: "IMAGE", Value: imageName})

	envs := []corev1.EnvVar{}

	timeout, err := time.ParseDuration(buildTimeout)
	if err != nil {
		timeout = 30 * time.Minute
	}

	for _, rt := range rnInfo.AvailableRuntimes {
		if rt.ID == fn.Spec.Runtime {
			args = append(args, buildv1alpha1.ArgumentSpec{Name: "DOCKERFILE", Value: rt.DockerfileName})
		}
	}

	vols := []corev1.Volume{
		{
			Name: "source",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &defaultMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fn.Name,
					},
				},
			},
		},
	}

	b := buildv1alpha1.Build{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "build.knative.dev/v1alpha1",
			Kind:       "Build",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildName,
			Namespace: fn.Namespace,
			Labels:    fn.Labels,
		},
		Spec: buildv1alpha1.BuildSpec{
			ServiceAccountName: rnInfo.ServiceAccount,
			Template: &buildv1alpha1.TemplateInstantiationSpec{
				Name:      "function-kaniko",
				Kind:      buildv1alpha1.BuildTemplateKind,
				Arguments: args,
				Env:       envs,
			},
			Volumes: vols,
		},
	}

	if b.Spec.Timeout == nil {
		b.Spec.Timeout = &metav1.Duration{Duration: timeout}
	}

	return &b
}

func GetBuildTemplateSpec(rnInfo *RuntimeInfo) buildv1alpha1.BuildTemplateSpec {
	parameters := []buildv1alpha1.ParameterSpec{
		{
			Name:        "IMAGE",
			Description: "The name of the image to push",
		},
		{
			Name:        "DOCKERFILE",
			Description: "name of the configmap that contains the Dockerfile",
		},
	}

	destination := "--destination=${IMAGE}"
	steps := []corev1.Container{
		{
			Name:  "build-and-push",
			Image: "gcr.io/kaniko-project/executor",
			Args: []string{
				"--dockerfile=/workspace/Dockerfile",
				destination,
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "${DOCKERFILE}",
					MountPath: "/workspace",
				},
				{
					Name:      "source",
					MountPath: "/src",
				},
			},
		},
	}

	var vol corev1.Volume
	var volumes []corev1.Volume
	for _, rt := range rnInfo.AvailableRuntimes {
		vol.Name = rt.DockerfileName
		vol.VolumeSource.ConfigMap = &corev1.ConfigMapVolumeSource{
			DefaultMode:          &defaultMode,
			LocalObjectReference: corev1.LocalObjectReference{Name: rt.DockerfileName},
		}
		volumes = append(volumes, vol)
	}

	bt := buildv1alpha1.BuildTemplateSpec{
		Parameters: parameters,
		Steps:      steps,
		Volumes:    volumes,
	}

	return bt
}
