package application

import (
	"strings"

	"github.com/kyma-project/kyma/components/application-operator/pkg/utils"
)

const (
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

type OverridesMap map[string]interface{}

type StringMap map[string]string

func MergeLabelOverrides(labels StringMap, target map[string]interface{}) {
	for key, value := range labels {
		if strings.HasPrefix(key, overridePrefix) {
			preParsedKey := strings.TrimPrefix(key, overridePrefix)
			preParsedKey = strings.TrimLeft(preParsedKey, ".")
			preParsedKey = strings.TrimRight(preParsedKey, ".")

			if len(preParsedKey) == 0 {
				continue
			}

			subMap := unwind(preParsedKey, value)
			utils.MergeMaps(target, subMap)
		}
	}
}

func unwind(key string, value string) map[string]interface{} {
	if len(key) == 0 {
		return map[string]interface{}{"": value}
	}

	beginIndex := 0
	endIndex := 0
	for i, ch := range key {
		if ch == '.' && i == beginIndex {
			beginIndex++ // ignore leading prefix chars
			endIndex++
		} else if ch != '.' {
			endIndex = i + 1
		} else if ch == '.' {
			break
		}
	}

	currentKey := key[beginIndex:endIndex]

	if endIndex < len(key) {
		subMap := unwind(key[endIndex:], value)
		return map[string]interface{}{currentKey: subMap}
	} else {
		return map[string]interface{}{currentKey: value}
	}
}
