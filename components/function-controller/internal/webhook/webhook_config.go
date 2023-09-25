package webhook

import (
	"os"
	"path/filepath"

	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type FunctionResources struct {
	MinRequestCpu    string `yaml:"minRequestCpu"`
	MinRequestMemory string `yaml:"minRequestMemory"`
}

type FunctionCfg struct {
	Resources FunctionResources `yaml:"resources"`
}

type BuildResources struct {
	MinRequestCpu    string `yaml:"minRequestCpu"`
	MinRequestMemory string `yaml:"minRequestMemory"`
}

type BuildJob struct {
	Resources BuildResources `yaml:"resources"`
}

type WebhookConfig struct {
	DefaultRuntime string      `yaml:"defaultRuntime"`
	Function       FunctionCfg `yaml:"function"`
	BuildJob       BuildJob    `yaml:"buildJob"`
	ReservedEnvs   []string    `yaml:"reservedEnvs"`
}

func LoadWebhookCfg(path string) (WebhookConfig, error) {
	cfg := WebhookConfig{DefaultRuntime: string(v1alpha2.NodeJs18)}

	cleanPath := filepath.Clean(path)
	yamlFile, err := os.ReadFile(cleanPath)
	if err != nil {
		return WebhookConfig{}, err
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	return cfg, errors.Wrap(err, "while unmarshalling yaml")
}

func (wc WebhookConfig) ToValidationConfig() v1alpha2.ValidationConfig {
	return v1alpha2.ValidationConfig{
		ReservedEnvs: wc.ReservedEnvs,
		Function: v1alpha2.MinFunctionValues{
			Resources: v1alpha2.MinFunctionResourcesValues{
				MinRequestCPU:    wc.Function.Resources.MinRequestCpu,
				MinRequestMemory: wc.Function.Resources.MinRequestMemory,
			},
		},
		BuildJob: v1alpha2.MinBuildJobValues{
			Resources: v1alpha2.MinBuildJobResourcesValues{
				MinRequestCPU:    wc.BuildJob.Resources.MinRequestCpu,
				MinRequestMemory: wc.BuildJob.Resources.MinRequestMemory,
			},
		},
	}
}
