package servicecatalog

import (
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
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

	broker := gqlschema.ServiceBroker{
		Name:              item.Name,
		Namespace:         item.Namespace,
		Status:            returnStatus,
		CreationTimestamp: item.CreationTimestamp.Time,
		Labels:            item.Labels,
		URL:               item.Spec.URL,
	}

	return &broker, nil
}

func (c *serviceBrokerConverter) ToGQLs(in []*v1beta1.ServiceBroker) ([]*gqlschema.ServiceBroker, error) {
	var result []*gqlschema.ServiceBroker
	for _, u := range in {
		converted, err := c.ToGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, converted)
		}
	}
	return result, nil
}
