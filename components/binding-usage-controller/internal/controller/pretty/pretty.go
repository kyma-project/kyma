package pretty

import (
	"fmt"

	scTypes "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/pretty"
	sbuTypes "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	settingsV1alpha1 "k8s.io/api/settings/v1alpha1"
)

// ClusterServiceClassName returns string with type and name of ClusterServiceClass
func ClusterServiceClassName(obj *scTypes.ClusterServiceClass) string {
	return fmt.Sprintf("%s %s", pretty.ClusterServiceClass, obj.Name)
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
func PodPresetName(obj *settingsV1alpha1.PodPreset) string {
	return fmt.Sprintf(`PodPreset "%s/%s"`, obj.Namespace, obj.Name)
}
