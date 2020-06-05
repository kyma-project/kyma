package extractor

import (
	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type ClusterAddonsUnstructuredExtractor struct{}

func (ext ClusterAddonsUnstructuredExtractor) Do(obj interface{}) (*v1alpha1.ClusterAddonsConfiguration, error) {
	u, err := ext.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return ext.FromUnstructured(u)
}

func (ext ClusterAddonsUnstructuredExtractor) ToUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", pretty.ClusterAddonsConfiguration, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func (ext ClusterAddonsUnstructuredExtractor) FromUnstructured(obj *unstructured.Unstructured) (*v1alpha1.ClusterAddonsConfiguration, error) {
	if obj == nil {
		return nil, nil
	}
	var addon v1alpha1.ClusterAddonsConfiguration
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &addon)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.ClusterAddonsConfiguration, obj.Object)
	}

	return &addon, nil
}
