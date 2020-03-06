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

	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"

	"github.com/gogo/protobuf/proto"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var buildTimeout = os.Getenv("BUILD_TIMEOUT")

const (
	// Timeout after which a TaskRun gets canceled. Can be overridden by the BUILD_TIMEOUT env
	// var.
	defaultBuildTimeout = 30 * time.Minute

	// Kaniko executor image used to build Function container images.
	// https://github.com/GoogleContainerTools/kaniko/blob/master/deploy/Dockerfile
	kanikoExecutorImage = "gcr.io/kaniko-project/executor:v0.17.1"

	// Standard volume names required during a Function build.
	sourceVolName     = "source"
	dockerfileVolName = "dockerfile"

	// https://github.com/tektoncd/pipeline/blob/v0.10.1/docs/auth.md#least-privilege
	tektonDockerVolume = "/tekton/home/.docker/"

	// Default mode of files mounted from ConfigMap volumes.
	defaultFileMode int32 = 420
)

func getResources(res map[corev1.ResourceName]string) (map[corev1.ResourceName]resource.Quantity, error) {
	memQ, err := resource.ParseQuantity(res[corev1.ResourceMemory])
	if err != nil {
		return nil, err
	}

	cpuQ, err := resource.ParseQuantity(res[corev1.ResourceCPU])
	if err != nil {
		return nil, err
	}

	return map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    cpuQ,
		corev1.ResourceMemory: memQ,
	}, nil
}

type ResourceConfig struct {
	Limits   map[corev1.ResourceName]string
	Requests map[corev1.ResourceName]string
}

// GetBuildTaskRunSpec generates a TaskRun spec from a RuntimeInfo.
func GetBuildTaskRunSpec(rnInfo *RuntimeInfo, fn *serverlessv1alpha1.Function, imageName string, resConf ResourceConfig) (*tektonv1alpha1.TaskRunSpec, error) {
	// find Dockerfile name for runtime
	var dockerfileName string
	for _, rt := range rnInfo.AvailableRuntimes {
		if fn.Spec.Runtime == rt.ID {
			dockerfileName = rt.DockerfileName
			break
		}
	}

	vols, volMounts := makeConfigMapVolumes(
		configmapVolumeSpec{
			name: sourceVolName,
			path: "/src",
			cmap: fn.Name,
		},
		configmapVolumeSpec{
			name: dockerfileVolName,
			path: "/workspace",
			cmap: dockerfileName,
		},
	)

	requests, err := getResources(resConf.Requests)
	if err != nil {
		return &tektonv1alpha1.TaskRunSpec{}, err
	}

	// tekton by default uses only limits
	// https://github.com/tektoncd/pipeline/blob/master/docs/taskruns.md#limitranges
	limits, err := getResources(resConf.Limits)
	if err != nil {
		return &tektonv1alpha1.TaskRunSpec{}, err
	}

	steps := []tektonv1alpha1.Step{
		{Container: corev1.Container{
			Name:  "build-and-push",
			Image: kanikoExecutorImage,
			Args: []string{
				fmt.Sprintf("--destination=%s", imageName), "--insecure", "--skip-tls-verify",
			},
			Env: []corev1.EnvVar{{
				// Environment variable read by Kaniko to locate the container
				// registry credentials.
				// The Tekton credentials initializer sources container registry
				// credentials from the Secrets referenced in TaskRun's
				// ServiceAccounts, and makes them available in this directory.
				// https://github.com/tektoncd/pipeline/blob/master/docs/auth.md
				// https://github.com/GoogleContainerTools/kaniko/blob/v0.17.1/deploy/Dockerfile#L45
				Name:  "DOCKER_CONFIG",
				Value: tektonDockerVolume,
			}},
			VolumeMounts: volMounts,
			Resources: corev1.ResourceRequirements{
				Requests: requests,
				Limits:   limits,
			},
		}},
	}

	timeout, err := time.ParseDuration(buildTimeout)
	if err != nil {
		timeout = defaultBuildTimeout
	}

	return &tektonv1alpha1.TaskRunSpec{
		ServiceAccountName: rnInfo.ServiceAccount,
		Timeout:            &metav1.Duration{Duration: timeout},
		TaskSpec: &tektonv1alpha1.TaskSpec{
			Steps:   steps,
			Volumes: vols,
		},
	}, nil
}

// configmapVolumeSpec is a succinct description of a ConfigMap Volume.
type configmapVolumeSpec struct {
	name string
	path string
	cmap string
}

// makeConfigMapVolumes returns a combination of Volumes and VolumeMounts for
// the given volume specs.
func makeConfigMapVolumes(vspecs ...configmapVolumeSpec) ([]corev1.Volume, []corev1.VolumeMount) {
	vols := make([]corev1.Volume, len(vspecs))
	vmounts := make([]corev1.VolumeMount, len(vspecs))

	for i, vspec := range vspecs {
		vols[i] = corev1.Volume{
			Name: vspec.name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: proto.Int32(defaultFileMode),
					LocalObjectReference: corev1.LocalObjectReference{
						Name: vspec.cmap,
					},
				},
			},
		}

		vmounts[i] = corev1.VolumeMount{
			Name:      vspec.name,
			MountPath: vspec.path,
		}
	}

	return vols, vmounts
}
