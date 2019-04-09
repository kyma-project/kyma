package cms

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/runtime"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
)

func NewClusterDocsTopicResolver(clusterDocsTopicService clusterDocsTopicSvc, assetStoreRetriever shared.AssetStoreRetriever) *clusterDocsTopicResolver {
	return newClusterDocsTopicResolver(clusterDocsTopicService, assetStoreRetriever)
}

func (r *clusterDocsTopicResolver) SetDocsTopicConverter(converter gqlClusterDocsTopicConverter) {
	r.clusterDocsTopicConverter = converter
}

func NewClusterDocsTopicService(informer cache.SharedIndexInformer) (*clusterDocsTopicService, error) {
	return newClusterDocsTopicService(informer)
}

func NewDocsTopicResolver(docsTopicService docsTopicSvc, assetStoreRetriever shared.AssetStoreRetriever) *docsTopicResolver {
	return newDocsTopicResolver(docsTopicService, assetStoreRetriever)
}

func (r *docsTopicResolver) SetDocsTopicConverter(converter gqlDocsTopicConverter) {
	r.docsTopicConverter = converter
}

func NewDocsTopicService(informer cache.SharedIndexInformer) (*docsTopicService, error) {
	return newDocsTopicService(informer)
}

func (r *PluggableContainer) SetFakeClient() {
	r.cfg.dynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
}
