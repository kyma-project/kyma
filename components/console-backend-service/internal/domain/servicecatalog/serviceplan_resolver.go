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
	return r.getClusterAssetGroup(obj.Name), nil
}

func (r *servicePlanResolver) ServicePlanAssetGroupField(ctx context.Context, obj *gqlschema.ServicePlan) (*gqlschema.AssetGroup, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("while getting assetGroup field obj is empty"))
		return nil, gqlerror.NewInternal()
	}
	return r.getAssetGroup(obj.Name, obj.Namespace), nil
}

func (r *servicePlanResolver) getClusterAssetGroup(name string) *gqlschema.ClusterAssetGroup {
	assetGroup, err := r.rafter.ClusterAssetGroup().Find(name)
	if err != nil {
		glog.Errorf("Couldn't find clusterAssetGroup with name %s", name)
		return nil
	}

	convertedAssetGroup, err := r.rafter.ClusterAssetGroupConverter().ToGQL(assetGroup)

	if err != nil {
		glog.Errorf("Couldn't convert clusterAssetGroup with name %s to GQL", name)
		return nil
	}
	return convertedAssetGroup
}

func (r *servicePlanResolver) getAssetGroup(name string, namespace string) *gqlschema.AssetGroup {
	assetGroup, err := r.rafter.AssetGroup().Find(name, namespace)
	if err != nil {
		glog.Errorf("Couldn't find assetGroup with name %s", name)
		return nil
	}

	convertedAssetGroup, err := r.rafter.AssetGroupConverter().ToGQL(assetGroup)

	if err != nil {
		glog.Errorf("Couldn't convert assetGroup with name %s to GQL", name)
		return nil
	}
	return convertedAssetGroup
}
