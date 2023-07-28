package runtime

import (
	"fmt"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	corev1 "k8s.io/api/core/v1"
)

type Runtime interface {
	SanitizeDependencies(dependencies string) string
}

type Config struct {
	Runtime                 serverlessv1alpha2.Runtime
	DependencyFile          string
	FunctionFile            string
	DockerfileConfigMapName string
	RuntimeEnvs             []corev1.EnvVar
}

func GetRuntimeConfig(runtime serverlessv1alpha2.Runtime) Config {
	config := Config{
		Runtime:                 runtime,
		DockerfileConfigMapName: fmt.Sprintf("dockerfile-%s", runtime),
		RuntimeEnvs: []corev1.EnvVar{
			{Name: "FUNC_RUNTIME", Value: string(runtime)},
		},
	}
	fillConfigFileNames(runtime, &config)
	fillConfigEnvVars(runtime, &config)
	return config
}

func fillConfigEnvVars(runtime serverlessv1alpha2.Runtime, config *Config) {
	switch runtime {
	case serverlessv1alpha2.Python39:
		config.RuntimeEnvs = append(config.RuntimeEnvs,
			[]corev1.EnvVar{
				// https://github.com/kubeless/runtimes/blob/master/stable/python/python.jsonnet#L45
				{Name: "PYTHONPATH", Value: "$(KUBELESS_INSTALL_VOLUME)/lib.python3.9/site-packages:$(KUBELESS_INSTALL_VOLUME)"},
				{Name: "PYTHONUNBUFFERED", Value: "TRUE"}}...)
		return
	}
}

func fillConfigFileNames(runtime serverlessv1alpha2.Runtime, config *Config) {
	switch runtime {
	case serverlessv1alpha2.NodeJs16, serverlessv1alpha2.NodeJs18:
		config.DependencyFile = "package.json"
		config.FunctionFile = "handler.js"
		return
	case serverlessv1alpha2.Python39:
		config.DependencyFile = "requirements.txt"
		config.FunctionFile = "handler.py"
		return
	}
}

func GetRuntime(r serverlessv1alpha2.Runtime) Runtime {
	switch r {
	case serverlessv1alpha2.Python39:
		return python{}
	default:
		return nodejs{}
	}
}
