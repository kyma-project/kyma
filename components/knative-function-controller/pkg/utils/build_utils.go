package utils

import (
	"os"
	"time"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	runtimev1alpha1 "github.com/kyma-project/kyma/components/knative-function-controller/pkg/apis/runtime/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var buildTimeout = os.Getenv("BUILD_TIMEOUT")

var defaultMode = int32(420)

func GetBuildResource(rnInfo *RuntimeInfo, fn *runtimev1alpha1.Function, imageName string, buildName string) *buildv1alpha1.Build {

	args := []buildv1alpha1.ArgumentSpec{}
	args = append(args, buildv1alpha1.ArgumentSpec{Name: "IMAGE", Value: imageName})

	for _, rt := range rnInfo.AvailableRuntimes {
		if rt.ID == fn.Spec.Runtime {
			args = append(args, buildv1alpha1.ArgumentSpec{Name: "DOCKERFILE", Value: rt.DockerFileName})
		}
	}

	envs := []corev1.EnvVar{}

	timeout, err := time.ParseDuration(buildTimeout)
	if err != nil {
		timeout = 30 * time.Minute
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

func GetBuildTemplateSpec(fn *runtimev1alpha1.Function) buildv1alpha1.BuildTemplateSpec {

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

	volumes := []corev1.Volume{
		{
			Name: "dockerfile-nodejs-6",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &defaultMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "dockerfile-nodejs-6",
					},
				},
			},
		},
		{
			Name: "dockerfile-nodejs-8",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &defaultMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "dockerfile-nodejs-8",
					},
				},
			},
		},
	}

	bt := buildv1alpha1.BuildTemplateSpec{
		Parameters: parameters,
		Steps:      steps,
		Volumes:    volumes,
	}

	return bt
}
