package servicecatalog

import (
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
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

	labels := item.Labels
	if labels == nil {
		labels = gqlschema.Labels{}
	}

	broker := gqlschema.ClusterServiceBroker{
		Name:              item.Name,
		Status:            returnStatus,
		CreationTimestamp: item.CreationTimestamp.Time,
		Labels:            labels,
		URL:               item.Spec.URL,
	}

	return &broker, nil
}

func (c *clusterServiceBrokerConverter) ToGQLs(in []*v1beta1.ClusterServiceBroker) ([]*gqlschema.ClusterServiceBroker, error) {
	var result []*gqlschema.ClusterServiceBroker
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
