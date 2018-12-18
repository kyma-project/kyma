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
	siLister  scListers.ServiceInstanceLister
	cscLister scListers.ClusterServiceClassLister
	scLister  scListers.ServiceClassLister
}

// NewBindingLabelsFetcher returnd BindingLabelsFetcher
func NewBindingLabelsFetcher(siLister scListers.ServiceInstanceLister,
	cscLister scListers.ClusterServiceClassLister, scLister scListers.ServiceClassLister) *BindingLabelsFetcher {
	return &BindingLabelsFetcher{
		siLister:  siLister,
		cscLister: cscLister,
		scLister:  scLister,
	}
}

// Fetch returns binding labels defined in ClusterServiceClass/ServiceClass
func (c *BindingLabelsFetcher) Fetch(svcBinding *scTypes.ServiceBinding) (map[string]string, error) {
	if svcBinding == nil {
		return nil, errors.New("cannot fetch labels from ClusterServiceClass/ServiceClass because binding is nil")
	}

	svcInstanceName := svcBinding.Spec.ServiceInstanceRef.Name
	svcInstance, err := c.siLister.ServiceInstances(svcBinding.Namespace).Get(svcInstanceName)

	if err != nil {
		return nil, errors.Wrapf(err, "while fetching ServiceInstance [%s] from namespace [%s] indicated by ServiceBinding", svcInstanceName, svcBinding.Namespace)
	}

	svcClusterServiceClassName := c.getClusterServiceClassNameFromInstance(svcInstance)
	svcServiceClassName := c.getServiceClassNameFromInstance(svcInstance)

	switch {
	case svcClusterServiceClassName != "" && svcServiceClassName != "":
		return nil, fmt.Errorf("unable to get class details because the ServiceInstance %s refers to ClusterServiceClass %s and ServiceClass %s",
			pretty.ServiceInstanceName(svcInstance), svcClusterServiceClassName, svcServiceClassName)
	case svcClusterServiceClassName != "":
		svcClass, err := c.cscLister.Get(svcClusterServiceClassName)
		if err != nil {
			return nil, errors.Wrapf(err, "while fetching ClusterServiceClass [%s]", svcClusterServiceClassName)
		}

		labels, err := c.getBindingLabelsFromClassSpec(&svcClass.Spec.CommonServiceClassSpec)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting labels from %s", pretty.ClusterServiceClassName(svcClass))
		}

		return labels, nil
	case svcServiceClassName != "":
		svcClass, err := c.scLister.ServiceClasses(svcBinding.Namespace).Get(svcServiceClassName)
		if err != nil {
			return nil, errors.Wrapf(err, "while fetching ServiceClass [%s/%s]", svcBinding.Namespace, svcServiceClassName)
		}

		labels, err := c.getBindingLabelsFromClassSpec(&svcClass.Spec.CommonServiceClassSpec)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting labels from %s", pretty.ServiceClassName(svcClass))
		}

		return labels, nil
	default:
		return nil, fmt.Errorf("cannot fetch ClusterServiceClass or ServiceClass from [%s]", pretty.ServiceInstanceName(svcInstance))
	}
}

func (c *BindingLabelsFetcher) getClusterServiceClassNameFromInstance(svcInstance *scTypes.ServiceInstance) string {
	if svcInstance.Spec.ClusterServiceClassExternalName != "" && svcInstance.Spec.ClusterServiceClassRef != nil {
		return svcInstance.Spec.ClusterServiceClassRef.Name
	}
	return svcInstance.Spec.ClusterServiceClassName
}

func (c *BindingLabelsFetcher) getServiceClassNameFromInstance(svcInstance *scTypes.ServiceInstance) string {
	if svcInstance.Spec.ServiceClassExternalName != "" && svcInstance.Spec.ServiceClassRef != nil {
		return svcInstance.Spec.ServiceClassRef.Name
	}
	return svcInstance.Spec.ServiceClassName
}

func (c *BindingLabelsFetcher) getBindingLabelsFromClassSpec(svcSpec *scTypes.CommonServiceClassSpec) (map[string]string, error) {
	var jsonDataAsAMap map[string]interface{}

	if svcSpec.ExternalMetadata == nil {
		return nil, nil
	}

	if err := json.Unmarshal(svcSpec.ExternalMetadata.Raw, &jsonDataAsAMap); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling raw metadata to json")
	}

	rawBindingLabels, exists := jsonDataAsAMap["bindingLabels"]
	if !exists {
		return nil, nil
	}

	bindingLabels, ok := rawBindingLabels.(map[string]interface{})
	if !ok {
		return nil, errors.New("cannot cast bindingLabels to map[string]interface{}")
	}

	out := make(map[string]string)
	for label, value := range bindingLabels {
		strValue, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("cannot cast bindingLabels[%s]=[%v] to string value", label, value)
		}
		out[label] = strValue
	}
	return out, nil
}
