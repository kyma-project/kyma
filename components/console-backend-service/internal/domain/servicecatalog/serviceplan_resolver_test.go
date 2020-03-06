package servicecatalog_test

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared/automock"
)

//pobieramy serviceplan z clusterAssetGroup i sprawdZamy, czy ona tam jest
//pobieramy serviceplan z assetGroup i sprawdZamy, czy ona tam jest

// symulacja błedu raftera
// symulacja błędu konwersji rafter.assetGroup -> gqlschema.assetGroup

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

func givenRafterForClusterAssetGroup() (*automock.RafterRetriever, *automock.ClusterAssetGroupGetter, *automock.GqlClusterAssetGroupConverter) {
	rafter := &automock.RafterRetriever{}
	cagGetter := &automock.ClusterAssetGroupGetter{}
	cagConverter := &automock.GqlClusterAssetGroupConverter{}
	rafter.On("ClusterAssetGroup").Return(cagGetter)
	rafter.On("ClusterAssetGroupConverter").Return(cagConverter)
	return rafter, cagGetter, cagConverter
}

func givenRafterForAssetGroup() (*automock.RafterRetriever, *automock.AssetGroupGetter, *automock.GqlAssetGroupConverter) {
	rafter := &automock.RafterRetriever{}
	agGetter := &automock.AssetGroupGetter{}
	agConverter := &automock.GqlAssetGroupConverter{}
	rafter.On("AssetGroup").Return(agGetter)
	rafter.On("AssetGroupConverter").Return(agConverter)
	return rafter, agGetter, agConverter
}

func givenClusterAssetGroup(name string) *v1beta1.ClusterAssetGroup {
	return &v1beta1.ClusterAssetGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func givenAssetGroup(name string) *v1beta1.AssetGroup {
	return &v1beta1.AssetGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func clusterAssetGroupToGQL(cag *v1beta1.ClusterAssetGroup) *gqlschema.ClusterAssetGroup {
	return &gqlschema.ClusterAssetGroup{
		Name: cag.Name,
	}
}

func assetGroupToGQL(cag *v1beta1.AssetGroup) *gqlschema.AssetGroup {
	return &gqlschema.AssetGroup{
		Name: cag.Name,
	}
}
