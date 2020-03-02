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
	rafter shared.RafterRetriever
}

func newServicePlanResolver(r shared.RafterRetriever) *servicePlanResolver {
	return &servicePlanResolver{r}
}

func (r *servicePlanResolver) ServicePlanClusterAssetGroupField(ctx context.Context, obj *gqlschema.ServicePlan) (*gqlschema.ClusterAssetGroup, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("while getting assetGroup field obj is empty"))
		return nil, gqlerror.NewInternal()
	}
	return r.getAssetGroup(obj.Name), nil
}

func (r *servicePlanResolver) getAssetGroup(name string) *gqlschema.ClusterAssetGroup {
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
