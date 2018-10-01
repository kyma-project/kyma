package servicecatalog

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
)

type clusterServiceBrokerConverter struct {
	extractor status.BrokerExtractor
}

func (c *clusterServiceBrokerConverter) ToGQL(item *v1beta1.ClusterServiceBroker) (*gqlschema.ClusterServiceBroker, error) {
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

	broker := gqlschema.ClusterServiceBroker{
		Name:              item.Name,
		Status:            returnStatus,
		CreationTimestamp: item.CreationTimestamp.Time,
		Labels:            *labels,
		URL:               item.Spec.URL,
	}

	return &broker, nil
}

func (c *clusterServiceBrokerConverter) ToGQLs(in []*v1beta1.ClusterServiceBroker) ([]gqlschema.ClusterServiceBroker, error) {
	var result []gqlschema.ClusterServiceBroker
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

func (c *clusterServiceBrokerConverter) mapStringMapToJson(labels map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range labels {
		result[k] = v
	}

	return result
}
