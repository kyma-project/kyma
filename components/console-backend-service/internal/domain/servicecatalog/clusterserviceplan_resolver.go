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
	rafter shared.RafterRetriever
}

func newClusterServicePlanResolver(r shared.RafterRetriever) *clusterServicePlanResolver {
	return &clusterServicePlanResolver{r}
}

func (r *clusterServicePlanResolver) ClusterServicePlanClusterAssetGroupField(ctx context.Context, obj *gqlschema.ClusterServicePlan) (*gqlschema.ClusterAssetGroup, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("while getting clusterAssetGroup field obj is empty"))
		return nil, gqlerror.NewInternal()
	}
	return r.getClusterAssetGroup(obj.Name), nil
}

func (r *clusterServicePlanResolver) getClusterAssetGroup(name string) *gqlschema.ClusterAssetGroup {
	clusterAssetGroup, err := r.rafter.ClusterAssetGroup().Find(name)
	if err != nil {
		glog.Errorf("Couldn't find clusterAssetGroup with name %s", name)
		return nil
	}

	convertedAssetGroup, err := r.rafter.ClusterAssetGroupConverter().ToGQL(clusterAssetGroup)

	if err != nil {
		glog.Errorf("Couldn't convert clusterAssetGroup with name %s to GQL", name)
		return nil
	}
	return convertedAssetGroup
}
