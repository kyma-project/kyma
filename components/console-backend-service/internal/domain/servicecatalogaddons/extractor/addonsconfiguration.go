package extractor

import (
	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type AddonsUnstructuredExtractor struct{}

func (ext AddonsUnstructuredExtractor) Do(obj interface{}) (*v1alpha1.AddonsConfiguration, error) {
	u, err := ext.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return ext.FromUnstructured(u)
}

func (ext AddonsUnstructuredExtractor) ToUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", pretty.AddonsConfiguration, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func (ext AddonsUnstructuredExtractor) FromUnstructured(obj *unstructured.Unstructured) (*v1alpha1.AddonsConfiguration, error) {
	if obj == nil {
		return nil, nil
	}
	var addon v1alpha1.AddonsConfiguration
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &addon)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.AddonsConfiguration, obj.Object)
	}

	return &addon, nil
}
