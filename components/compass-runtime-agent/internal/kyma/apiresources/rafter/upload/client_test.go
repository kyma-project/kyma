package upload

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"kyma-project.io/compass-runtime-agent/internal/httpconsts"
)

func TestUploadClient(t *testing.T) {

	testContent := []byte("test content")

	t.Run("Should upload single file", func(t *testing.T) {
		// given
		testServer := getTestServer(t)
		uploadClient := NewClient(testServer.URL)

		// when
		output, err := uploadClient.Upload("testfile", testContent)
		require.NoError(t, err)

		// then
		assert.Equal(t, "testBucket", output.Bucket)
		assert.Equal(t, "testFile", output.FileName)
		assert.Equal(t, "testDir/testPath", output.RemotePath)
		assert.Equal(t, int64(10), output.Size)
	})

	t.Run("Should fail when uploading file failed", func(t *testing.T) {
		// given
		uploadClient := NewClient("non-existent-url")

		// when
		_, err := uploadClient.Upload("testfile", testContent)

		// then
		assert.Error(t, err)
	})

	t.Run("Should fail when upload service returned 500 status", func(t *testing.T) {
		// given
		testServer := getTestServerWithStatus(t, http.StatusInternalServerError)
		uploadClient := NewClient(testServer.URL)

		// when
		_, err := uploadClient.Upload("testfile", testContent)

		// then
		assert.Error(t, err)
	})
}

func getTestServerWithStatus(t *testing.T, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			response := Response{
				UploadedFiles: []UploadedFile{},
				Errors: []ResponseError{
					{
						FileName: "testfile",
						Message:  "failed",
					},
				},
			}

			var b bytes.Buffer
			err := json.NewEncoder(&b).Encode(response)
			require.NoError(t, err)

			w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
			w.WriteHeader(status)
		}))
}

func getTestServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseMultipartForm(32 << 20)
			defer func() {
				err := r.MultipartForm.RemoveAll()
				require.NoError(t, err)
			}()

			defer func() {
				err = r.Body.Close()
				require.NoError(t, err)
			}()

			require.NoError(t, err)
			require.NotNil(t, r.MultipartForm)

			validateMultipartForm(t, r)

			outputFile := UploadedFile{
				FileName:   "testFile",
				RemotePath: "testDir/testPath",
				Bucket:     "testBucket",
				Size:       10,
			}

			response := Response{
				UploadedFiles: []UploadedFile{outputFile},
			}

			var b bytes.Buffer
			err = json.NewEncoder(&b).Encode(response)
			require.NoError(t, err)

			w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
			w.WriteHeader(http.StatusOK)

			_, err = w.Write(b.Bytes())
			assert.NoError(t, err)
		}))
}

func validateMultipartForm(t *testing.T, r *http.Request) {
	files := r.MultipartForm.File
	require.Equal(t, len(files), 1)

	filesList, found := files[PublicFileField]
	require.True(t, found)
	require.Equal(t, len(filesList), 1)

	fileHeader := filesList[0]
	assert.Equal(t, "testfile", fileHeader.Filename)
	f, err := fileHeader.Open()
	require.NoError(t, err)

	defer func() {
		err = f.Close()
		assert.NoError(t, err)
	}()

	b := make([]byte, fileHeader.Size)

	_, err = f.Read(b)
	require.NoError(t, err)

	assert.Equal(t, []byte("test content"), b)
	assert.NotZero(t, fileHeader.Size)
}
