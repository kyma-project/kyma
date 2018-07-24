package servicecatalog

import (
	"sort"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
)

type brokerConverter struct{}

func (c *brokerConverter) ToGQL(item *v1beta1.ClusterServiceBroker) (*gqlschema.ServiceBroker, error) {
	if item == nil {
		return nil, nil
	}

	conditions := item.Status.Conditions
	c.sortConditions(conditions)
	returnStatus := c.conditionToBrokerStatus(conditions)

	labels := new(gqlschema.JSON)
	err := labels.UnmarshalGQL(c.mapStringMapToJson(item.Labels))

	if err != nil {
		return nil, errors.Wrap(err, "While unmarshalling labels")
	}

	broker := gqlschema.ServiceBroker{
		Name:              item.Name,
		Status:            returnStatus,
		CreationTimestamp: item.CreationTimestamp.Time,
		Labels:            *labels,
		Url:               item.Spec.URL,
	}

	return &broker, nil
}

func (c *brokerConverter) ToGQLs(in []*v1beta1.ClusterServiceBroker) ([]gqlschema.ServiceBroker, error) {
	var result []gqlschema.ServiceBroker
	for _, u := range in {
		converted, err := c.ToGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}

func (c *brokerConverter) sortConditions(conditions []v1beta1.ServiceBrokerCondition) {
	sort.SliceStable(conditions, func(i, j int) bool {
		return conditions[j].LastTransitionTime.Before(&conditions[i].LastTransitionTime)
	})
}

func (c *brokerConverter) conditionToBrokerStatus(conditions []v1beta1.ServiceBrokerCondition) gqlschema.ServiceBrokerStatus {
	readyStatus, exists := c.findReadyCondition(conditions)
	if exists {
		return gqlschema.ServiceBrokerStatus{
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

	return gqlschema.ServiceBrokerStatus{
		Ready:   false,
		Reason:  reason,
		Message: message,
	}
}

func (c *brokerConverter) findReadyCondition(conditions []v1beta1.ServiceBrokerCondition) (v1beta1.ServiceBrokerCondition, bool) {
	for _, condition := range conditions {
		if condition.Type == "Ready" {
			return condition, true
		}
	}

	return v1beta1.ServiceBrokerCondition{}, false
}

func (c *brokerConverter) mapStringMapToJson(labels map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range labels {
		result[k] = v
	}

	return result
}
