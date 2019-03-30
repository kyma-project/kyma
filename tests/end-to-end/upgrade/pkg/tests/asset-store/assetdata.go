package asset_store

import (
	"fmt"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
)

const (
	AssetPrefix 	 	 = "e2eupgrade"
	AssetDataName = "petstore.json"
	AssetDataURL  = "https://petstore.swagger.io/v2/swagger.json"
)

type assetData struct {
	name string
	url  string
	mode v1alpha2.AssetMode
}

func fixSimpleAssetData() assetData {
	return assetData{
		name: fmt.Sprintf("%s-%s", AssetPrefix, AssetDataName),
		url:  AssetDataURL,
		mode: v1alpha2.AssetSingle,
	}
}
