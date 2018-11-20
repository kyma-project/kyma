package overrides

import (
	"github.com/kyma-project/kyma/components/installer/pkg/feature_gates"
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
	client      *kubernetes.Clientset
	featureData feature_gates.Data
}

//Input overrides data (from ConfigMaps/Secrets)
type inputMap map[string]string

type component struct {
	name      string
	overrides inputMap
}

// Returns overrides for components
// Returned slice may contain several components with the same name!
// Overrides labeled with 'feature' are added at the end of result slice
// in order of their appearance in feature gates list
func (r reader) readComponentOverrides() ([]component, error) {

	var components []component
	featureOverrides := map[string][]component{}

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
		if feature, ok := cMap.ObjectMeta.Labels["feature"]; !ok {
			components = append(components, comp)
		} else if r.featureData.IsEnabled(feature) {
			featureOverrides[feature] = append(featureOverrides[feature], comp)
		}

	}

	for _, sec := range secrets {
		comp := component{
			name:      sec.Labels["component"],
			overrides: toInputMap(sec.Data),
		}
		if f, ok := sec.ObjectMeta.Labels["feature"]; !ok {
			components = append(components, comp)
		} else if r.featureData.IsEnabled(f) {
			featureOverrides[f] = append(featureOverrides[f], comp)
		}
	}

	for _, f := range r.featureData.Enabled() {
		if _, ok := featureOverrides[f]; ok {
			components = append(components, featureOverrides[f]...)
		}
	}

	return components, nil
}

func (r reader) readCommonOverrides() ([]inputMap, error) {

	var res []inputMap
	featureOverrides := map[string][]inputMap{}

	configMaps, err := r.getLabeledConfigMaps(commonListOpts)
	if err != nil {
		return nil, err
	}

	secrets, err := r.getLabeledSecrets(commonListOpts)
	if err != nil {
		return nil, err
	}

	for _, cMap := range configMaps {
		if feature, ok := cMap.ObjectMeta.Labels["feature"]; !ok {
			res = append(res, cMap.Data)
		} else if r.featureData.IsEnabled(feature) {
			featureOverrides[feature] = append(featureOverrides[feature], cMap.Data)
		}
	}

	for _, sec := range secrets {
		if feature, ok := sec.ObjectMeta.Labels["feature"]; !ok {
			res = append(res, toInputMap(sec.Data))
		} else if r.featureData.IsEnabled(feature) {
			featureOverrides[feature] = append(featureOverrides[feature], toInputMap(sec.Data))
		}
	}

	for _, f := range r.featureData.Enabled() {
		if _, ok := featureOverrides[f]; ok {
			res = append(res, featureOverrides[f]...)
		}
	}

	return res, nil
}

func (r reader) getLabeledConfigMaps(opts metav1.ListOptions) ([]core.ConfigMap, error) {

	configMaps, err := r.client.CoreV1().ConfigMaps(namespace).List(opts)
	if err != nil {
		return nil, err
	}
	return configMaps.Items, nil
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
