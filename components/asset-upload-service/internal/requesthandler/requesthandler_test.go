package requesthandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/asset-upload-service/internal/bucket"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/requesthandler"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/uploader"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/uploader/automock"
	"github.com/minio/minio-go"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
)

func TestRequestHandler_ServeHTTP(t *testing.T) {
	anyReaderFn := func(reader io.Reader) bool { return true }
	anySizeFn := func(size int64) bool { return true }
	ctxArgFn := func(ctx context.Context) bool { return true }
	randomDirFn := func(name string) func(string) bool {
		return func(filename string) bool { return strings.HasSuffix(filename, name) }
	}

	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		client := &automock.MinioClient{}
		client.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), "public", mock.MatchedBy(randomDirFn("sample.yaml")), mock.MatchedBy(anyReaderFn), mock.MatchedBy(anySizeFn), minio.PutObjectOptions{}).Return(int64(1), nil).Once()
		client.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), "private", mock.MatchedBy(randomDirFn("sample.txt")), mock.MatchedBy(anyReaderFn), mock.MatchedBy(anySizeFn), minio.PutObjectOptions{}).Return(int64(1), nil).Once()
		defer client.AssertExpectations(t)

		files := []RequestFile{
			{
				FieldName: "private",
				Path:      "./testdata/sample.txt",
			},
			{
				FieldName: "public",
				Path:      "./testdata/sample.yaml",
			},
		}

		expectedResult := []uploader.UploadResult{
			{
				FileName:   "sample.yaml",
				RemotePath: "https://example.com/public/",
				Bucket:     "public",
				Size:       53,
			},
			{
				FileName:   "sample.txt",
				RemotePath: "https://example.com/private/",
				Bucket:     "private",
				Size:       16,
			},
		}

		// When

		httpResp, result := testServeHTTP(g, client, files, "")

		// Then

		g.Expect(httpResp.StatusCode).To(gomega.Equal(http.StatusOK))

		removeRemotePathFromFiles(&result)

		g.Expect(result.Errors).To(gomega.BeEmpty())

		for _, file := range expectedResult {
			g.Expect(result.UploadedFiles).To(gomega.ContainElement(file))
		}
	})

	t.Run("Custom Directory", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		client := &automock.MinioClient{}
		client.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), "public", mock.MatchedBy(randomDirFn("sample.yaml")), mock.MatchedBy(anyReaderFn), mock.MatchedBy(anySizeFn), minio.PutObjectOptions{}).Return(int64(1), nil).Once()
		client.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), "private", mock.MatchedBy(randomDirFn("sample.txt")), mock.MatchedBy(anyReaderFn), mock.MatchedBy(anySizeFn), minio.PutObjectOptions{}).Return(int64(1), nil).Once()
		defer client.AssertExpectations(t)

		files := []RequestFile{
			{
				FieldName: "private",
				Path:      "./testdata/sample.txt",
			},
			{
				FieldName: "public",
				Path:      "./testdata/sample.yaml",
			},
		}
		directoryName := "test"
		expectedResult := []uploader.UploadResult{
			{
				FileName:   "sample.yaml",
				RemotePath: "https://example.com/public/test/sample.yaml",
				Bucket:     "public",
				Size:       53,
			},
			{
				FileName:   "sample.txt",
				RemotePath: "https://example.com/private/test/sample.txt",
				Bucket:     "private",
				Size:       16,
			},
		}

		// When
		httpResp, result := testServeHTTP(g, client, files, directoryName)

		// Then
		g.Expect(httpResp.StatusCode).To(gomega.Equal(http.StatusOK))

		g.Expect(result.Errors).To(gomega.BeEmpty())

		for _, file := range expectedResult {
			g.Expect(result.UploadedFiles).To(gomega.ContainElement(file))
		}
	})

	t.Run("No files to upload", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		client := &automock.MinioClient{}
		var files []RequestFile

		// When
		httpResp, result := testServeHTTP(g, client, files, "")

		// Then
		g.Expect(httpResp.StatusCode).To(gomega.Equal(http.StatusBadRequest))
		g.Expect(result.Errors).To(gomega.HaveLen(1))
		g.Expect(result.Errors[0].Message).To(gomega.ContainSubstring("No files"))
	})

	t.Run("Partial Errors", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		client := &automock.MinioClient{}

		testErr1 := errors.New("Test err 1")
		client.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), "public", mock.MatchedBy(randomDirFn("sample.yaml")), mock.MatchedBy(anyReaderFn), mock.MatchedBy(anySizeFn), minio.PutObjectOptions{}).Return(int64(1), nil).Once()
		client.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), "private", mock.MatchedBy(randomDirFn("sample.txt")), mock.MatchedBy(anyReaderFn), mock.MatchedBy(anySizeFn), minio.PutObjectOptions{}).Return(int64(1), testErr1).Once()
		defer client.AssertExpectations(t)

		files := []RequestFile{
			{
				FieldName: "private",
				Path:      "./testdata/sample.txt",
			},
			{
				FieldName: "public",
				Path:      "./testdata/sample.yaml",
			},
		}
		directoryName := "test"
		expectedErrors := []requesthandler.ResponseError{
			{Message: "Error while uploading file `test/sample.txt` into `private`: Test err 1", FileName: "sample.txt"},
		}
		expectedResult := []uploader.UploadResult{
			{
				FileName:   "sample.yaml",
				RemotePath: "https://example.com/public/",
				Bucket:     "public",
				Size:       53,
			},
		}

		// When
		httpResp, result := testServeHTTP(g, client, files, directoryName)

		// Then
		g.Expect(httpResp.StatusCode).To(gomega.Equal(http.StatusMultiStatus))

		removeRemotePathFromFiles(&result)
		for _, file := range expectedResult {
			g.Expect(result.UploadedFiles).To(gomega.ContainElement(file))
		}

		for _, responseError := range expectedErrors {
			g.Expect(result.Errors).To(gomega.ContainElement(responseError))
		}
	})

	t.Run("Errors", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		client := &automock.MinioClient{}

		testErr1 := errors.New("Test err 1")
		testErr2 := errors.New("Test err 2")
		client.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), "public", mock.MatchedBy(randomDirFn("sample.yaml")), mock.MatchedBy(anyReaderFn), mock.MatchedBy(anySizeFn), minio.PutObjectOptions{}).Return(int64(1), testErr1).Once()
		client.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), "private", mock.MatchedBy(randomDirFn("sample.txt")), mock.MatchedBy(anyReaderFn), mock.MatchedBy(anySizeFn), minio.PutObjectOptions{}).Return(int64(1), testErr2).Once()
		defer client.AssertExpectations(t)

		files := []RequestFile{
			{
				FieldName: "private",
				Path:      "./testdata/sample.txt",
			},
			{
				FieldName: "public",
				Path:      "./testdata/sample.yaml",
			},
		}
		directoryName := "test"
		expectedResult := []requesthandler.ResponseError{
			{Message: "Error while uploading file `test/sample.yaml` into `public`: Test err 1", FileName: "sample.yaml"},
			{Message: "Error while uploading file `test/sample.txt` into `private`: Test err 2", FileName: "sample.txt"},
		}

		// When
		httpResp, result := testServeHTTP(g, client, files, directoryName)

		// Then
		g.Expect(httpResp.StatusCode).To(gomega.Equal(http.StatusBadGateway))

		g.Expect(result.UploadedFiles).To(gomega.BeEmpty())

		for _, responseError := range expectedResult {
			g.Expect(result.Errors).To(gomega.ContainElement(responseError))
		}
	})
}

