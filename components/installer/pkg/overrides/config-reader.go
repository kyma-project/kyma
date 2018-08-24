package overrides

import (
	"strings"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	namespace              = "kyma-installer"
	overridesLabelSelector = "installer=overrides"
	commonLabelSelector    = "!component"
	componentLabelSelector = "component"
)

var commonListOpts = metav1.ListOptions{LabelSelector: concatLabels(overridesLabelSelector, commonLabelSelector)}
var componentListOpts = metav1.ListOptions{LabelSelector: concatLabels(overridesLabelSelector, componentLabelSelector)}

type reader struct {
	client *kubernetes.Clientset
}

type component struct {
	name      string
	overrides map[string]string
}

func (r reader) getComponents() ([]component, error) {

	var components = []component{}

	configmaps, err := r.getLabeledConfigMaps(componentListOpts)
	if err != nil {
		return nil, err
	}

	secrets, err := r.getLabeledSecrets(componentListOpts)
	if err != nil {
		return nil, err
	}

	for _, cMap := range configmaps {
		comp := component{
			name:      cMap.Labels["component"],
			overrides: cMap.Data,
		}
		components = append(components, comp)
	}

	for _, sec := range secrets {
		comp := component{
			name:      sec.Labels["component"],
			overrides: sec.StringData,
		}
		components = append(components, comp)
	}

	return components, nil
}

func (r reader) getCommonConfig() (map[string]string, error) {

	var combined = make(map[string]string)

	configmaps, err := r.getLabeledConfigMaps(commonListOpts)
	if err != nil {
		return nil, err
	}

	secrets, err := r.getLabeledSecrets(commonListOpts)
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

func (r reader) getLabeledConfigMaps(opts metav1.ListOptions) ([]core.ConfigMap, error) {

	configmaps, err := r.client.CoreV1().ConfigMaps(namespace).List(opts)
	if err != nil {
		return nil, err
	}
	return configmaps.Items, nil
}

func (r reader) getLabeledSecrets(opts metav1.ListOptions) ([]core.Secret, error) {

	secrets, err := r.client.CoreV1().Secrets(namespace).List(opts)
	if err != nil {
		return nil, err
	}
	return secrets.Items, nil
}

func concatLabels(labels ...string) string {
	return strings.Join(labels, ", ")
}
