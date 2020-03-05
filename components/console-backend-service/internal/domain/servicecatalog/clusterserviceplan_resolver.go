package servicecatalog

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type clusterServicePlanResolver struct {
	sc     *clusterServicePlanService
	rafter shared.RafterRetriever
}

func newClusterServicePlanResolver(sc *clusterServicePlanService, r shared.RafterRetriever) *clusterServicePlanResolver {
	return &clusterServicePlanResolver{sc, r}
}

func (r *clusterServicePlanResolver) ClusterServicePlanClusterAssetGroupField(ctx context.Context, obj *gqlschema.ClusterServicePlan) (*gqlschema.ClusterAssetGroup, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("while getting clusterAssetGroup field obj is empty"))
		return nil, gqlerror.NewInternal()
	}

	clusterAssetGroup, err := r.rafter.ClusterAssetGroup().Find(obj.Name)
	if err != nil {
		glog.Errorf("Couldn't find clusterAssetGroup with name %s", obj.Name)
		return nil, nil
	}

	convertedAssetGroup, err := r.rafter.ClusterAssetGroupConverter().ToGQL(clusterAssetGroup)

	if err != nil {
		glog.Errorf("Couldn't convert clusterAssetGroup with name %s to GQL", obj.Name)
		return nil, nil
	}
	return convertedAssetGroup, nil
}
