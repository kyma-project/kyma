package testsuite

import (
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/file"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/upload"
	"github.com/pkg/errors"
	"os"
)

const basePath = "internal/testsuite/testdata"

type testData struct {
	url string
}

func newTestData(url string) *testData {
	return &testData{url: url}
}

// TODO: Test private bucket
func (u *testData) Upload() (*upload.Response, error) {
	publicFiles, err := u.loadFiles(paths())
	if err != nil {
		return nil, errors.Wrapf(err, "while loading files to upload")
	}

	resp, err := upload.Do("", upload.UploadInput{
		PrivateFiles: nil,
		PublicFiles: publicFiles,
		Directory: "",
	}, u.url)
	if err != nil {
		return nil, errors.Wrapf(err, "while uploading files")
	}

	return resp, nil
}

func (u *testData) loadFiles(paths []string) ([]*os.File, error){
	var files []*os.File

	for _, path := range paths {
		file, err := file.Open(path)
		if err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	return files, nil
}

func paths() []string {
	return []string{
		localPath("foo.yaml"),
		localPath("bar.yaml"),
		localPath("package.tar.gz"),
	}
}

func localPath(fileName string) string {
	return fmt.Sprintf("%s/%s", basePath, fileName)
}

func verifyUploadedAsset(kind string, name string, ref v1alpha2.AssetStatusRef, shouldExist bool) error {
	for _, singleAsset := range ref.Assets {
		url := fmt.Sprintf("%s/%s", ref.BaseUrl, singleAsset)

		if !shouldExist {
			exists, err := file.Exists(url)
			if err != nil {
				return errors.Wrapf(err, "while checking if remote file from URL %s exist", url)
			}

			if exists {
				return fmt.Errorf("File %s defined in %s %s should not exist", url, kind, name)
			}
		} else {
			path := localPath(singleAsset)
			equal, err := file.CompareLocalAndRemote(path, url)
			if err != nil {
				return errors.Wrapf(err, "while comparing files %s and remote file from URL %s, defined in %s %s", path, url, kind, name)
			}

			if !equal {
				return fmt.Errorf("Files from %s and %s are not equal, defined in %s %s", path, url, kind, name)
			}
		}
	}

	return nil
}