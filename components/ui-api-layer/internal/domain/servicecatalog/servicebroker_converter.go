package servicecatalog

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
)

type serviceBrokerConverter struct {
	extractor status.BrokerExtractor
}

func (c *serviceBrokerConverter) ToGQL(item *v1beta1.ServiceBroker) (*gqlschema.ServiceBroker, error) {
	if item == nil {
		return nil, nil
	}

	conditions := item.Status.Conditions
	returnStatus := c.extractor.Status(conditions)

	labels := new(gqlschema.JSON)
	err := labels.UnmarshalGQL(c.mapStringMapToJson(item.Labels))

	if err != nil {
		return nil, errors.Wrap(err, "While unmarshalling labels")
	}

	broker := gqlschema.ServiceBroker{
		Name:              item.Name,
		Environment:       item.Namespace,
		Status:            returnStatus,
		CreationTimestamp: item.CreationTimestamp.Time,
		Labels:            *labels,
		URL:               item.Spec.URL,
	}

	return &broker, nil
}

func (c *serviceBrokerConverter) ToGQLs(in []*v1beta1.ServiceBroker) ([]gqlschema.ServiceBroker, error) {
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

func (c *serviceBrokerConverter) mapStringMapToJson(labels map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range labels {
		result[k] = v
	}

	return result
}
