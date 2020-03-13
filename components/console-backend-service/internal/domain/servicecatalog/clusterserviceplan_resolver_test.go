package servicecatalog_test

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClusterServicePlanResolver_ServicePlanClusterAssetGroupField(t *testing.T) {
	t.Run("Returns nil if Rafter can't find the ClusterAssetGroup", func(t *testing.T) {
		servicePlan := givenClusterServicePlan()
		rafter, cagGetter, _ := givenRafterForClusterAssetGroup()
		resolver := givenClusterServicePlanResolver(rafter)

		cagGetter.On("Find", servicePlan.Name).Return(nil, errors.New("ClusterAssetGroup not found"))

		result, err := resolver.ClusterServicePlanClusterAssetGroupField(context.Background(), servicePlan)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Returns nil if Rafter can't conver the ClusterAssetGroup to GQL", func(t *testing.T) {
		servicePlan := givenClusterServicePlan()
		rafter, cagGetter, cagConverter := givenRafterForClusterAssetGroup()
		resolver := givenClusterServicePlanResolver(rafter)

		assetGroup := givenClusterAssetGroup(servicePlan.Name)

		cagGetter.On("Find", servicePlan.Name).Return(assetGroup, nil)
		cagConverter.On("ToGQL", assetGroup).Return(nil, errors.New("Can't convert it to GQL"))

		result, err := resolver.ClusterServicePlanClusterAssetGroupField(context.Background(), servicePlan)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Returns clusterAssetGroup when it should", func(t *testing.T) {
		servicePlan := givenClusterServicePlan()
		rafter, cagGetter, cagConverter := givenRafterForClusterAssetGroup()
		resolver := givenClusterServicePlanResolver(rafter)

		assetGroup := givenClusterAssetGroup(servicePlan.Name)
		assetGroupGQL := clusterAssetGroupToGQL(assetGroup)

		cagGetter.On("Find", servicePlan.Name).Return(assetGroup, nil)
		cagConverter.On("ToGQL", assetGroup).Return(assetGroupGQL, nil)

		result, err := resolver.ClusterServicePlanClusterAssetGroupField(context.Background(), servicePlan)
		require.NoError(t, err)
		assert.Equal(t, servicePlan.Name, result.Name)
	})

}
func givenClusterServicePlan() *gqlschema.ClusterServicePlan {
	return &gqlschema.ClusterServicePlan{
		Name: "testname",
	}
}

func givenClusterServicePlanResolver(rafter shared.RafterRetriever) *servicecatalog.ClusterServicePlanResolver {
	return servicecatalog.NewClusterServicePlanResolver(rafter)
}
