package testsuite

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/upload"
)

type assetData struct {
	Name string
	URL  string
	Mode v1alpha2.AssetMode
}

func convertToAssetResourceDetails(response *upload.Response, prefix string) []assetData {
	var assets []assetData
	for _, file := range response.UploadedFiles {
		var mode v1alpha2.AssetMode
		if strings.HasSuffix(file.FileName, ".tar.gz") || strings.HasSuffix(file.FileName, ".zip") {
			mode = v1alpha2.AssetPackage
		} else {
			mode = v1alpha2.AssetSingle
		}

		asset := assetData{
			Name: fmt.Sprintf("%s-%s", prefix, file.FileName),
			URL:  file.RemotePath,
			Mode: mode,
		}
		assets = append(assets, asset)
	}

	return assets
}
