package extractor

import (
	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apicontroller/pretty"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type ApiUnstructuredExtractor struct{}

func (ext ApiUnstructuredExtractor) Do(obj interface{}) (*v1alpha2.Api, error) {
	u, err := ext.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return ext.FromUnstructured(u)
}

func (ext ApiUnstructuredExtractor) ToUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", pretty.API, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func (ext ApiUnstructuredExtractor) FromUnstructured(obj *unstructured.Unstructured) (*v1alpha2.Api, error) {
	if obj == nil {
		return nil, nil
	}
	var api v1alpha2.Api
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &api)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.API, obj.Object)
	}

	return &api, nil
}
