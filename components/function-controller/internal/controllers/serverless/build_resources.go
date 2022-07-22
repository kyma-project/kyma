package serverless

import (
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	destinationArg        = "--destination"
	functionContainerName = "function"
	baseDir               = "/workspace/src/"
	workspaceMountPath    = "/workspace"
)

var (
	istioSidecarInjectFalse = map[string]string{
		"sidecar.istio.io/inject": "false",
	}
	svcTargetPort = intstr.FromInt(8080) // https://github.com/kubeless/runtimes/blob/master/stable/nodejs/kubeless.js#L28
)

func boolPtr(b bool) *bool {
	return &b
}

func buildRepoFetcherEnvVars(instance *serverlessv1alpha1.Function, gitOptions git.Options) []corev1.EnvVar {
	vars := []corev1.EnvVar{
		{
			Name:  "APP_REPOSITORY_URL",
			Value: gitOptions.URL,
		},
		{
			Name:  "APP_REPOSITORY_COMMIT",
			Value: instance.Status.Commit,
		},
		{
			Name:  "APP_MOUNT_PATH",
			Value: workspaceMountPath,
		},
	}

	if gitOptions.Auth != nil {
		vars = append(vars, corev1.EnvVar{
			Name:  "APP_REPOSITORY_AUTH_TYPE",
			Value: string(gitOptions.Auth.Type),
		})

		switch gitOptions.Auth.Type {
		case git.RepositoryAuthBasic:
			vars = append(vars, []corev1.EnvVar{
				{
					Name: "APP_REPOSITORY_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: gitOptions.Auth.SecretName,
							},
							Key: git.UsernameKey,
						},
					},
				},
				{
					Name: "APP_REPOSITORY_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: gitOptions.Auth.SecretName,
							},
							Key: git.PasswordKey,
						},
					},
				},
			}...)
		case git.RepositoryAuthSSHKey:
			vars = append(vars, corev1.EnvVar{
				Name: "APP_REPOSITORY_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: gitOptions.Auth.SecretName,
						},
						Key: git.KeyKey,
					},
				},
			})
			if _, ok := gitOptions.Auth.Credentials[git.PasswordKey]; ok {
				vars = append(vars, corev1.EnvVar{
					Name: "APP_REPOSITORY_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: gitOptions.Auth.SecretName,
							},
							Key: git.PasswordKey,
						},
					},
				})
			}
		}
	}

	return vars
}

func (r *FunctionReconciler) internalFunctionLabels(instance *serverlessv1alpha1.Function) map[string]string {
	labels := make(map[string]string, 3)

	labels[serverlessv1alpha1.FunctionNameLabel] = instance.Name
	labels[serverlessv1alpha1.FunctionManagedByLabel] = serverlessv1alpha1.FunctionControllerValue
	labels[serverlessv1alpha1.FunctionUUIDLabel] = string(instance.GetUID())

	return labels
}

func (r *FunctionReconciler) deploymentSelectorLabels(instance *serverlessv1alpha1.Function) map[string]string {
	return r.mergeLabels(map[string]string{serverlessv1alpha1.FunctionResourceLabel: serverlessv1alpha1.FunctionResourceLabelDeploymentValue}, r.internalFunctionLabels(instance))
}

func (r *FunctionReconciler) podLabels(instance *serverlessv1alpha1.Function) map[string]string {
	return r.mergeLabels(instance.Spec.Labels, r.deploymentSelectorLabels(instance))
}

func (r *FunctionReconciler) mergeLabels(labelsCollection ...map[string]string) map[string]string {
	result := make(map[string]string, 0)
	for _, labels := range labelsCollection {
		for key, value := range labels {
			result[key] = value
		}
	}
	return result
}
