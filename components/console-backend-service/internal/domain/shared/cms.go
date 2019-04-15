package shared

import (
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=CmsRetriever -output=automock -outpkg=automock -case=underscore
type CmsRetriever interface {
	ClusterDocsTopic() ClusterDocsTopicGetter
	DocsTopic() DocsTopicGetter
	ClusterDocsTopicConverter() GqlClusterDocsTopicConverter
	DocsTopicConverter() GqlDocsTopicConverter
}

//go:generate mockery -name=ClusterDocsTopicGetter -output=automock -outpkg=automock -case=underscore
type ClusterDocsTopicGetter interface {
	Find(name string) (*v1alpha1.ClusterDocsTopic, error)
}

//go:generate mockery -name=DocsTopicGetter -output=automock -outpkg=automock -case=underscore
type DocsTopicGetter interface {
	Find(namespace, name string) (*v1alpha1.DocsTopic, error)
}

//go:generate mockery -name=GqlClusterDocsTopicConverter -output=automock -outpkg=automock -case=underscore
type GqlClusterDocsTopicConverter interface {
	ToGQL(item *v1alpha1.ClusterDocsTopic) (*gqlschema.ClusterDocsTopic, error)
	ToGQLs(in []*v1alpha1.ClusterDocsTopic) ([]gqlschema.ClusterDocsTopic, error)
}

//go:generate mockery -name=GqlDocsTopicConverter -output=automock -outpkg=automock -case=underscore
type GqlDocsTopicConverter interface {
	ToGQL(item *v1alpha1.DocsTopic) (*gqlschema.DocsTopic, error)
	ToGQLs(in []*v1alpha1.DocsTopic) ([]gqlschema.DocsTopic, error)
}
