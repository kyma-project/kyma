package extractor

import (
	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/pretty"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type MFUnstructuredExtractor struct{}

func (ext MFUnstructuredExtractor) Do(obj interface{}) (*v1alpha1.MicroFrontend, error) {
	u, err := ext.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return ext.FromUnstructured(u)
}

func (ext MFUnstructuredExtractor) ToUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", pretty.MicroFrontend, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func (ext MFUnstructuredExtractor) FromUnstructured(obj *unstructured.Unstructured) (*v1alpha1.MicroFrontend, error) {
	if obj == nil {
		return nil, nil
	}
	var microFrontend v1alpha1.MicroFrontend
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &microFrontend)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.MicroFrontend, obj.Object)
	}

	return &microFrontend, nil
}

type CMFUnstructuredExtractor struct{}

func (ext CMFUnstructuredExtractor) Do(obj interface{}) (*v1alpha1.ClusterMicroFrontend, error) {
	u, err := ext.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return ext.FromUnstructured(u)
}

func (ext CMFUnstructuredExtractor) ToUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", pretty.ClusterMicroFrontend, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func (ext CMFUnstructuredExtractor) FromUnstructured(obj *unstructured.Unstructured) (*v1alpha1.ClusterMicroFrontend, error) {
	if obj == nil {
		return nil, nil
	}
	var clusterMicroFrontend v1alpha1.ClusterMicroFrontend
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &clusterMicroFrontend)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.ClusterMicroFrontend, obj.Object)
	}

	return &clusterMicroFrontend, nil
}
