package plugins

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
)

const (
	integrationNamespace = "kyma-integration"
	ignoredLabelPrefix   = "serving.knative.dev/"
)

// IgnoreKnative ignores Services associated with knative services
type IgnoreKnative struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a selector that determines what objects this plugin applies to.
func (p *IgnoreKnative) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedNamespaces: []string{integrationNamespace},
	}, nil
}

// Execute contains executes the plugin logic on the received object.
func (p *IgnoreKnative) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
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

	for label := range meta.GetLabels() {
		if strings.HasPrefix(label, ignoredLabelPrefix) {
			p.Log.Infof("Ignoring restore of %s", itemStr)
			return restoreOutput.WithoutRestore(), nil
		}
	}

	p.Log.Infof("Restoring %s", itemStr)
	return restoreOutput, nil
}
