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
	corev1 "k8s.io/api/core/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// GetServiceSpec gets ServiceSpec for a function
func GetServiceSpec(imageName string, imagePullSecretName string, rnInfo *RuntimeInfo) servingv1.ServiceSpec {
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

	configuration := servingv1.ConfigurationSpec{
		Template: servingv1.RevisionTemplateSpec{
			Spec: servingv1.RevisionSpec{
				PodSpec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: imageName,
						Env:   envVarsForRevision,
					}},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: imagePullSecretName},
					},
					ServiceAccountName: rnInfo.ServiceAccount,
				},
			},
		},
	}

	return servingv1.ServiceSpec{
		ConfigurationSpec: configuration,
	}
}
