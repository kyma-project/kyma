package servicecatalog_test

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog"
)

func TestServicePlanResolver_ServicePlanClusterAssetGroupField(t *testing.T) {
	t.Run("Returns nil if Rafter can't find the ClusterAssetGroup", func(t *testing.T) {
		servicePlan := givenServicePlan()
		rafter, cagGetter, _ := givenRafterForClusterAssetGroup()
		resolver := givenServicePlanResolver(rafter)

		cagGetter.On("Find", servicePlan.Name).Return(nil, errors.New("ClusterAssetGroup not found"))

		result, err := resolver.ServicePlanClusterAssetGroupField(context.Background(), servicePlan)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Returns nil if Rafter can't conver the ClusterAssetGroup to GQL", func(t *testing.T) {
		servicePlan := givenServicePlan()
		rafter, cagGetter, cagConverter := givenRafterForClusterAssetGroup()
		resolver := givenServicePlanResolver(rafter)

		assetGroup := givenClusterAssetGroup(servicePlan.Name)

		cagGetter.On("Find", servicePlan.Name).Return(assetGroup, nil)
		cagConverter.On("ToGQL", assetGroup).Return(nil, errors.New("Can't convert it to GQL"))

		result, err := resolver.ServicePlanClusterAssetGroupField(context.Background(), servicePlan)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Returns clusterAssetGroup when it should", func(t *testing.T) {
		servicePlan := givenServicePlan()
		rafter, cagGetter, cagConverter := givenRafterForClusterAssetGroup()
		resolver := givenServicePlanResolver(rafter)

		assetGroup := givenClusterAssetGroup(servicePlan.Name)
		assetGroupGQL := clusterAssetGroupToGQL(assetGroup)

		cagGetter.On("Find", servicePlan.Name).Return(assetGroup, nil)
		cagConverter.On("ToGQL", assetGroup).Return(assetGroupGQL, nil)

		result, err := resolver.ServicePlanClusterAssetGroupField(context.Background(), servicePlan)
		require.NoError(t, err)
		assert.Equal(t, servicePlan.Name, result.Name)
	})

}

func TestServicePlanResolver_ServicePlanAssetGroupField(t *testing.T) {
	t.Run("Returns assetGroup when it should", func(t *testing.T) {
		servicePlan := givenServicePlan()
		rafter, agGetter, agConverter := givenRafterForAssetGroup()
		resolver := givenServicePlanResolver(rafter)

		assetGroup := givenAssetGroup(servicePlan.Name)
		assetGroupGQL := assetGroupToGQL(assetGroup)

		agGetter.On("Find", servicePlan.Name, servicePlan.Namespace).Return(assetGroup, nil)
		agConverter.On("ToGQL", assetGroup).Return(assetGroupGQL, nil)

		result, err := resolver.ServicePlanAssetGroupField(context.Background(), servicePlan)
		require.NoError(t, err)
		assert.Equal(t, servicePlan.Name, result.Name)
	})
}

func givenServicePlan() *gqlschema.ServicePlan {
	return &gqlschema.ServicePlan{
		Name:      "testname",
		Namespace: "testnamespace",
	}
}

func givenServicePlanResolver(rafter shared.RafterRetriever) *servicecatalog.ServicePlanResolver {
	return servicecatalog.NewServicePlanResolver(rafter)
}
