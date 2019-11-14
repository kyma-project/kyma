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

type assetResolver struct {
	assetSvc       assetSvc
	assetConverter gqlAssetConverter
	fileSvc        fileSvc
	fileConverter  gqlFileConverter
}

func newAssetResolver(assetService assetSvc, assetConverter gqlAssetConverter, fileService fileSvc, fileConverter gqlFileConverter) *assetResolver {
	return &assetResolver{
		assetSvc:       assetService,
		assetConverter: assetConverter,
		fileSvc:        fileService,
		fileConverter:  fileConverter,
	}
}

func (r *assetResolver) AssetFilesField(ctx context.Context, obj *gqlschema.RafterAsset, filterExtensions []string) ([]gqlschema.File, error) {
	if obj == nil {
		glog.Error(errors.Errorf("%s cannot be empty in order to resolve `files` field", pretty.Asset))
		return nil, gqlerror.NewInternal()
	}

	asset, err := r.assetSvc.Find(obj.Namespace, obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", pretty.Asset, pretty.Asset, obj.Name))
		return nil, gqlerror.New(err, pretty.Asset)
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
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", pretty.Files, pretty.Asset, obj.Name))
		return nil, gqlerror.New(err, pretty.Asset)
	}

	files, err := r.fileConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.Files))
		return nil, gqlerror.New(err, pretty.Asset)
	}

	return files, nil
}

func (r *assetResolver) AssetEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.RafterAssetEvent, error) {
	channel := make(chan gqlschema.RafterAssetEvent, 1)
	filter := func(entity *v1beta1.Asset) bool {
		return entity != nil && entity.Namespace == namespace
	}

	assetListener := listener.NewAsset(channel, filter, r.assetConverter)

	r.assetSvc.Subscribe(assetListener)
	go func() {
		defer close(channel)
		defer r.assetSvc.Unsubscribe(assetListener)
		<-ctx.Done()
	}()

	return channel, nil
}
