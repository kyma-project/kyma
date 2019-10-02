
package convert

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/pretty"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)


func FunctionToUnstructured(obj *v1alpha1.Function) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %T to unstructured", pretty.Function, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func UnstructuredToFunction(obj *unstructured.Unstructured) (*v1alpha1.Function, error) {
	if obj == nil {
		return nil, nil
	}

	var function v1alpha1.Function
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &function)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.Function, obj.Object)
	}

	return &function, nil
}

