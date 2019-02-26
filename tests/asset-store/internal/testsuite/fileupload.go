package testsuite

import (
	"github.com/kyma-project/kyma/tests/asset-store/pkg/upload"
	"github.com/pkg/errors"
	"log"
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
	publicFiles, err := u.loadFiles()
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


	log.Println(resp)

	return resp, nil
}

func (u *fileUpload) loadFiles() ([]*os.File, error){
	var files []*os.File

	paths := []string{
		"internal/testsuite/testdata/single/foo.yaml",
		"internal/testsuite/testdata/single/bar.yaml",
		"internal/testsuite/testdata/package/package.tar.gz",
	}

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, errors.Wrapf(err, "while constructing absolute path from %s", path)
		}

		file, err := os.Open(absPath)
		if err != nil {
			return nil, errors.Wrapf(err, "while loading file from path %s", path)
		}

		files = append(files, file)
	}

	return files, nil
}