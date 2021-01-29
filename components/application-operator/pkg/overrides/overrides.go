package overrides

import (
	. "strings"

	"github.com/kyma-project/kyma/components/application-operator/pkg/utils"
)

const (
	overridesKey   = "overrides"
	overridePrefix = "override."
)

type OverridesData struct {
	DomainName                            string `json:"domainName,omitempty"`
	ApplicationGatewayImage               string `json:"applicationGatewayImage,omitempty"`
	ApplicationGatewayTestsImage          string `json:"applicationGatewayTestsImage,omitempty"`
	EventServiceImage                     string `json:"eventServiceImage,omitempty"`
	EventServiceTestsImage                string `json:"eventServiceTestsImage,omitempty"`
	ApplicationConnectivityValidatorImage string `json:"applicationConnectivityValidatorImage,omitempty"`
	Tenant                                string `json:"tenant,omitempty"`
	Group                                 string `json:"group,omitempty"`
	GatewayOncePerNamespace               bool   `json:"deployGatewayOncePerNamespace,omitempty"`
	StrictMode                            string `json:"strictMode,omitempty"`
	IsBEBEnabled                          bool   `json:"isBEBEnabled,omitempty"`
}

func NewFlatOverridesMap(labels map[string]string) utils.StringMap {
	overridesMap := make(utils.StringMap)
	for key, value := range labels {
		if HasPrefix(key, overridePrefix) {
			overridesMap[removePrefix(key)] = value
		}
	}
	return overridesMap
}

func NewExtractedOverridesMap(config utils.InterfaceMap) utils.StringMap {
	previousOverrides := make(map[string]interface{})
	if val, ok := config[overridesKey]; ok {
		if casted, ok := val.(map[string]interface{}); ok {
			previousOverrides = casted
		}
	}
	return utils.NewStringMap(previousOverrides)
}

func MergeLabelOverrides(labels utils.StringMap, target map[string]interface{}) {
	for key, value := range labels {
		preKey := removePrefix(key)

		if preKey != "" {
			subMap := unwind(preKey, value)
			utils.MergeMaps(target, subMap)
			// add prop to override map for further reference
			utils.MergeMaps(target, map[string]interface{}{overridesKey: subMap})
		}
	}
}

func removePrefix(key string) string {
	preKey := TrimPrefix(key, overridePrefix)
	preKey = Trim(preKey, utils.Separator)
	return preKey
}

func unwind(key string, value string) map[string]interface{} {
	index := IndexRune(key, utils.RuneSeparator)

	if index == -1 {
		return map[string]interface{}{key: value}
	}

	currentKey := key[:index]
	subMap := unwind(key[index+1:], value)

	return map[string]interface{}{currentKey: subMap}
}
