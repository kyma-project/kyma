package convert

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func FunctionToUnstructured(obj *v1alpha1.Function) (*unstructured.Unstructured, error) {
	return resource.ToUnstructured(obj)
}

func UnstructuredToFunction(obj *unstructured.Unstructured) (*v1alpha1.Function, error) {
	function := &v1alpha1.Function{}
	err := resource.FromUnstructured(obj, function)
	return function, err
}

func FunctionToGQL(item *v1alpha1.Function) (*gqlschema.Function, error) {
	if item == nil {
		return nil, nil
	}

	function := gqlschema.Function{
		Name:      item.Name,
		Namespace: item.Namespace,
		Labels:    item.Labels,
		Runtime:   item.Spec.Runtime,
		Size:      item.Spec.Size,
	}

	return &function, nil
}

func FunctionsToGQLs(in []*v1alpha1.Function) ([]gqlschema.Function, error) {
	var result []gqlschema.Function
	for _, u := range in {
		converted, err := FunctionToGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}
