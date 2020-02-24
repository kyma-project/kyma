package plugins

import (
	"fmt"

	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/sirupsen/logrus"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
)

const (
	natssChannelLabelKey   = "messaging.knative.dev/role"
	natssChannelLabelValue = "natss-channel"
)

// IgnoreNatssChannelService ignores Services associated to NATSS Channels during restore.
type IgnoreNatssChannelService struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a selector that determines what objects this plugin applies to.
func (p *IgnoreNatssChannelService) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"services"},
	}, nil
}

// Execute contains executes the plugin logic on the received object.
func (p *IgnoreNatssChannelService) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
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

	if meta.GetLabels()[natssChannelLabelKey] == natssChannelLabelValue {
		p.Log.Infof("Ignoring restore of %s", itemStr)
		return restoreOutput.WithoutRestore(), nil
	}

	p.Log.Infof("Restoring %s", itemStr)
	return restoreOutput, nil
}
