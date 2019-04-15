package cms

import (
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=gqlClusterDocsTopicConverter -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=gqlClusterDocsTopicConverter -case=underscore -output disabled -outpkg disabled
type gqlClusterDocsTopicConverter interface {
	ToGQL(in *v1alpha1.ClusterDocsTopic) (*gqlschema.ClusterDocsTopic, error)
	ToGQLs(in []*v1alpha1.ClusterDocsTopic) ([]gqlschema.ClusterDocsTopic, error)
}

type clusterDocsTopicConverter struct {
	extractor extractor.DocsTopicStatusExtractor
}

func (c *clusterDocsTopicConverter) ToGQL(item *v1alpha1.ClusterDocsTopic) (*gqlschema.ClusterDocsTopic, error) {
	if item == nil {
		return nil, nil
	}

	status := c.extractor.Status(item.Status.CommonDocsTopicStatus)

	clusterDocsTopic := gqlschema.ClusterDocsTopic{
		Name:        item.Name,
		Description: item.Spec.Description,
		DisplayName: item.Spec.DisplayName,
		GroupName:   item.Labels[GroupNameLabel],
		Status:      status,
	}

	return &clusterDocsTopic, nil
}

func (c *clusterDocsTopicConverter) ToGQLs(in []*v1alpha1.ClusterDocsTopic) ([]gqlschema.ClusterDocsTopic, error) {
	var result []gqlschema.ClusterDocsTopic
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
