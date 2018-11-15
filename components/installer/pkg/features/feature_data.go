package features

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

const (
	namespace     = "kyma-installer"
	configMapName = "installer-feature-gates"
)

type FeatureData interface {
	IsEnabled(feature string) bool
}

type Reader interface {
	Read() (FeatureData, error)
}

func New(client *kubernetes.Clientset) (FeatureData, error) {
	reader := &reader{client: client}
	return reader.Read()
}

type reader struct {
	client *kubernetes.Clientset
}

func (r *reader) Read() (FeatureData, error) {
	configMap, err := r.client.CoreV1().ConfigMaps(namespace).Get(configMapName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if featuresString, ok := configMap.Data["features"]; !ok || featuresString == "" {
		return &featureData{features: make([]string, 0)}, nil
	} else {
		features := strings.Split(strings.Trim(featuresString, ", "), ",")
		return &featureData{features: features}, nil
	}

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