type RequestFile struct {
	Path      string
	FieldName string
}

func testServeHTTP(g *gomega.GomegaWithT, minioClient uploader.MinioClient, files []RequestFile, directoryName string) (*http.Response, requesthandler.Response) {
	buckets := bucket.SystemBucketNames{
		Private: "private",
		Public:  "public",
	}

	handler := requesthandler.New(minioClient, buckets, "https://example.com", 10*time.Second, 5)

	w := httptest.NewRecorder()
	rq, err := fixRequest(files, directoryName)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	handler.ServeHTTP(w, rq)

	resp := w.Result()
	g.Expect(resp).NotTo(gomega.BeNil())

	defer func() {
		err := resp.Body.Close()
		g.Expect(err).NotTo(gomega.HaveOccurred())
	}()

	var result requesthandler.Response
	err = json.NewDecoder(resp.Body).Decode(&result)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	return resp, result
}

func fixRequest(files []RequestFile, directoryName string) (*http.Request, error) {
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

	if directoryName != "" {
		err := writer.WriteField("directory", directoryName)
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

func removeRemotePathFromFiles(result *requesthandler.Response) {
	var newFiles []uploader.UploadResult
	for _, file := range result.UploadedFiles {
		file.RemotePath = strings.SplitAfter(file.RemotePath, fmt.Sprintf("https://example.com/%s/", file.Bucket))[0]
		newFiles = append(newFiles, file)
	}

	result.UploadedFiles = newFiles
}
