package runtimes

import (
	"fmt"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
)

func BasicNodeJSFunction(msg string, rtm serverlessv1alpha1.Runtime) *function.FunctionData {
	return &function.FunctionData{
		Body:        fmt.Sprintf(`module.exports = { main: function(event, context) { return "%s" } }`, msg),
		Deps:        `{ "name": "hellobasic", "version": "0.0.1", "dependencies": {} }`,
		MaxReplicas: 2,
		MinReplicas: 1,
		Runtime:     rtm,
	}
}

func BasicNodeJSFunctionWithCustomDependency(msg string, rtm serverlessv1alpha1.Runtime) *function.FunctionData {
	return &function.FunctionData{
		Body:        fmt.Sprintf(`module.exports = { main: function(event, context) { return "%s" } }`, msg),
		Deps:        `{ "name": "hellobasic", "version": "0.0.1", "dependencies": { "@kyma/kyma-npm-test": "^1.0.0" } }`,
		MaxReplicas: 2,
		MinReplicas: 1,
		Runtime:     rtm,
	}
}

func NodeJSFunctionWithEnvFromConfigMapAndSecret(configMapName, cmEnvKey, secretName, secretEnvKey string, rtm serverlessv1alpha1.Runtime) *function.FunctionData {
	mappedCmEnvKey := "CM_KEY"
	mappedSecretEnvKey := "SECRET_KEY"

	return &function.FunctionData{
		Body:        fmt.Sprintf(`module.exports = { main: function(event, context) { return process.env["%s"] + "-" + process.env["%s"]; } }`, mappedCmEnvKey, mappedSecretEnvKey),
		Deps:        `{ "name": "hellowithconfigmapsecretenvs", "version": "0.0.1", "dependencies": { } }`,
		MaxReplicas: 1,
		MinReplicas: 1,
		Runtime:     rtm,
		Env: []corev1.EnvVar{
			{
				Name: mappedCmEnvKey,
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configMapName,
						},
						Key: cmEnvKey,
					},
				}},
			{
				Name: mappedSecretEnvKey,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: secretName,
						},
						Key: secretEnvKey,
					},
				}},
		},
	}
}
