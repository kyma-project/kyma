package route_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/asset-metadata-service/internal/route"
	"github.com/onsi/gomega/gstruct"

	"github.com/onsi/gomega"
)

func TestRequestHandler_ServeHTTP(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		files := []RequestFile{
			{
				FieldName: "/testdata/success.md",
				Path:      "./testdata/success.md",
			},
			{
				FieldName: "/testdata/success.yaml",
				Path:      "./testdata/success.yaml",
			},
		}

		expectedSuccess := []ExpectedSuccess{
			{
				FilePath: files[1].FieldName,
				MetadataKeys: gstruct.Keys{
					"title":  gomega.Equal("Hello world"),
					"number": gomega.Equal(float64(9)),
					"url":    gomega.Equal("https://kyma-project.io"),
				},
			},
			{
				FilePath: files[0].FieldName,
				MetadataKeys: gstruct.Keys{
					"title": gomega.Equal("Access logs"),
					"type":  gomega.Equal("Details"),
					"no":    gomega.Equal(float64(3)),
				},
			},
		}

		// When

		httpResp, result := testServeHTTP(g, files)

		// Then

		g.Expect(httpResp.StatusCode).Should(gomega.Equal(http.StatusOK))
		g.Expect(result.Errors).To(gomega.BeEmpty())

		assertResponseDataEqual(t, g, result.Data, expectedSuccess)
	})

	t.Run("No files to process", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		var files []RequestFile

		// When
		httpResp, result := testServeHTTP(g, files)

		// Then
		g.Expect(httpResp.StatusCode).To(gomega.Equal(http.StatusBadRequest))
		g.Expect(result.Errors).To(gomega.HaveLen(1))
		g.Expect(result.Errors[0].Message).To(gomega.ContainSubstring("No files"))
	})

	t.Run("Partial Errors", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		files := []RequestFile{
			{
				FieldName: "/testdata/success.md",
				Path:      "./testdata/success.md",
			},
			{
				FieldName: "/testdata/error.md",
				Path:      "./testdata/error.md",
			},
		}

		expectedSuccess := []ExpectedSuccess{
			{
				FilePath: files[0].FieldName,
				MetadataKeys: gstruct.Keys{
					"title": gomega.Equal("Access logs"),
					"type":  gomega.Equal("Details"),
					"no":    gomega.Equal(float64(3)),
				},
			},
		}
		expectedErrors := []route.ResultError{
			{Message: "Error while processing file `/testdata/error.md`: while reading metadata from file error.md: front: unknown delim", FilePath: "/testdata/error.md"},
		}

		// When
		httpResp, result := testServeHTTP(g, files)

		// Then
		g.Expect(httpResp.StatusCode).To(gomega.Equal(http.StatusMultiStatus))

		assertResponseDataEqual(t, g, result.Data, expectedSuccess)

		for _, responseError := range expectedErrors {
			g.Expect(result.Errors).To(gomega.ContainElement(responseError))
		}
	})

	t.Run("Errors", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		files := []RequestFile{
			{
				FieldName: "sample/error.md",
				Path:      "./testdata/error.md",
			},
			{
				FieldName: "sample/error.yaml",
				Path:      "./testdata/error.yaml",
			},
		}
		expectedResult := []route.ResultError{
			{Message: "Error while processing file `sample/error.md`: while reading metadata from file error.md: front: unknown delim", FilePath: files[0].FieldName},
			{Message: "Error while processing file `sample/error.yaml`: while reading metadata from file error.yaml: front: unknown delim", FilePath: files[1].FieldName},
		}

		// When
		httpResp, result := testServeHTTP(g, files)

		// Then
		g.Expect(httpResp.StatusCode).To(gomega.Equal(http.StatusUnprocessableEntity))

		g.Expect(result.Data).To(gomega.BeEmpty())

		for _, responseError := range expectedResult {
			g.Expect(result.Errors).To(gomega.ContainElement(responseError))
		}
	})
}

type RequestFile struct {
	Path      string
	FieldName string
}

type ExpectedSuccess struct {
	FilePath     string
	MetadataKeys gstruct.Keys
}

func testServeHTTP(g *gomega.GomegaWithT, files []RequestFile) (*http.Response, route.Response) {
	handler := route.NewExtractHandler(5, 10*time.Second)

	w := httptest.NewRecorder()
	rq, err := fixRequest(files)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	handler.ServeHTTP(w, rq)

	resp := w.Result()
	g.Expect(resp).NotTo(gomega.BeNil())

	defer func() {
		err := resp.Body.Close()
		g.Expect(err).NotTo(gomega.HaveOccurred())
	}()

	var result route.Response
	err = json.NewDecoder(resp.Body).Decode(&result)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	return resp, result
}

func fixRequest(files []RequestFile) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, f := range files {
		file, err := os.Open(f.Path)
		if err != nil {
			return nil, err
		}

		part, err := writer.CreateFormFile(f.FieldName, filepath.Base(file.Name()))
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}

		err = file.Close()
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	rq, err := http.NewRequest("POST", "example.com", body)
	if err != nil {
		return nil, err
	}

	rq.Header.Add("Content-Type", writer.FormDataContentType())

	return rq, nil
}

func assertResponseDataEqual(t *testing.T, g *gomega.GomegaWithT, respData []route.ResultSuccess, expectedSuccess []ExpectedSuccess) {
	g.Expect(respData).To(gomega.HaveLen(len(expectedSuccess)))
	for _, successResult := range respData {
		idx := -1
		for index, expected := range expectedSuccess {
			if expected.FilePath == successResult.FilePath {
				idx = index
			}
		}

		if idx == -1 {
			t.Errorf("Unexpected item with FilePath %s", successResult.FilePath)
		}

		g.Expect(successResult.Metadata).To(gstruct.MatchAllKeys(expectedSuccess[idx].MetadataKeys))
	}
}
