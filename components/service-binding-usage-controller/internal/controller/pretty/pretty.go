package pretty

import (
	"fmt"

	scTypes "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/pretty"
	sbuTypes "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	svcatSettings "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/settings/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ClusterServiceClassName returns string with type and name of ClusterServiceClass
func ClusterServiceClassName(obj *scTypes.ClusterServiceClass) string {
	return fmt.Sprintf("%s %s", pretty.ClusterServiceClass, obj.Name)
}

// ServiceClassName returns string with type and name of ServiceClass
func ServiceClassName(obj *scTypes.ServiceClass) string {
	return fmt.Sprintf("%s %s/%s", pretty.ServiceClass, obj.Namespace, obj.Name)
}

// ServiceInstanceName returns string with type, namespace and name of ServiceInstance
func ServiceInstanceName(obj *scTypes.ServiceInstance) string {
	return fmt.Sprintf("%s %s/%s", pretty.ServiceInstance, obj.Namespace, obj.Name)
}

// ServiceBindingName returns string with the type, namespace and name of ServiceBinding.
func ServiceBindingName(obj *scTypes.ServiceBinding) string {
	return fmt.Sprintf(`%s "%s/%s"`, pretty.ServiceBinding, obj.Namespace, obj.Name)
}

// ServiceBindingUsageName returns string with the type, namespace and name of ServiceBindingUsage.
func ServiceBindingUsageName(obj *sbuTypes.ServiceBindingUsage) string {
	return fmt.Sprintf(`ServiceBindingUsage "%s/%s"`, obj.Namespace, obj.Name)
}

// PodPresetName returns string with the type, namespace and name of PodPreset.
func PodPresetName(obj *svcatSettings.PodPreset) string {
	return fmt.Sprintf(`PodPreset "%s/%s"`, obj.Namespace, obj.Name)
}

// UnstructuredName returns string with the type, namespace and name of Unstructured object.
func UnstructuredName(obj *unstructured.Unstructured) string {
	return fmt.Sprintf(`Unstructured object "%s/%s"`, obj.GetName(), obj.GetNamespace())
}

// KeyItem returns string with queue item key, made from namespace and name
func KeyItem(namespace string, name string) string {
	return fmt.Sprintf(`"%s/%s"`, namespace, name)
}

// Key returns string item key, made from namespace and name
func Key(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
