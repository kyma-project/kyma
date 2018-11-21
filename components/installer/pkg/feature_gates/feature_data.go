package feature_gates

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

const (
	namespace             = "kyma-installer"
	configMapName         = "installer-feature-gates"
	configMapPropertyName = "featureGates"
)

type Data interface {
	IsEnabled(feature string) bool
}

func New(client *kubernetes.Clientset) (Data, error) {
	configMap, err := client.CoreV1().ConfigMaps(namespace).Get(configMapName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	featuresString, ok := configMap.Data[configMapPropertyName]
	if !ok || featuresString == "" {
		return &featureData{features: make([]string, 0)}, nil
	}

	features := strings.Split(strings.Trim(featuresString, ", "), ",")
	return &featureData{features: features}, nil
}

type featureData struct {
	features []string
}

func (d *featureData) IsEnabled(feature string) bool {
	for _, enabled := range d.features {
		if enabled == feature {
			return true
		}
	}
	return false
}
