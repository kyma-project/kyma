package plugins

import (
	"fmt"

	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
)

const ()

// IgnoreByLabel ignores resources that carry a specific label
type IgnoreByLabel struct {
	Log logrus.FieldLogger
}

func ignoredLabels() labels.Set {
	return labels.Set{
		"eventing.knative.dev/broker": "default",
	}

}

// AppliesTo returns a selector that determines what objects this plugin applies to.
func (p *IgnoreByLabel) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		ExcludedNamespaces: []string{
			"kyma-system",
			"knative-eventing",
			"knative-serving",
			"istio-system",
		},
		ExcludedResources: []string{
			"trigger.eventing.knative.dev",
		},
	}, nil
}

// Execute contains executes the plugin logic on the received object.
func (p *IgnoreByLabel) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	item := input.Item

	meta, err := apimeta.Accessor(item)
	if err != nil {
		return nil, fmt.Errorf("accessing item metadata: %s", err)
	}

	itemStr := fmt.Sprintf("%s %s/%s",
		item.GetObjectKind().GroupVersionKind().GroupKind(),
		meta.GetNamespace(),
		meta.GetName(),
	)

	restoreOutput := velero.NewRestoreItemActionExecuteOutput(input.Item)
	for labelKey, labelValue := range meta.GetLabels() {
		if ignoredValue, isIgnored := ignoredLabels()[labelKey]; isIgnored && labelValue == ignoredValue {
			p.Log.Infof("Ignoring restore of %s as it had label %v:%v", itemStr, labelKey, labelValue)
			return restoreOutput.WithoutRestore(), nil
		}
	}

	p.Log.Infof("Restoring %s", itemStr)
	return restoreOutput, nil
}
