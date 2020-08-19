package testsuite

import (
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

type assetData struct {
	Name string
	URL  string
	Mode v1beta1.AssetMode
	Type v1beta1.AssetGroupSourceType
}

type file struct {
	URL string
}
