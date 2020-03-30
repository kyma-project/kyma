package extractor

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
)

type TriggerUnstructuredExtractor struct{}

func (e *TriggerUnstructuredExtractor) Do(obj interface{}) (*v1alpha1.Trigger, error) {
	u, err := resource.ToUnstructured(obj)
	if err != nil || u == nil {
		return nil, err
	}

	trigger := &v1alpha1.Trigger{}
	err = resource.FromUnstructured(u, trigger)
	return trigger, err
}
