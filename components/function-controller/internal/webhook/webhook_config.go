package webhook

import (
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/json"
	"os"
	"path/filepath"
	"strconv"
)

type ReplicasPreset map[string]struct {
	Min int32 `yaml:"min"`
	Max int32 `yaml:"max"`
}

type Replicas struct {
	MinValue      string         `yaml:"minValue"`
	DefaultPreset string         `yaml:"defaultPreset"`
	Presets       ReplicasPreset `yaml:"presets"`
}

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
	Replicas  Replicas          `yaml:"replicas"`
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
	Function     FunctionCfg `yaml:"function"`
	BuildJob     BuildJob    `yaml:"buildJob"`
	ReservedEnvs []string    `yaml:"reservedEnvs"`
}

func LoadWebhookCfg(path string) (WebhookConfig, error) {
	cfg := WebhookConfig{
		Function: FunctionCfg{
			Replicas:  Replicas{DefaultPreset: "S"},
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

func (rp *ReplicasPreset) UnmarshalYAML(unmarshal func(interface{}) error) error {
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
	minReplicas, err := strconv.Atoi(wc.Function.Replicas.MinValue)
	if err != nil {
		return v1alpha2.ValidationConfig{}, nil
	}
	cfg := v1alpha2.ValidationConfig{
		ReservedEnvs: wc.ReservedEnvs,
		Function: v1alpha2.MinFunctionValues{
			Replicas: v1alpha2.MinFunctionReplicasValues{
				MinValue: int32(minReplicas),
			},
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
		Function: v1alpha2.FunctionDefaulting{
			Replicas: v1alpha2.FunctionReplicasDefaulting{
				DefaultPreset: wc.Function.Replicas.DefaultPreset,
				Presets:       wc.Function.Replicas.Presets.toDefaultingReplicaPreset(),
			},
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

func (rp ReplicasPreset) toDefaultingReplicaPreset() map[string]v1alpha2.ReplicasPreset {
	out := map[string]v1alpha2.ReplicasPreset{}
	for k, v := range rp {
		out[k] = v1alpha2.ReplicasPreset{
			Min: v.Min,
			Max: v.Max,
		}
	}
	return out
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
