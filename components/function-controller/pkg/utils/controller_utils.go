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
	"fmt"

	"github.com/ghodss/yaml"

	corev1 "k8s.io/api/core/v1"
)

type Runtime interface {
	ReadConfigMap() error
}

type RuntimeInfo struct {
	RegistryInfo      string
	AvailableRuntimes []RuntimesSupported
	ServiceAccount    string
	Defaults          DefaultConfig
	FuncTypes         []FuncType
	FuncSizes         []FuncSize
}

type FuncType struct {
	Type string `json:"type"`
}

type FuncSize struct {
	Size string `json:"size"`
}

type DefaultConfig struct {
	Runtime         string `json:"runtime"`
	Size            string `json:"size"`
	TimeOut         int32  `json:"timeOut"`
	FuncContentType string `json:"funcContentType"`
}

type RuntimesSupported struct {
	ID             string `json:"ID"`
	DockerfileName string `json:"dockerfileName"`
}

func New(config *corev1.ConfigMap) (*RuntimeInfo, error) {
	mandatoryConfKeys := []string{
		"dockerRegistry",
		"serviceAccountName",
		"runtimes",
		"defaults",
		"funcTypes",
		"funcSizes",
	}

	if err := validateMandatoryConfig(config, mandatoryConfKeys); err != nil {
		return nil, err
	}

	var runtimes []RuntimesSupported
	if err := yaml.Unmarshal([]byte(config.Data["runtimes"]), &runtimes); err != nil {
		return nil, fmt.Errorf("unable to read supported runtimes: %s", err)
	}

	var defaultConfig DefaultConfig
	if err := yaml.Unmarshal([]byte(config.Data["defaults"]), &defaultConfig); err != nil {
		return nil, fmt.Errorf("unable to read default function config: %s", err)
	}

	var funcTypes []FuncType
	if err := yaml.Unmarshal([]byte(config.Data["funcTypes"]), &funcTypes); err != nil {
		return nil, fmt.Errorf("unable to read function types: %s", err)
	}

	var funcSizes []FuncSize
	if err := yaml.Unmarshal([]byte(config.Data["funcSizes"]), &funcSizes); err != nil {
		return nil, fmt.Errorf("unable to read function sizes: %s", err)
	}

	return &RuntimeInfo{
		RegistryInfo:      config.Data["dockerRegistry"],
		ServiceAccount:    config.Data["serviceAccountName"],
		AvailableRuntimes: runtimes,
		Defaults:          defaultConfig,
		FuncTypes:         funcTypes,
		FuncSizes:         funcSizes,
	}, nil
}

// validateMandatoryConfig asserts that all required values are defined in the
// given ConfigMap.
func validateMandatoryConfig(c *corev1.ConfigMap, mandatoryKeys []string) error {
	var missing []string
	for _, k := range mandatoryKeys {
		if c.Data[k] == "" {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing mandatory attributes in ConfigMap data %q: %q", c.Name, missing)
	}
	return nil
}
