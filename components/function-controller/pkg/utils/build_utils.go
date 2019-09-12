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
	"fmt"
	"os"
	"time"

	"github.com/gogo/protobuf/proto"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var buildTimeout = os.Getenv("BUILD_TIMEOUT")

const (
	// Timeout after which a TaskRun gets canceled. Can be overridden by the BUILD_TIMEOUT env
	// var.
	defaultBuildTimeout = 30 * time.Minute
	// Kaniko executor image used to build Function container images.
	// https://github.com/GoogleContainerTools/kaniko/blob/master/deploy/Dockerfile
	kanikoExecutorImage = "gcr.io/kaniko-project/executor:v0.12.0"
	// Default mode of files mounted from ConfiMap volumes.
	defaultFileMode int32 = 420
)

// Task inputs and outputs
const (
	imageInputName      = "imageName"
	dockerfileInputName = "dockerfileConfigmapName"
	sourceInputName     = "sourceConfigmapName"
)

// GetBuildTaskRunSpec generates a TaskRun spec from a RuntimeInfo.
func GetBuildTaskRunSpec(rnInfo *RuntimeInfo, fn *serverlessv1alpha1.Function, imageName, buildTaskName string) *tektonv1alpha1.TaskRunSpec {
	ref := &tektonv1alpha1.TaskRef{
		Name: buildTaskName,
		Kind: tektonv1alpha1.NamespacedTaskKind,
	}

	timeout, err := time.ParseDuration(buildTimeout)
	if err != nil {
		timeout = defaultBuildTimeout
	}

	// find Dockerfile name for runtime
	var dockerfileName string
	for _, rt := range rnInfo.AvailableRuntimes {
		if fn.Spec.Runtime == rt.ID {
			dockerfileName = rt.DockerfileName
			break
		}
	}

	inputs := tektonv1alpha1.TaskRunInputs{
		Params: []tektonv1alpha1.Param{
			{
				Name: imageInputName,
				Value: tektonv1alpha1.ArrayOrString{
					Type:      tektonv1alpha1.ParamTypeString,
					StringVal: imageName,
				},
			},
			{
				Name: dockerfileInputName,
				Value: tektonv1alpha1.ArrayOrString{
					Type:      tektonv1alpha1.ParamTypeString,
					StringVal: dockerfileName,
				},
			},
			{
				Name: sourceInputName,
				Value: tektonv1alpha1.ArrayOrString{
					Type:      tektonv1alpha1.ParamTypeString,
					StringVal: fn.Name,
				},
			},
		},
	}

	return &tektonv1alpha1.TaskRunSpec{
		ServiceAccount: rnInfo.ServiceAccount,
		TaskRef:        ref,
		Timeout:        &metav1.Duration{Duration: timeout},
		Inputs:         inputs,
	}
}

// GetBuildTaskSpec generates a Task spec from a RuntimeInfo.
func GetBuildTaskSpec(rnInfo *RuntimeInfo) *tektonv1alpha1.TaskSpec {
	inputs := &tektonv1alpha1.Inputs{
		Params: []tektonv1alpha1.ParamSpec{
			{
				Name:        imageInputName,
				Description: "Name of the container image to build, including the repository",
				Type:        tektonv1alpha1.ParamTypeString,
			},
			{
				Name:        dockerfileInputName,
				Description: "Name of the ConfigMap that contains the Dockerfile",
				Type:        tektonv1alpha1.ParamTypeString,
			},
			{
				Name:        sourceInputName,
				Description: "Name of the ConfigMap that contains the Function source",
				Type:        tektonv1alpha1.ParamTypeString,
			},
		},
	}

	steps := []tektonv1alpha1.Step{
		{Container: corev1.Container{
			Name:  "build-and-push",
			Image: kanikoExecutorImage,
			Args: []string{
				fmt.Sprintf("--destination=$(inputs.params.%s)", imageInputName),
			},
			Env: []corev1.EnvVar{{
				// Environment variable read by Kaniko to locate the container
				// registry credentials.
				// The Tekton credentials initializer sources container registry
				// credentials from the Secrets referenced in TaskRun's
				// ServiceAccounts, and makes them available in this directory.
				// https://github.com/tektoncd/pipeline/blob/master/docs/auth.md
				Name:  "DOCKER_CONFIG",
				Value: "/builder/home/.docker/",
			}},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "source",
					MountPath: "/src",
				},
				{
					Name:      fmt.Sprintf("$(inputs.params.%s)", dockerfileInputName),
					MountPath: "/workspace",
				},
			},
		}},
	}

	// populate Volume list with runtimes Dockerfiles
	vols := make([]corev1.Volume, len(rnInfo.AvailableRuntimes)+1)
	for i, rt := range rnInfo.AvailableRuntimes {
		vols[i] = corev1.Volume{
			Name: rt.DockerfileName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: proto.Int32(defaultFileMode),
					LocalObjectReference: corev1.LocalObjectReference{
						Name: rt.DockerfileName,
					},
				},
			},
		}
	}

	vols[len(rnInfo.AvailableRuntimes)] = corev1.Volume{
		Name: "source",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				DefaultMode: proto.Int32(defaultFileMode),
				LocalObjectReference: corev1.LocalObjectReference{
					Name: fmt.Sprintf("$(inputs.params.%s)", sourceInputName),
				},
			},
		},
	}

	return &tektonv1alpha1.TaskSpec{
		Inputs:  inputs,
		Steps:   steps,
		Volumes: vols,
	}
}
