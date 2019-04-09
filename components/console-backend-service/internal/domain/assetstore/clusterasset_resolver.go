package assetstore

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

type clusterAssetResolver struct {
	clusterAssetSvc       clusterAssetSvc
	clusterAssetConverter gqlClusterAssetConverter
	fileSvc               fileSvc
	fileConverter         gqlFileConverter
}

func newClusterAssetResolver(clusterAssetService clusterAssetSvc) *clusterAssetResolver {
	return &clusterAssetResolver{
		clusterAssetSvc:       clusterAssetService,
		clusterAssetConverter: &clusterAssetConverter{},
		fileSvc:               &fileService{},
		fileConverter:         &fileConverter{},
	}
}

func (r *clusterAssetResolver) ClusterAssetFilesField(ctx context.Context, obj *gqlschema.ClusterAsset, filterExtensions []string) ([]gqlschema.File, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve `files` field"), pretty.ClusterAsset)
		return nil, gqlerror.NewInternal()
	}

	asset, err := r.clusterAssetSvc.Find(obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", pretty.ClusterAsset, pretty.ClusterAsset, obj.Name))
		return nil, gqlerror.New(err, pretty.ClusterAsset)
	}

	if asset == nil {
		return nil, nil
	}

	var items []*File
	if len(filterExtensions) == 0 {
		items, err = r.fileSvc.Extract(&asset.Status.AssetRef)
	} else {
		items, err = r.fileSvc.FilterByExtensionsAndExtract(&asset.Status.AssetRef, filterExtensions)
	}
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", pretty.Files, pretty.ClusterAsset, obj.Name))
		return nil, gqlerror.New(err, pretty.ClusterAsset)
	}

	files, err := r.fileConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.Files))
		return nil, gqlerror.New(err, pretty.ClusterAsset)
	}

	return files, nil
}

func (r *clusterAssetResolver) ClusterAssetEventSubscription(ctx context.Context) (<-chan gqlschema.ClusterAssetEvent, error) {
	channel := make(chan gqlschema.ClusterAssetEvent, 1)
	filter := func(entity *v1alpha2.ClusterAsset) bool {
		return entity != nil
	}

	clusterAssetListener := listener.NewClusterAsset(channel, filter, r.clusterAssetConverter)

	r.clusterAssetSvc.Subscribe(clusterAssetListener)
	go func() {
		defer close(channel)
		defer r.clusterAssetSvc.Unsubscribe(clusterAssetListener)
		<-ctx.Done()
	}()

	return channel, nil
}
