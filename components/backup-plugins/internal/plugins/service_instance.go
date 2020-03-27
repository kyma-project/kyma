package plugins

import (
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// RemoveServiceInstanceFields is a plugin for velero to remove several fields before creating restored object
type RemoveServiceInstanceFields struct {
	Log logrus.FieldLogger
}

// AppliesTo return list of resource kinds which should be handled by this plugin
func (p *RemoveServiceInstanceFields) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"serviceinstance"},
	}, nil
}

// Execute contains main logic for plugin
// nolint
func (p *RemoveServiceInstanceFields) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	metadata, err := meta.Accessor(input.Item)
	if err != nil {
		return nil, err
	}

	p.Log.Infof("Removing serviceClassRef/servicePlanRef fields from instance %s in namespace %s", metadata.GetName(), metadata.GetNamespace())
	unstructured.RemoveNestedField(input.Item.UnstructuredContent(), "spec", "serviceClassRef")
	unstructured.RemoveNestedField(input.Item.UnstructuredContent(), "spec", "servicePlanRef")

	return velero.NewRestoreItemActionExecuteOutput(input.Item), nil
}
