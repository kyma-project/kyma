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
		f, err := file.Open(path)
		if err != nil {
			return nil, err
		}

		files = append(files, f)
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

func verifyUploadedAsset(files []uploadedFile, shouldExist bool) error {
	for _, f := range files {
		if !shouldExist {
			exists, err := file.Exists(f.URL)
			if err != nil {
				return errors.Wrapf(err, "while checking if remote file from URL %s exist", f.URL)
			}

			if exists {
				return fmt.Errorf("File %s defined in %s %s should not exist", f.URL, f.Owner.Kind, f.Owner.Name)
			}
		} else {
			path := localPath(f.AssetPath)
			equal, err := file.CompareLocalAndRemote(path, f.URL)
			if err != nil {
				return errors.Wrapf(err, "while comparing files %s and remote file from URL %s, defined in %s %s", path, f.URL, f.Owner.Kind, f.Owner.Name)
			}

			if !equal {
				return fmt.Errorf("Files from %s and %s are not equal, defined in %s %s", path, f.URL, f.Owner.Kind, f.Owner.Name)
			}
		}
	}

	return nil
}

type uploadedFileOwner struct {
	Kind string
	Name string
}

type uploadedFile struct {
	AssetPath string
	URL string
	Owner uploadedFileOwner
}

func uploadedFiles(ref v1alpha2.AssetStatusRef, ownerName, ownerKind string) []uploadedFile {
	var files []uploadedFile

	for _, singleAsset := range ref.Assets {
		url := fmt.Sprintf("%s/%s", ref.BaseUrl, singleAsset)
		files = append(files, uploadedFile{
			AssetPath: singleAsset,
			URL: url,
			Owner: uploadedFileOwner{
				Name: ownerName,
				Kind: ownerKind,
			},
		})
	}

	return files
}