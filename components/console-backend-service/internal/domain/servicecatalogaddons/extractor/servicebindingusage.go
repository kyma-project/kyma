package extractor

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type BindingUsageUnstructuredExtractor struct{}

func (ext BindingUsageUnstructuredExtractor) Do(obj interface{}) (*v1alpha1.ServiceBindingUsage, error) {
	u, err := ext.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return ext.FromUnstructured(u)
}

func (ext BindingUsageUnstructuredExtractor) ToUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", pretty.ServiceBindingUsage, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func (ext BindingUsageUnstructuredExtractor) FromUnstructured(obj *unstructured.Unstructured) (*v1alpha1.ServiceBindingUsage, error) {
	if obj == nil {
		return nil, nil
	}
	var addon v1alpha1.ServiceBindingUsage
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &addon)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.ServiceBindingUsage, obj.Object)
	}

	return &addon, nil
}
