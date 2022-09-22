package v1alpha1

import (
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this Subscription to the Hub version (v2).
func (s *Subscription) ConvertTo(rawObj conversion.Hub) error {
	_ = rawObj.(*v1alpha2.Subscription)

	return nil
}

func (s *Subscription) ConvertFrom(rawObj conversion.Hub) error {
	_ = rawObj.(*v1alpha2.Subscription)

	return nil
}
