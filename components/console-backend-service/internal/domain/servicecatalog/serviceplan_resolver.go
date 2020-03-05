package servicecatalog

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type servicePlanResolver struct {
	sc     *servicePlanService
	rafter shared.RafterRetriever
}

func newServicePlanResolver(sc *servicePlanService, r shared.RafterRetriever) *servicePlanResolver {
	return &servicePlanResolver{sc, r}
}

func (r *servicePlanResolver) ServicePlanClusterAssetGroupField(ctx context.Context, obj *gqlschema.ServicePlan) (*gqlschema.ClusterAssetGroup, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("while getting clusterAssetGroup field obj is empty"))
		return nil, gqlerror.NewInternal()
	}

	assetGroup, err := r.rafter.ClusterAssetGroup().Find(obj.Name)
	if err != nil {
		glog.Errorf("Couldn't find clusterAssetGroup with name %s", obj.Name)
		return nil, nil
	}

	convertedAssetGroup, err := r.rafter.ClusterAssetGroupConverter().ToGQL(assetGroup)

	if err != nil {
		glog.Errorf("Couldn't convert clusterAssetGroup with name %s to GQL", obj.Name)
		return nil, nil
	}
	return convertedAssetGroup, nil
}

func (r *servicePlanResolver) ServicePlanAssetGroupField(ctx context.Context, obj *gqlschema.ServicePlan) (*gqlschema.AssetGroup, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("while getting assetGroup field obj is empty"))
		return nil, gqlerror.NewInternal()
	}

	assetGroup, err := r.rafter.AssetGroup().Find(obj.Name, obj.Namespace)
	if err != nil {
		glog.Errorf("Couldn't find assetGroup with name %s", obj.Name)
		return nil, nil
	}

	convertedAssetGroup, err := r.rafter.AssetGroupConverter().ToGQL(assetGroup)

	if err != nil {
		glog.Errorf("Couldn't convert assetGroup with name %s to GQL", obj.Name)
		return nil, nil
	}
	return convertedAssetGroup, nil
}
