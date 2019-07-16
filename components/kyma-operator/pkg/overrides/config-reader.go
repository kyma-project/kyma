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
	client kubernetes.Interface
}

//Input overrides data (from ConfigMaps/Secrets)
type inputMap map[string]string

type component struct {
	name      string
	overrides inputMap
}

// Returns overrides for components
// Returned slice may contain several components with the same name!
func (r reader) readComponentOverrides() ([]component, error) {

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
			overrides: toInputMap(sec.Data),
		}
		components = append(components, comp)
	}

	return components, nil
}

func (r reader) readCommonOverrides() ([]inputMap, error) {

	res := []inputMap{}

	configmaps, err := r.getLabeledConfigMaps(commonListOpts)
	if err != nil {
		return nil, err
	}

	secrets, err := r.getLabeledSecrets(commonListOpts)
	if err != nil {
		return nil, err
	}

	for _, cMap := range configmaps {
		res = append(res, cMap.Data)
	}

	for _, sec := range secrets {
		res = append(res, toInputMap(sec.Data))
	}

	return res, nil
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

func toInputMap(input map[string][]byte) inputMap {
	var output = make(inputMap)
	for key, value := range input {
		output[key] = string(value)
	}

	return output
}
