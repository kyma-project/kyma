package plugins

import (
	"github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/restore"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// RemoveServiceInstanceFields is a plugin for velero to remove several fields before creating restored object
type RemoveServiceInstanceFields struct {
	Log logrus.FieldLogger
}

// AppliesTo return list of resource kinds which should be handled by this plugin
func (p *RemoveServiceInstanceFields) AppliesTo() (restore.ResourceSelector, error) {
	return restore.ResourceSelector{
		IncludedResources: []string{"serviceinstance"},
	}, nil
}

// Execute contains main logic for plugin
// nolint
func (p *RemoveServiceInstanceFields) Execute(item runtime.Unstructured, restore *v1.Restore) (runtime.Unstructured, error, error) {
	metadata, err := meta.Accessor(item)
	if err != nil {
		return nil, nil, err
	}

	p.Log.Infof("Removing serviceClassRef/servicePlanRef fields from instance %s in namespace %s", metadata.GetName(), metadata.GetNamespace())
	unstructured.RemoveNestedField(item.UnstructuredContent(), "spec", "serviceClassRef")
	unstructured.RemoveNestedField(item.UnstructuredContent(), "spec", "servicePlanRef")

	return item, nil, nil
}
