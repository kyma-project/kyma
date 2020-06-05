package rafter

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

type File struct {
	URL      string
	Metadata *runtime.RawExtension
}

//go:generate mockery -name=fileSvc -output=automock -outpkg=automock -case=underscore
type fileSvc interface {
	Extract(statusRef *v1beta1.AssetStatusRef) ([]*File, error)
	FilterByExtensionsAndExtract(statusRef *v1beta1.AssetStatusRef, filterExtensions []string) ([]*File, error)
}

type fileService struct{}

func newFileService() *fileService {
	return &fileService{}
}

func (svc *fileService) Extract(statusRef *v1beta1.AssetStatusRef) ([]*File, error) {
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

func (svc *fileService) FilterByExtensionsAndExtract(statusRef *v1beta1.AssetStatusRef, filterExtensions []string) ([]*File, error) {
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
