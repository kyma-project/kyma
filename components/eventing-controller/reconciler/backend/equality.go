package backend

import (
	"reflect"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	"k8s.io/apimachinery/pkg/conversion"
)

// Semantic can do semantic deep equality checks for API objects. Fields which
// are not relevant for the reconciliation logic are intentionally omitted.
var Semantic = conversion.EqualitiesOrDie(
	backendEqual,
)

// peerAuthenticationEqual asserts the equality of two PeerAuthentication objects.
func backendEqual(b1, b2 *eventingv1alpha1.EventingBackend) bool {
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
