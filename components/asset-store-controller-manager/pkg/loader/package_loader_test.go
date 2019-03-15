package loader

import (
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestLoader_Load_Package(t *testing.T) {
	expected := []string{
		"structure/swagger.json",
		"structure/docs/README.md",
	}

	for testName, testCase := range map[string]struct {
		path string
	}{
		"ZipArchive": {
			path: "./testdata/structure.zip",
		},
		"TarGzArchive": {
			path: "./testdata/structure.tar.gz",
		},
		"TarArchive": {
			path: "./testdata/structure.tar.gz",
		},
		"TgzArchive": {
			path: "./testdata/structure.tgz",
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// Given
			g := gomega.NewGomegaWithT(t)

			tmpDir := "../../tmp"
			err := os.MkdirAll(tmpDir, os.ModePerm)
			g.Expect(err).NotTo(gomega.HaveOccurred())
			defer os.RemoveAll(tmpDir)

			loader := &loader{
				temporaryDir:    tmpDir,
				osRemoveAllFunc: os.RemoveAll,
				osCreateFunc:    os.Create,
				httpGetFunc:     getFile(testCase.path),
				ioutilTempDir:   ioutil.TempDir,
			}

			// When
			basePath, files, err := loader.Load(testCase.path, "asset", v1alpha2.AssetPackage, "")
			defer loader.Clean(basePath)

			// Then
			g.Expect(err).NotTo(gomega.HaveOccurred())
			g.Expect(files).To(gomega.HaveLen(2))
			g.Expect(files).To(gomega.ConsistOf(expected))
		})
	}
}

func TestLoader_Load_WithFilter(t *testing.T) {
	for testName, testCase := range map[string]struct {
		filter   string
		expected []string
	}{
		"AllFiles": {
			filter: ".*",
			expected: []string{
				"test/README.md",
				"test/nested/nested.md",
				"swagger.json",
				"docs/README.md",
			},
		},
		"MarkdownFiles": {
			filter: ".*\\.md$",
			expected: []string{
				"test/README.md",
				"test/nested/nested.md",
				"docs/README.md",
			},
		},
		"ReadmeFiles": {
			filter: "(^|/)README\\.md$",
			expected: []string{
				"test/README.md",
				"docs/README.md",
			},
		},
		"FilesFromTestDirectory": {
			filter: "^test/.*",
			expected: []string{
				"test/README.md",
				"test/nested/nested.md",
			},
		},
		"NoFiles": {
			filter:   "^nomatch$",
			expected: []string{},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// Given
			g := gomega.NewGomegaWithT(t)

			tmpDir := "../../tmp"
			testPath := "./testdata/complex.zip"
			err := os.MkdirAll(tmpDir, os.ModePerm)
			g.Expect(err).NotTo(gomega.HaveOccurred())
			defer os.RemoveAll(tmpDir)

			loader := &loader{
				temporaryDir:    tmpDir,
				osRemoveAllFunc: os.RemoveAll,
				osCreateFunc:    os.Create,
				httpGetFunc:     getFile(testPath),
				ioutilTempDir:   ioutil.TempDir,
			}

			// When
			basePath, files, err := loader.Load(testPath, "asset", v1alpha2.AssetPackage, testCase.filter)
			defer loader.Clean(basePath)

			// Then
			g.Expect(err).NotTo(gomega.HaveOccurred())
			g.Expect(files).To(gomega.HaveLen(len(testCase.expected)))
			g.Expect(files).To(gomega.ConsistOf(testCase.expected))
		})
	}
}

func getFile(path string) func(url string) (*http.Response, error) {
	file, err := os.Open(path)
	if err != nil {
		return func(url string) (*http.Response, error) {
			return nil, err
		}
	}

	get := func(url string) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       file,
		}, nil
	}

	return get
}
