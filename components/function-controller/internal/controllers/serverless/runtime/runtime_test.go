package runtime_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func TestGetRuntimeConfig(t *testing.T) {
	for testName, testData := range map[string]struct {
		name    string
		runtime serverlessv1alpha1.Runtime
		want    runtime.Config
	}{
		"nodejs12": {
			name:    "nodejs12",
			runtime: serverlessv1alpha1.Nodejs12,
			want: runtime.Config{
				Runtime:                 serverlessv1alpha1.Nodejs12,
				DependencyFile:          "package.json",
				FunctionFile:            "handler.js",
				DockerfileConfigMapName: "dockerfile-nodejs-12",
				RuntimeEnvs: []corev1.EnvVar{{Name: "NODE_PATH", Value: "$(KUBELESS_INSTALL_VOLUME)/node_modules"},
					{Name: "FUNC_RUNTIME", Value: "nodejs12"}},
			},
		},
		"python38": {
			name:    "python38",
			runtime: serverlessv1alpha1.Python38,
			want: runtime.Config{
				Runtime:                 serverlessv1alpha1.Python38,
				DependencyFile:          "requirements.txt",
				FunctionFile:            "handler.py",
				DockerfileConfigMapName: "dockerfile-python-38",
				RuntimeEnvs: []corev1.EnvVar{{Name: "PYTHONPATH", Value: "$(KUBELESS_INSTALL_VOLUME)/lib.python3.8/site-packages:$(KUBELESS_INSTALL_VOLUME)"},
					{Name: "FUNC_RUNTIME", Value: "python38"}},
			},
		},
		"nodej14": {
			name:    "nodejs14 config",
			runtime: serverlessv1alpha1.Nodejs14,
			want: runtime.Config{
				Runtime:                 serverlessv1alpha1.Nodejs14,
				DependencyFile:          "package.json",
				FunctionFile:            "handler.js",
				DockerfileConfigMapName: "dockerfile-nodejs-14",
				RuntimeEnvs: []corev1.EnvVar{{Name: "NODE_PATH", Value: "$(KUBELESS_INSTALL_VOLUME)/node_modules"},
					{Name: "FUNC_RUNTIME", Value: "nodejs14"}},
			},
		},
		"default": {
			name:    "nodejs12",
			runtime: serverlessv1alpha1.Nodejs12,
			want: runtime.Config{
				Runtime:                 serverlessv1alpha1.Nodejs12,
				DependencyFile:          "package.json",
				FunctionFile:            "handler.js",
				DockerfileConfigMapName: "dockerfile-nodejs-12",
				RuntimeEnvs: []corev1.EnvVar{{Name: "NODE_PATH", Value: "$(KUBELESS_INSTALL_VOLUME)/node_modules"},
					{Name: "FUNC_RUNTIME", Value: "nodejs12"}},
			},
		}} {
		t.Run(testName, func(t *testing.T) {
			//given
			g := gomega.NewWithT(t)

			// when
			config := runtime.GetRuntimeConfig(testData.runtime)

			// then
			g.Expect(config).To(gomega.BeEquivalentTo(testData.want))
		})
	}
}
