package backend

import (
	"reflect"

	v1 "k8s.io/api/core/v1"

	appsv1 "k8s.io/api/apps/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	"k8s.io/apimachinery/pkg/conversion"
)

// Semantic can do semantic deep equality checks for API objects. Fields which
// are not relevant for the reconciliation logic are intentionally omitted.
var Semantic = conversion.EqualitiesOrDie(
	eventingBackendEqual,
)

// eventingBackendEqual asserts the equality of two EventingBackend objects.
func eventingBackendEqual(b1, b2 *eventingv1alpha1.EventingBackend) bool {
	if b1 == b2 {
		return true
	}

	if b1 == nil || b2 == nil {
		return false
	}

	if !reflect.DeepEqual(b1.Labels, b2.Labels) {
		return false
	}

	if !reflect.DeepEqual(b1.Spec, b2.Spec) {
		return false
	}

	return true
}

// publisherProxyDeploymentEqual asserts the equality of two Deployment objects for event publisher proxy deployments.
func publisherProxyDeploymentEqual(b1, b2 *appsv1.Deployment) bool {
	if b1 == b2 {
		return true
	}

	if b1 == nil || b2 == nil {
		return false
	}

	if !reflect.DeepEqual(b1.Labels, b2.Labels) {
		return false
	}

	if !reflect.DeepEqual(b1.Spec, b2.Spec) {
		return false
	}

	return true
}

// secretEqual asserts the equality of two Secret objects for event publisher proxy deployments.
func secretEqual(b1, b2 *v1.Secret) bool {
	if b1 == b2 {
		return true
	}

	if b1 == nil || b2 == nil {
		return false
	}

	if !reflect.DeepEqual(b1.Labels, b2.Labels) {
		return false
	}

	if !reflect.DeepEqual(b1.Data, b2.Data) {
		return false
	}

	return true
}
