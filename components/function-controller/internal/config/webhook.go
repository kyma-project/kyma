package config

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/json"
	"os"
	"path/filepath"
)

type ReplicasPreset map[string]struct {
	Min int `yaml:"min"`
	Max int `yaml:"max"`
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

type Resources struct {
	MinRequestCpu    string         `yaml:"minRequestCpu"`
	MinRequestMemory string         `yaml:"minRequestMemory"`
	DefaultPreset    string         `yaml:"defaultPreset"`
	Presets          ResourcePreset `yaml:"presets"`
}

type FunctionCfg struct {
	Replicas  Replicas  `yaml:"replicas"`
	Resources Resources `yaml:"resources"`
}

type WebhookConfig struct {
	Function FunctionCfg `yaml:"function"`
}

func LoadWebhookCfg(path string) (WebhookConfig, error) {
	//TODO: put default values in cfg
	cfg := WebhookConfig{}

	cleanPath := filepath.Clean(path)
	yamlFile, err := os.ReadFile(cleanPath)
	if err != nil {
		return WebhookConfig{}, err
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	return WebhookConfig{}, errors.Wrap(err, "while unmarshalling yaml")
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

func (r *ReplicasPreset) UnmarshalYAML(unmarshal func(interface{}) error) error {
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
