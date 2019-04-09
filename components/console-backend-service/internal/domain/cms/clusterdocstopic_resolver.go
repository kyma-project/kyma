package cms

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	assetstorePretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/pkg/errors"
)

type clusterDocsTopicResolver struct {
	clusterDocsTopicSvc       clusterDocsTopicSvc
	assetStoreRetriever       shared.AssetStoreRetriever
	clusterDocsTopicConverter gqlClusterDocsTopicConverter
}

func newClusterDocsTopicResolver(clusterDocsTopicService clusterDocsTopicSvc, assetStoreRetriever shared.AssetStoreRetriever) *clusterDocsTopicResolver {
	return &clusterDocsTopicResolver{
		clusterDocsTopicSvc:       clusterDocsTopicService,
		assetStoreRetriever:       assetStoreRetriever,
		clusterDocsTopicConverter: &clusterDocsTopicConverter{},
	}
}

func (r *clusterDocsTopicResolver) ClusterDocsTopicsQuery(ctx context.Context, viewContext *string, groupName *string) ([]gqlschema.ClusterDocsTopic, error) {
	items, err := r.clusterDocsTopicSvc.List(viewContext, groupName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.ClusterDocsTopics))
		return nil, gqlerror.New(err, pretty.ClusterDocsTopics)
	}

	clusterDocsTopics, err := r.clusterDocsTopicConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ClusterDocsTopics))
		return nil, gqlerror.New(err, pretty.ClusterDocsTopics)
	}

	return clusterDocsTopics, nil
}

func (r *clusterDocsTopicResolver) ClusterDocsTopicAssetsField(ctx context.Context, obj *gqlschema.ClusterDocsTopic, types []string) ([]gqlschema.ClusterAsset, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve `assets` field"), pretty.ClusterDocsTopic)
		return nil, gqlerror.NewInternal()
	}

	items, err := r.assetStoreRetriever.ClusterAsset().ListForDocsTopicByType(obj.Name, types)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", assetstorePretty.ClusterAssets, pretty.ClusterDocsTopic, obj.Name))
		return nil, gqlerror.New(err, assetstorePretty.ClusterAssets)
	}

	clusterAssets, err := r.assetStoreRetriever.ClusterAssetConverter().ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", assetstorePretty.ClusterAssets))
		return nil, gqlerror.New(err, assetstorePretty.ClusterAssets)
	}

	return clusterAssets, nil
}

func (r *clusterDocsTopicResolver) ClusterDocsTopicEventSubscription(ctx context.Context) (<-chan gqlschema.ClusterDocsTopicEvent, error) {
	channel := make(chan gqlschema.ClusterDocsTopicEvent, 1)
	filter := func(entity *v1alpha1.ClusterDocsTopic) bool {
		return entity != nil
	}

	clusterDocsTopicListener := listener.NewClusterDocsTopic(channel, filter, r.clusterDocsTopicConverter)

	r.clusterDocsTopicSvc.Subscribe(clusterDocsTopicListener)
	go func() {
		defer close(channel)
		defer r.clusterDocsTopicSvc.Unsubscribe(clusterDocsTopicListener)
		<-ctx.Done()
	}()

	return channel, nil
}
