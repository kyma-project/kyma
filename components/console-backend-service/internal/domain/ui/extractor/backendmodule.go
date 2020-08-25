package extractor

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/apis/ui/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type BMUnstructuredExtractor struct{}

func (ext BMUnstructuredExtractor) Do(obj interface{}) (*v1alpha1.BackendModule, error) {
	u, err := ext.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return ext.FromUnstructured(u)
}

func (ext BMUnstructuredExtractor) ToUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", pretty.BackendModule, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func (ext BMUnstructuredExtractor) FromUnstructured(obj *unstructured.Unstructured) (*v1alpha1.BackendModule, error) {
	if obj == nil {
		return nil, nil
	}
	var backendModule v1alpha1.BackendModule
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &backendModule)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.BackendModule, obj.Object)
	}

	return &backendModule, nil
}
