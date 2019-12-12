package testsuite

import (
	"fmt"
	"os"
	"strings"

	"github.com/minio/minio-go"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/file"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/upload"
	"github.com/pkg/errors"
)

type MinioConfig struct {
	Endpoint  string `envconfig:"default=minio.kyma.local"`
	AccessKey string `envconfig:""`
	SecretKey string `envconfig:""`
	UseSSL    bool   `envconfig:"default=true"`
}

const basePath = "testdata"

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
		PublicFiles:  publicFiles,
		Directory:    "",
	}, u.url)
	if err != nil {
		return nil, errors.Wrapf(err, "while uploading files")
	}

	return resp, nil
}

func (u *testData) loadFiles(paths []string) ([]*os.File, error) {
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
		localPath("package.zip"),
	}
}

func localPath(fileName string) string {
	return fmt.Sprintf("%s/%s", basePath, fileName)
}

func verifyUploadedAssets(files []uploadedFile, logFn func(format string, args ...interface{})) error {
	for _, f := range files {
		path := localPath(f.AssetPath)
		logFn("Comparing file from URL %s and local file from path %s...", f.URL, f.AssetPath)
		equal, err := file.CompareLocalAndRemote(path, f.URL)
		if err != nil {
			return errors.Wrapf(err, "while comparing files %s and remote file from URL %s, defined in %s %s", path, f.URL, f.Owner.Kind, f.Owner.Name)
		}

		if !equal {
			return fmt.Errorf("files from %s and %s are not equal, defined in %s %s", path, f.URL, f.Owner.Kind, f.Owner.Name)
		}
	}

	return nil
}

func verifyDeletedAssets(files []uploadedFile, logFn func(format string, args ...interface{})) error {
	for _, f := range files {
		logFn("Checking if file %s exists...", f.URL)
		exists, err := file.Exists(f.URL)
		if err != nil {
			return errors.Wrapf(err, "while checking if remote file from URL %s exist", f.URL)
		}

		if exists {
			return fmt.Errorf("file %s defined in %s %s should not exist", f.URL, f.Owner.Kind, f.Owner.Name)
		}
	}

	return nil
}

func deleteFiles(minioCli *minio.Client, uploadResult *upload.Response, logFn func(format string, args ...interface{})) error {
	for _, res := range uploadResult.UploadedFiles {
		path := strings.SplitAfter(res.RemotePath, fmt.Sprintf("%s/", res.Bucket))[1]
		logFn("Deleting '%s' from bucket '%s'...", path, res.Bucket)
		err := minioCli.RemoveObject(res.Bucket, path)
		if err != nil {
			return errors.Wrapf(err, "while deleting file '%s' from bucket '%s'", path, res.Bucket)
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
	URL       string
	Owner     uploadedFileOwner
}

func uploadedFiles(ref v1alpha2.AssetStatusRef, ownerName, ownerKind string) []uploadedFile {
	var files []uploadedFile

	for _, file := range ref.Files {
		url := fmt.Sprintf("%s/%s", ref.BaseURL, file.Name)
		files = append(files, uploadedFile{
			AssetPath: file.Name,
			URL:       url,
			Owner: uploadedFileOwner{
				Name: ownerName,
				Kind: ownerKind,
			},
		})
	}

	return files
}
