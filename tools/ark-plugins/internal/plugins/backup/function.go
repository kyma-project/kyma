package backup

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/heptio/ark/pkg/apis/ark/v1"
	"github.com/heptio/ark/pkg/backup"

	"github.com/sirupsen/logrus"
)

// FunctionPlugin is a plugin for ark to backup several fields before creating restored object
type FunctionPlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo return list of resource kinds which should be handled by this plugin
func (p *FunctionPlugin) AppliesTo() (backup.ResourceSelector, error) {
	return backup.ResourceSelector{IncludedResources: []string{"all", "serviceinstance", "servicebinding", "servicebindingusage", "function", "subscription", "api", "eventactivation"},
		ExcludedNamespaces: []string{"default", "heptio-ark", "istio-system", "kube-public", "kube-system", "kyma-installer", "kyma-integration", "kyma-system"},
		LabelSelector:      "function",
	}, nil
}

// Execute sets a custom annotation on the item being backed up.
func (p *FunctionPlugin) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []backup.ResourceIdentifier, error) {
	p.Log.Println("Execute plugin backup function")

	return item, nil, nil
}
