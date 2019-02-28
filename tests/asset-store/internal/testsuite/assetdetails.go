package testsuite

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/upload"
	"strings"
)

type assetData struct {
	Name string
	URL string
	Mode v1alpha2.AssetMode
}

func convertToAssetResourceDetails(response *upload.Response) []assetData {
	var assets []assetData
	for _, file := range response.UploadedFiles {
		var mode v1alpha2.AssetMode
		if strings.HasSuffix(file.FileName, ".tar.gz") {
			mode = v1alpha2.AssetPackage
		} else {
			mode = v1alpha2.AssetSingle
		}

		//TODO: Remove workaround
		url := strings.Replace(file.RemotePath, "https://minio.kyma.local", "http://assetstore-minio.kyma-system.svc.cluster.local:9000", -1)

		asset := assetData{
			Name: file.FileName,
			URL: url,
			Mode:mode,
		}
		assets = append(assets, asset)
	}

	return assets
}
