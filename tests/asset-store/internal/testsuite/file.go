package testsuite

import (
	"github.com/kyma-project/kyma/tests/asset-store/pkg/upload"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

type fileUpload struct {
	url string
}

func newFileUpload(url string) *fileUpload {
	return &fileUpload{url: url}
}

// TODO: Test private bucket
func (u *fileUpload) Do() (*upload.Response, error) {
	paths := []string{
		"internal/testsuite/testdata/single/foo.yaml",
		"internal/testsuite/testdata/single/bar.yaml",
		"internal/testsuite/testdata/package/package.tar.gz",
	}

	publicFiles, err := u.loadFiles(paths)
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

func (u *fileUpload) loadFiles(paths []string) ([]*os.File, error){
	var files []*os.File

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, errors.Wrapf(err, "while constructing absolute path from %s", path)
		}

		file, err := os.Open(absPath)
		if err != nil {
			return nil, errors.Wrapf(err, "while loading fileUpload from path %s", path)
		}

		files = append(files, file)
	}

	return files, nil
}