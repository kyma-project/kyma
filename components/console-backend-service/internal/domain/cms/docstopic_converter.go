package cms

import (
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=gqlDocsTopicConverter -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=gqlDocsTopicConverter -case=underscore -output disabled -outpkg disabled
type gqlDocsTopicConverter interface {
	ToGQL(in *v1alpha1.DocsTopic) (*gqlschema.DocsTopic, error)
	ToGQLs(in []*v1alpha1.DocsTopic) ([]gqlschema.DocsTopic, error)
}

type docsTopicConverter struct {
	extractor extractor.DocsTopicStatusExtractor
}

func (c *docsTopicConverter) ToGQL(item *v1alpha1.DocsTopic) (*gqlschema.DocsTopic, error) {
	if item == nil {
		return nil, nil
	}

	status := c.extractor.Status(item.Status.CommonDocsTopicStatus)

	docsTopic := gqlschema.DocsTopic{
		Name:        item.Name,
		Namespace:   item.Namespace,
		Description: item.Spec.Description,
		DisplayName: item.Spec.DisplayName,
		GroupName:   item.Labels[GroupNameLabel],
		Status:      status,
	}

	return &docsTopic, nil
}

func (c *docsTopicConverter) ToGQLs(in []*v1alpha1.DocsTopic) ([]gqlschema.DocsTopic, error) {
	var result []gqlschema.DocsTopic
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
