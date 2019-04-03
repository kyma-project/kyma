package testsuite

import (
	"fmt"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
)

const (
	AssetDataName = "petstore.json"
	AssetDataURL  = "https://petstore.swagger.io/v2/swagger.json"
)

type assetData struct {
	Name string
	URL  string
	Mode v1alpha2.AssetMode
}

func fixSimpleAssetData(prefix string) assetData {
	return assetData{
		Name: fmt.Sprintf("%s-%s", prefix, AssetDataName),
		URL:  AssetDataURL,
		Mode: v1alpha2.AssetSingle,
	}
}
