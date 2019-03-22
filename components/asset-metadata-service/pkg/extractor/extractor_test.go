package extractor_test

import (
	"fmt"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/extractor"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/fileheader"
	fautomock "github.com/kyma-project/kyma/components/asset-metadata-service/pkg/fileheader/automock"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractor_ReadMetadata(t *testing.T) {
	testCases := []struct {
		Name                 string
		Path                 string
		ExpectedMetadata     map[string]interface{}
		ExpectedErrorMessage string
	}{
		{
			Name: "Markdown success",
			Path: "./testdata/success.md",
			ExpectedMetadata: map[string]interface{}{
				"title": "Access logs",
				"type":  "Details",
				"no":    3,
			},
		},
		{
			Name: "YAML success",
			Path: "./testdata/success.yaml",
			ExpectedMetadata: map[string]interface{}{
				"title":  "Hello world",
				"number": 9,
				"url":    "https://kyma-project.io",
			},
		},
		{
			Name:                 "Markdown error",
			Path:                 "./testdata/error.md",
			ExpectedErrorMessage: "while reading metadata from file fileName.md",
		},
		{
			Name:                 "YAML error",
			Path:                 "./testdata/error.yaml",
			ExpectedErrorMessage: "while reading metadata from file fileName.md",
		},
	}

	for tN, tC := range testCases {
		t.Run(fmt.Sprintf("%d: %s", tN, tC.Name), func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			f, err := openFile(tC.Path)
			g.Expect(err).NotTo(gomega.HaveOccurred())

			fHeader := &fautomock.FileHeader{}
			fHeader.On("Filename").Return("fileName.md")
			fHeader.On("Size").Return(int64(10)).Once()
			fHeader.On("Open").Return(f, nil).Once()

			m := extractor.New()
			metadata, err := m.ReadMetadata(fHeader)

			if tC.ExpectedErrorMessage != "" {
				g.Expect(err).To(gomega.HaveOccurred())
				g.Expect(err.Error()).Should(gomega.ContainSubstring(tC.ExpectedErrorMessage))
				return
			}

			g.Expect(err).NotTo(gomega.HaveOccurred())
			g.Expect(metadata).Should(gomega.Equal(tC.ExpectedMetadata))
		})
	}
}

func openFile(relativePath string) (fileheader.File, error) {
	absPath, err := filepath.Abs(relativePath)
	if err != nil {
		return nil, errors.Wrapf(err, "while constructing absolute path from %s", relativePath)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading file from path %s", absPath)
	}

	return file, nil
}
