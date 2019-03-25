package upload

import (
	"bytes"
	"encoding/json"
	"github.com/kyma-project/kyma/components/application-registry/internal/httpconsts"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUploadClient(t *testing.T) {

	testContent := []byte("test content")

	t.Run("Should upload single file", func(t *testing.T) {
		// given
		testServer := getTestServer(t)
		client := &http.Client{}

		uploadClient := NewClient(testServer.URL, client)

		// when
		input := InputFile{
			Directory: "testDir",
			Name:      "testfile",
			Contents:  testContent,
		}
		output, err := uploadClient.Upload(input)
		require.NoError(t, err)

		// then
		assert.Equal(t, output.Bucket, "testBucket")
		assert.Equal(t, output.FileName, "testFile")
		assert.Equal(t, output.RemotePath, "testDir/testPath")
		assert.Equal(t, output.Size, 10)
	})

	t.Run("Should fail when uploading file failed", func(t *testing.T) {
		// given

		// when

		// then

	})
}

func getTestServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseMultipartForm(32 << 20)
			defer func() {
				err := r.MultipartForm.RemoveAll()
				if err != nil {
					logrus.Error("Cleaning up multipart form data failed.")
				}
			}()

			defer r.Body.Close()

			require.NoError(t, err)
			require.NotNil(t, r.MultipartForm)

			files := r.MultipartForm.File
			require.Equal(t, len(files), 1)

			filesList, found := files[PrivateFileField]
			require.True(t, found)
			require.Equal(t, len(filesList), 1)

			fileHeader := filesList[0]
			assert.Equal(t, fileHeader.Filename, "testfile")
			assert.NotZero(t, fileHeader.Size)

			values := r.MultipartForm.Value
			directory, found := values[DirectoryField]
			assert.True(t, found)
			assert.Equal(t, []string{"testDir"}, directory)

			outputFile := UploadedFile{
				FileName:   "testFile",
				RemotePath: "testDir/testPath",
				Bucket:     "testBucket",
				Size:       10,
			}

			var b bytes.Buffer
			err = json.NewEncoder(&b).Encode(outputFile)
			require.NoError(t, err)

			w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
			w.WriteHeader(http.StatusOK)

			w.Write(b.Bytes())

		}))

}
