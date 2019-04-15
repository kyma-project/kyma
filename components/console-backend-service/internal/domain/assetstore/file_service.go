package assetstore

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
)

type fileService struct{}

func (svc *fileService) Extract(statusRef *v1alpha2.AssetStatusRef) ([]*File, error) {
	if statusRef == nil {
		return nil, nil
	}

	var files []*File
	for _, file := range statusRef.Files {
		files = append(files, &File{
			URL:      fmt.Sprintf("%s/%s", statusRef.BaseURL, file.Name),
			Metadata: file.Metadata,
		})
	}
	return files, nil
}

func (svc *fileService) FilterByExtensionsAndExtract(statusRef *v1alpha2.AssetStatusRef, filterExtensions []string) ([]*File, error) {
	if statusRef == nil {
		return nil, nil
	}

	var files []*File
	for _, file := range statusRef.Files {
		for _, extension := range filterExtensions {
			var suffix string
			if strings.HasPrefix(extension, ".") {
				suffix = extension
			} else {
				suffix = fmt.Sprintf(".%s", extension)
			}

			if strings.HasSuffix(file.Name, suffix) {
				files = append(files, &File{
					URL:      fmt.Sprintf("%s/%s", statusRef.BaseURL, file.Name),
					Metadata: file.Metadata,
				})
			}
		}
	}
	return files, nil
}
