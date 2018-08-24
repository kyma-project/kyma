package overrides

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	namespace     = "kyma-installer"
	labelSelector = "installer=overrides"
)

var listOpts = metav1.ListOptions{LabelSelector: labelSelector}

// ReaderInterface exposes functions
type ReaderInterface interface {
	GetCommonConfig() (map[string]string, error)
	GetComponentConfig(componentName string) (map[string]string, error)
}

type reader struct {
	client *kubernetes.Clientset
}

// NewReader returns a ready to use configmapClient
func NewReader(client *kubernetes.Clientset) (ReaderInterface, error) {
	r := &reader{
		client: client,
	}
	return r, nil
}

func (r reader) GetComponentConfig(componentName string) (map[string]string, error) {
	return nil, nil
}

func (r reader) GetCommonConfig() (map[string]string, error) {

	var combined = make(map[string]string)

	configmaps, err := r.getLabeledConfigMaps()
	if err != nil {
		return nil, err
	}

	secrets, err := r.getLabeledSecrets()
	if err != nil {
		return nil, err
	}

	for _, cMap := range configmaps {
		for key, val := range cMap.Data {
			combined[key] = val
		}
	}

	for _, sec := range secrets {
		for key, val := range sec.Data {
			combined[key] = string(val)
		}
	}

	return combined, nil
}

func (r reader) getLabeledConfigMaps() ([]core.ConfigMap, error) {

	configmaps, err := r.client.CoreV1().ConfigMaps(namespace).List(listOpts)
	if err != nil {
		return nil, err
	}
	return configmaps.Items, nil
}

func (r reader) getLabeledSecrets() ([]core.Secret, error) {

	secrets, err := r.client.CoreV1().Secrets(namespace).List(listOpts)
	if err != nil {
		return nil, err
	}
	return secrets.Items, nil
}
