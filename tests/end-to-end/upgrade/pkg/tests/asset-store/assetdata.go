package assetstore

import (
	"fmt"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
)

const (
	assetPrefix   = "e2eupgrade"
	assetDataName = "petstore.json"
	assetDataURL  = "https://petstore.swagger.io/v2/swagger.json"
)

type assetData struct {
	name string
	url  string
	mode v1alpha2.AssetMode
}

func fixSimpleAssetData() assetData {
	return assetData{
		name: fmt.Sprintf("%s-%s", assetPrefix, assetDataName),
		url:  assetDataURL,
		mode: v1alpha2.AssetSingle,
	}
}
