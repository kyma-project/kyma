package controller

import (
	"encoding/json"
	"fmt"

	scTypes "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scListers "github.com/kubernetes-incubator/service-catalog/pkg/client/listers_generated/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/pretty"
	"github.com/pkg/errors"
)

// BindingLabelsFetcher extracts binding labels defined in ClusterServiceClass for given ServiceBinding
type BindingLabelsFetcher struct {
	siLister scListers.ServiceInstanceLister
	scLister scListers.ClusterServiceClassLister
}

// NewBindingLabelsFetcher returnd BindingLabelsFetcher
func NewBindingLabelsFetcher(siLister scListers.ServiceInstanceLister,
	scLister scListers.ClusterServiceClassLister) *BindingLabelsFetcher {
	return &BindingLabelsFetcher{
		siLister: siLister,
		scLister: scLister,
	}
}

// Fetch returns binding labels defined in ClusterServiceClass
func (c *BindingLabelsFetcher) Fetch(svcBinding *scTypes.ServiceBinding) (map[string]string, error) {
	if svcBinding == nil {
		return nil, errors.New("cannot fetch labels from ClusterServiceClass because binding is nil")
	}

	svcInstanceName := svcBinding.Spec.ServiceInstanceRef.Name
	svcInstance, err := c.siLister.ServiceInstances(svcBinding.Namespace).Get(svcInstanceName)

	if err != nil {
		return nil, errors.Wrapf(err, "while fetching ServiceInstance [%s] from namespace [%s] indicated by ServiceBinding", svcInstanceName, svcBinding.Namespace)
	}

	svcClassName := c.getClassNameFromInstance(svcInstance)
	if svcClassName == "" {
		return nil, fmt.Errorf("cannot fetch ClusterServiceClass from [%s]", pretty.ServiceInstanceName(svcInstance))
	}

	svcClass, err := c.scLister.Get(svcClassName)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching ClusterServiceClass [%s]", svcClassName)
	}

	return c.getBindingLabelsFromClass(svcClass)
}

func (c *BindingLabelsFetcher) getClassNameFromInstance(svcInstance *scTypes.ServiceInstance) string {
	if svcInstance.Spec.ClusterServiceClassExternalName != "" {
		return svcInstance.Spec.ClusterServiceClassRef.Name
	}
	return svcInstance.Spec.ClusterServiceClassName
}

func (c *BindingLabelsFetcher) getBindingLabelsFromClass(svcClass *scTypes.ClusterServiceClass) (map[string]string, error) {
	var jsonDataAsAMap map[string]interface{}

	if svcClass.Spec.ExternalMetadata == nil {
		return nil, nil
	}

	if err := json.Unmarshal(svcClass.Spec.ExternalMetadata.Raw, &jsonDataAsAMap); err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling raw metadata to json from [%s]", pretty.ClusterServiceClassName(svcClass))
	}

	rawBindingLabels, exists := jsonDataAsAMap["bindingLabels"]
	if !exists {
		return nil, nil
	}

	bindingLabels, ok := rawBindingLabels.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot cast bindingLabels to map[string]interface{} from [%s]", pretty.ClusterServiceClassName(svcClass))
	}

	out := make(map[string]string)
	for label, value := range bindingLabels {
		strValue, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("cannot cast bindingLabels[%s]=[%v] to string value from [%s] ", label, value, pretty.ClusterServiceClassName(svcClass))
		}
		out[label] = strValue
	}
	return out, nil
}
