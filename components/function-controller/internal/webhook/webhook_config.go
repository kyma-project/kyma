package webhook

import (
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/json"
	"os"
	"path/filepath"
)

type ResourcePreset map[string]struct {
	RequestCpu    string `yaml:"requestCpu"`
	RequestMemory string `yaml:"requestMemory"`
	LimitMemory   string `yaml:"limitMemory"`
	LimitCpu      string `yaml:"limitCpu"`
}

type RuntimePreset map[string]string

type FunctionResources struct {
	MinRequestCpu    string         `yaml:"minRequestCpu"`
	MinRequestMemory string         `yaml:"minRequestMemory"`
	DefaultPreset    string         `yaml:"defaultPreset"`
	Presets          ResourcePreset `yaml:"presets"`
	RuntimePresets   RuntimePreset  `yaml:"runtimePresets"`
}

type FunctionCfg struct {
	Resources FunctionResources `yaml:"resources"`
}

type BuildResources struct {
	MinRequestCpu    string         `yaml:"minRequestCpu"`
	MinRequestMemory string         `yaml:"minRequestMemory"`
	DefaultPreset    string         `yaml:"defaultPreset"`
	Presets          ResourcePreset `yaml:"presets"`
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
	cfg := WebhookConfig{
		DefaultRuntime: string(v1alpha2.NodeJs18),
		Function: FunctionCfg{
			Resources: FunctionResources{DefaultPreset: "M"}},
		BuildJob: BuildJob{Resources: BuildResources{DefaultPreset: "normal"}},
	}

	cleanPath := filepath.Clean(path)
	yamlFile, err := os.ReadFile(cleanPath)
	if err != nil {
		return WebhookConfig{}, err
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	return cfg, errors.Wrap(err, "while unmarshalling yaml")
}

func (r *ResourcePreset) UnmarshalYAML(unmarshal func(interface{}) error) error {
	rawPresets := ""
	err := unmarshal(&rawPresets)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(rawPresets), r); err != nil {
		return err
	}
	return nil
}

func (rp *RuntimePreset) UnmarshalYAML(unmarshal func(interface{}) error) error {
	rawPresets := ""
	err := unmarshal(&rawPresets)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(rawPresets), rp); err != nil {
		return err
	}
	return nil
}

func (wc WebhookConfig) ToValidationConfig() (v1alpha2.ValidationConfig, error) {
	cfg := v1alpha2.ValidationConfig{
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
	return cfg, nil
}

func (wc WebhookConfig) ToDefaultingConfig() (v1alpha2.DefaultingConfig, error) {
	cfg := v1alpha2.DefaultingConfig{
		Runtime: v1alpha2.Runtime(wc.DefaultRuntime),
		Function: v1alpha2.FunctionDefaulting{
			Resources: v1alpha2.FunctionResourcesDefaulting{
				DefaultPreset:  wc.Function.Resources.DefaultPreset,
				Presets:        wc.Function.Resources.Presets.toDefaultingResourcePreset(),
				RuntimePresets: wc.Function.Resources.RuntimePresets,
			},
		},
		BuildJob: v1alpha2.BuildJobDefaulting{
			Resources: v1alpha2.BuildJobResourcesDefaulting{
				DefaultPreset: wc.BuildJob.Resources.DefaultPreset,
				Presets:       wc.BuildJob.Resources.Presets.toDefaultingResourcePreset(),
			},
		},
	}
	return cfg, nil
}

func (rp ResourcePreset) toDefaultingResourcePreset() map[string]v1alpha2.ResourcesPreset {
	out := map[string]v1alpha2.ResourcesPreset{}
	for k, v := range rp {
		out[k] = v1alpha2.ResourcesPreset{
			RequestCPU:    v.RequestCpu,
			RequestMemory: v.RequestMemory,
			LimitCPU:      v.LimitCpu,
			LimitMemory:   v.LimitMemory,
		}
	}
	return out
}
