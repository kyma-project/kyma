package rafter

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
)

type clusterAssetGroupResolver struct {
	clusterAssetGroupService   clusterAssetGroupSvc
	clusterAssetGroupConverter gqlClusterAssetGroupConverter
	clusterAssetService        clusterAssetSvc
	clusterAssetConverter      gqlClusterAssetConverter
}

func newClusterAssetGroupResolver(clusterAssetGroupService clusterAssetGroupSvc, clusterAssetGroupConverter gqlClusterAssetGroupConverter, clusterAssetService clusterAssetSvc, clusterAssetConverter gqlClusterAssetConverter) *clusterAssetGroupResolver {
	return &clusterAssetGroupResolver{
		clusterAssetGroupService:   clusterAssetGroupService,
		clusterAssetGroupConverter: clusterAssetGroupConverter,
		clusterAssetService:        clusterAssetService,
		clusterAssetConverter:      clusterAssetConverter,
	}
}

func (r *clusterAssetGroupResolver) ClusterAssetGroupsQuery(ctx context.Context, viewContext *string, groupName *string) ([]gqlschema.ClusterAssetGroup, error) {
	items, err := r.clusterAssetGroupService.List(viewContext, groupName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.ClusterAssetGroups))
		return nil, gqlerror.New(err, pretty.ClusterAssetGroups)
	}

	clusterAssetGroups, err := r.clusterAssetGroupConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ClusterAssetGroups))
		return nil, gqlerror.New(err, pretty.ClusterAssetGroups)
	}

	return clusterAssetGroups, nil
}

func (r *clusterAssetGroupResolver) ClusterAssetGroupAssetsField(ctx context.Context, obj *gqlschema.ClusterAssetGroup, types []string) ([]gqlschema.RafterClusterAsset, error) {
	if obj == nil {
		glog.Error(errors.Errorf("%s cannot be empty in order to resolve `assets` field", pretty.ClusterAssetGroup))
		return nil, gqlerror.NewInternal()
	}

	items, err := r.clusterAssetService.ListForClusterAssetGroupByType(obj.Name, types)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", pretty.ClusterAssets, pretty.ClusterAssetGroup, obj.Name))
		return nil, gqlerror.New(err, pretty.ClusterAssets)
	}

	clusterAssets, err := r.clusterAssetConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ClusterAssets))
		return nil, gqlerror.New(err, pretty.ClusterAssets)
	}

	return clusterAssets, nil
}

func (r *clusterAssetGroupResolver) ClusterAssetGroupEventSubscription(ctx context.Context) (<-chan gqlschema.ClusterAssetGroupEvent, error) {
	channel := make(chan gqlschema.ClusterAssetGroupEvent, 1)
	filter := func(entity *v1beta1.ClusterAssetGroup) bool {
		return entity != nil
	}

	clusterAssetGroupListener := listener.NewClusterAssetGroup(channel, filter, r.clusterAssetGroupConverter)

	r.clusterAssetGroupService.Subscribe(clusterAssetGroupListener)
	go func() {
		defer close(channel)
		defer r.clusterAssetGroupService.Unsubscribe(clusterAssetGroupListener)
		<-ctx.Done()
	}()

	return channel, nil
}
