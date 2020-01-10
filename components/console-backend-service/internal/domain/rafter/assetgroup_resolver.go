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

type assetGroupResolver struct {
	assetGroupService   assetGroupSvc
	assetGroupConverter gqlAssetGroupConverter
	assetService        assetSvc
	assetConverter      gqlAssetConverter
}

func newAssetGroupResolver(assetGroupService assetGroupSvc, assetGroupConverter gqlAssetGroupConverter, assetService assetSvc, assetConverter gqlAssetConverter) *assetGroupResolver {
	return &assetGroupResolver{
		assetGroupService:   assetGroupService,
		assetGroupConverter: assetGroupConverter,
		assetService:        assetService,
		assetConverter:      assetConverter,
	}
}

func (r *assetGroupResolver) AssetGroupAssetsField(ctx context.Context, obj *gqlschema.AssetGroup, types []string) ([]gqlschema.Asset, error) {
	if obj == nil {
		glog.Error(errors.Errorf("%s cannot be empty in order to resolve `assets` field", pretty.AssetGroup))
		return nil, gqlerror.NewInternal()
	}

	items, err := r.assetService.ListForAssetGroupByType(obj.Namespace, obj.Name, types)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", pretty.Assets, pretty.AssetGroup, obj.Name))
		return nil, gqlerror.New(err, pretty.Assets)
	}

	assets, err := r.assetConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.Assets))
		return nil, gqlerror.New(err, pretty.Assets)
	}

	return assets, nil
}

func (r *assetGroupResolver) AssetGroupEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.AssetGroupEvent, error) {
	channel := make(chan gqlschema.AssetGroupEvent, 1)
	filter := func(entity *v1beta1.AssetGroup) bool {
		return entity != nil && entity.Namespace == namespace
	}

	assetGroupServiceListener := listener.NewAssetGroup(channel, filter, r.assetGroupConverter)

	r.assetGroupService.Subscribe(assetGroupServiceListener)
	go func() {
		defer close(channel)
		defer r.assetGroupService.Unsubscribe(assetGroupServiceListener)
		<-ctx.Done()
	}()

	return channel, nil
}
