package utils

import (
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1beta1"
	runtimev1alpha1 "github.com/kyma-project/kyma/components/knative-function-controller/pkg/apis/runtime/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// GetServiceSpec gets ServiceSpec for a function
func GetServiceSpec(imageName string, fn runtimev1alpha1.Function, rnInfo *RuntimeInfo) servingv1alpha1.ServiceSpec {

	// TODO: Make it constant for nodejs8/nodejs6
	envVarsForRevision := []corev1.EnvVar{
		{
			Name:  "FUNC_HANDLER",
			Value: "main",
		},
		{
			Name:  "MOD_NAME",
			Value: "handler",
		},
		{
			Name:  "FUNC_TIMEOUT",
			Value: "180",
		},
		{
			Name:  "FUNC_RUNTIME",
			Value: "nodejs8",
		},
		{
			Name:  "FUNC_MEMORY_LIMIT",
			Value: "128Mi",
		},
		{
			Name:  "FUNC_PORT",
			Value: "8080",
		},
		{
			Name:  "NODE_PATH",
			Value: "$(KUBELESS_INSTALL_VOLUME)/node_modules",
		},
	}

	configuration := servingv1alpha1.ConfigurationSpec{
		Template: &servingv1alpha1.RevisionTemplateSpec{
			Spec: servingv1alpha1.RevisionSpec{
				RevisionSpec: v1beta1.RevisionSpec{
					PodSpec: v1beta1.PodSpec{
						Containers: []corev1.Container{{
							Image: imageName,
							Env:   envVarsForRevision,
						}},
						ServiceAccountName: rnInfo.ServiceAccount,
					},
				},
			},
		},
	}

	return servingv1alpha1.ServiceSpec{
		ConfigurationSpec: configuration,
	}

}
