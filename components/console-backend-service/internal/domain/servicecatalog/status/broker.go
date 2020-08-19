package status

import (
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type BrokerExtractor struct{}

func (e *BrokerExtractor) Status(conditions []v1beta1.ServiceBrokerCondition) *gqlschema.ServiceBrokerStatus {
	readyStatus, exists := e.findReadyCondition(conditions)
	if exists {
		return &gqlschema.ServiceBrokerStatus{
			Ready:   readyStatus.Status == v1beta1.ConditionTrue,
			Reason:  readyStatus.Reason,
			Message: readyStatus.Message,
		}
	}

	var reason, message string
	if len(conditions) > 0 {
		condition := conditions[0]
		reason = condition.Reason
		message = condition.Message
	}

	return &gqlschema.ServiceBrokerStatus{
		Ready:   false,
		Reason:  reason,
		Message: message,
	}
}

func (e *BrokerExtractor) findReadyCondition(conditions []v1beta1.ServiceBrokerCondition) (v1beta1.ServiceBrokerCondition, bool) {
	for _, condition := range conditions {
		if condition.Type == "Ready" {
			return condition, true
		}
	}

	return v1beta1.ServiceBrokerCondition{}, false
}
