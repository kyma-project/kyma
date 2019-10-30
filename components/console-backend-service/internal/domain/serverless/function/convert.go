package function

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func toUnstructured(obj *v1alpha1.Function) (*unstructured.Unstructured, error) {
	return resource.ToUnstructured(obj)
}

func fromUnstructured(obj *unstructured.Unstructured) (*v1alpha1.Function, error) {
	function := &v1alpha1.Function{}
	err := resource.FromUnstructured(obj, function)
	return function, err
}

func ToGQL(item *v1alpha1.Function) *gqlschema.Function {
	if item == nil {
		return nil
	}

	function := gqlschema.Function{
		Name:         item.Name,
		Namespace:    item.Namespace,
		Labels:       item.Labels,
		Runtime:      item.Spec.Runtime,
		Size:         item.Spec.Size,
		Status:       getStatus(item.Status.Condition),
		Content:      item.Spec.Function,
		Dependencies: item.Spec.Deps,
	}

	return &function
}

func ToGQLs(in []*v1alpha1.Function) []gqlschema.Function {
	var result []gqlschema.Function
	for _, u := range in {
		converted := ToGQL(u)

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}

func getStatus(status v1alpha1.FunctionCondition) gqlschema.FunctionStatusType {
	switch status {
	case v1alpha1.FunctionConditionUnknown:
		return gqlschema.FunctionStatusTypeUnknown
	case v1alpha1.FunctionConditionRunning:
		return gqlschema.FunctionStatusTypeRunning
	case v1alpha1.FunctionConditionBuilding:
		return gqlschema.FunctionStatusTypeBuilding
	case v1alpha1.FunctionConditionError:
		return gqlschema.FunctionStatusTypeError
	case v1alpha1.FunctionConditionDeploying:
		return gqlschema.FunctionStatusTypeDeploying
	case v1alpha1.FunctionConditionUpdating:
		return gqlschema.FunctionStatusTypeUpdating
	default:
		return gqlschema.FunctionStatusTypeError
	}
}
