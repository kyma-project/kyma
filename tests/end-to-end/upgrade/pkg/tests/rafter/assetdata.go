package rafter

import (
	"fmt"

	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

const (
	assetPrefix   = "e2eupgrade"
	assetDataName = "petstore.json"
	assetDataURL  = "http://rafter-controller-manager.kyma-system.svc.cluster.local:8080/metrics"
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
