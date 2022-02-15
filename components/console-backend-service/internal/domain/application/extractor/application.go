package extractor

import (
	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type ApplicationUnstructuredExtractor struct{}

func (ext ApplicationUnstructuredExtractor) Do(obj interface{}) (*v1alpha1.Application, error) {
	u, err := ext.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return ext.FromUnstructured(u)
}

func (ext ApplicationUnstructuredExtractor) ToUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", pretty.Application, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func (ext ApplicationUnstructuredExtractor) FromUnstructured(obj *unstructured.Unstructured) (*v1alpha1.Application, error) {
	if obj == nil {
		return nil, nil
	}

	var application v1alpha1.Application
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &application)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.Application, obj.Object)
	}

	return &application, nil
}

type ApplicationMappingUnstructuredExtractor struct{}

func (ext ApplicationMappingUnstructuredExtractor) Do(obj interface{}) (*mappingTypes.ApplicationMapping, error) {
	u, err := ext.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return ext.FromUnstructured(u)
}

func (ext ApplicationMappingUnstructuredExtractor) ToUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", pretty.ApplicationMapping, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func (ext ApplicationMappingUnstructuredExtractor) FromUnstructured(obj *unstructured.Unstructured) (*mappingTypes.ApplicationMapping, error) {
	if obj == nil {
		return nil, nil
	}

	var applicationMapping mappingTypes.ApplicationMapping
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &applicationMapping)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.ApplicationMapping, obj.Object)
	}

	return &applicationMapping, nil
}
