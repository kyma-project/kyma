package serverless

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/pretty"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type functionUnstructuredExtractor struct{}

func newFunctionUnstructuredExtractor() *functionUnstructuredExtractor {
	return &functionUnstructuredExtractor{}
}

func (ext *functionUnstructuredExtractor) do(obj interface{}) (*v1alpha1.Function, error) {
	u, err := ext.toUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return ext.fromUnstructured(u)
}

func (ext *functionUnstructuredExtractor) toUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", pretty.FunctionType, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func (ext *functionUnstructuredExtractor) fromUnstructured(obj *unstructured.Unstructured) (*v1alpha1.Function, error) {
	if obj == nil {
		return nil, nil
	}

	var function v1alpha1.Function
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &function)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.FunctionType, obj.Object)
	}

	return &function, nil
}
