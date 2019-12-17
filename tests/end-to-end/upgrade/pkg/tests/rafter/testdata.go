package rafter

import (
	"fmt"

	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

const (
	assetPrefix   string = "e2eupgrade"
	assetDataName string = "petstore.json"
	assetDataURL  string = "https://petstore.swagger.io/v2/swagger.json"

	sourceType v1beta1.AssetGroupSourceType = "openapi"
	sourceName v1beta1.AssetGroupSourceName = "openapi"
	sourceURL  string                       = "https://petstore.swagger.io/v2/swagger.json"
)

type assetData struct {
	name string
	url  string
	mode v1beta1.AssetMode
}

func fixSimpleAssetData() assetData {
	return assetData{
		name: fmt.Sprintf("%s-%s", assetPrefix, assetDataName),
		url:  assetDataURL,
		mode: v1beta1.AssetSingle,
	}
}

func fixSimpleAssetGroupSpec() v1beta1.CommonAssetGroupSpec {
	return v1beta1.CommonAssetGroupSpec{
		DisplayName: "Asset Group Sample",
		Description: "Asset Group Description",
		Sources: []v1beta1.Source{
			{
				Name: sourceName,
				Type: sourceType,
				Mode: v1beta1.AssetGroupSingle,
				URL:  sourceURL,
			},
		},
	}
}
