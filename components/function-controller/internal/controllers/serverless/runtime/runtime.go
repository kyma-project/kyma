package runtime

import (
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type Runtime interface {
	SanitizeDependencies(dependencies string) string
}

type Config struct {
	Runtime                 serverlessv1alpha1.Runtime
	DependencyFile          string
	FunctionFile            string
	DockerfileConfigMapName string
	RuntimeEnvs             []corev1.EnvVar
}

func GetRuntimeConfig(r serverlessv1alpha1.Runtime) Config {
	switch r {
	case serverlessv1alpha1.Nodejs12:
		return Config{
			Runtime:                 serverlessv1alpha1.Nodejs12,
			DependencyFile:          "package.json",
			FunctionFile:            "handler.js",
			DockerfileConfigMapName: "dockerfile-nodejs-12",
			RuntimeEnvs: []corev1.EnvVar{{Name: "NODE_PATH", Value: "$(KUBELESS_INSTALL_VOLUME)/node_modules"},
				{Name: "FUNC_RUNTIME", Value: "nodejs12"},
			},
		}
	case serverlessv1alpha1.Python38:
		return Config{
			Runtime:                 serverlessv1alpha1.Python38,
			DependencyFile:          "requirements.txt",
			FunctionFile:            "handler.py",
			DockerfileConfigMapName: "dockerfile-python-38",
			RuntimeEnvs: []corev1.EnvVar{{Name: "PYTHONPATH", Value: "$(KUBELESS_INSTALL_VOLUME)/lib.python3.8/site-packages:$(KUBELESS_INSTALL_VOLUME)"}, // https://github.com/kubeless/runtimes/blob/master/stable/python/python.jsonnet#L45
				{Name: "FUNC_RUNTIME", Value: "python38"}},
		}
	default:
		return Config{
			Runtime:                 serverlessv1alpha1.Nodejs12,
			DependencyFile:          "package.json",
			FunctionFile:            "handler.js",
			DockerfileConfigMapName: "dockerfile-nodejs-12",
			RuntimeEnvs: []corev1.EnvVar{{Name: "NODE_PATH", Value: "$(KUBELESS_INSTALL_VOLUME)/node_modules"},
				{Name: "FUNC_RUNTIME", Value: "nodejs12"},
			},
		}
	}
}

func GetRuntime(r serverlessv1alpha1.Runtime) Runtime {
	switch r {
	case serverlessv1alpha1.Nodejs12:
		return nodejs{}
	case serverlessv1alpha1.Python38:
		return python{}
	default:
		return nodejs{}
	}
}
