package servicecatalog_test

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
