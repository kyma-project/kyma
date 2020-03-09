package servicecatalog

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type ServicePlanResolver struct {
	rafter shared.RafterRetriever
}

func NewServicePlanResolver(r shared.RafterRetriever) *ServicePlanResolver {
	return &ServicePlanResolver{r}
}

func (r *ServicePlanResolver) ServicePlanClusterAssetGroupField(ctx context.Context, obj *gqlschema.ServicePlan) (*gqlschema.ClusterAssetGroup, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("while getting %s since Service Plan is empty", pretty.ClusterAssetGroup))
		return nil, gqlerror.NewInternal()
	}

	assetGroup, err := r.rafter.ClusterAssetGroup().Find(obj.Name)
	if err != nil {
		glog.Errorf("Couldn't find %s with name %s", pretty.ClusterAssetGroup, obj.Name)
		return nil, nil
	}

	convertedAssetGroup, err := r.rafter.ClusterAssetGroupConverter().ToGQL(assetGroup)

	if err != nil {
		glog.Errorf("Couldn't convert %s with name %s to GQL", pretty.ClusterAssetGroup, obj.Name)
		return nil, nil
	}
	return convertedAssetGroup, nil
}

func (r *ServicePlanResolver) ServicePlanAssetGroupField(ctx context.Context, obj *gqlschema.ServicePlan) (*gqlschema.AssetGroup, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("while getting %s since Service Plan is empty", pretty.AssetGroup))
		return nil, gqlerror.NewInternal()
	}

	assetGroup, err := r.rafter.AssetGroup().Find(obj.Name, obj.Namespace)
	if err != nil {
		glog.Errorf("Couldn't find %s with name %s", pretty.AssetGroup, obj.Name)
		return nil, nil
	}

	convertedAssetGroup, err := r.rafter.AssetGroupConverter().ToGQL(assetGroup)

	if err != nil {
		glog.Errorf("Couldn't convert %s with name %s to GQL", pretty.AssetGroup, obj.Name)
		return nil, nil
	}
	return convertedAssetGroup, nil
}
