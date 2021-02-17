package object

import (
	"reflect"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"

	"k8s.io/apimachinery/pkg/conversion"
)

// Semantic can do semantic deep equality checks for API objects. Fields which
// are not relevant for the reconciliation logic are intentionally omitted.
var Semantic = conversion.EqualitiesOrDie(
	apiRuleEqual,
)

// channelEqual asserts the equality of two Channel objects.
func apiRuleEqual(a1, a2 *apigatewayv1alpha1.APIRule) bool {
	if a1 == a2 {
		return true
	}
	if a1 == nil || a2 == nil {
		return false
	}

	if !reflect.DeepEqual(a1.Labels, a2.Labels) {
		return false
	}

	if !reflect.DeepEqual(a1.OwnerReferences, a2.OwnerReferences) {
		return false
	}
	if !reflect.DeepEqual(a1.Spec.Service.Name, a2.Spec.Service.Name) {
		return false
	}
	if !reflect.DeepEqual(a1.Spec.Service.IsExternal, a2.Spec.Service.IsExternal) {
		return false
	}
	if !reflect.DeepEqual(a1.Spec.Service.Port, a2.Spec.Service.Port) {
		return false
	}
	if !reflect.DeepEqual(a1.Spec.Rules, a2.Spec.Rules) {
		return false
	}
	if !reflect.DeepEqual(a1.Spec.Gateway, a2.Spec.Gateway) {
		return false
	}

	return true
}
