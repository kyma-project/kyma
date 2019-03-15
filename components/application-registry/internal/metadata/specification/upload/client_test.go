package upload

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestUploadClient(t *testing.T){

	testContent := []byte("test content")

	t.Run("Should upload single file", func (t *testing.T){
		// given
		testServer := getTestServer(t)
		client := http.Client{}
		u, err := url.Parse(testServer.URL)
		require.NoError(t, err)
		uploadClient := NewUploadClient(u, client)

		// when
		input := InputFile{
			Directory: "testDir",
			Name: "filename",
			Contents: testContent,
		}
		output, err := uploadClient.Upload(input)

		assert.Equal(t, output.Bucket, "test bucket")
		assert.Equal(t, output.FileName, "testFile")
		assert.Equal(t, output.RemotePath, "testPath")

		// then

	})

	t.Run("Should fail when uploading file failed", func (t *testing.T){
		// given

		// when

		// then

	})
}

func getTestServer(t *testing.T) *httptest.Server{
	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseMultipartForm(32 << 20)
			require.NoError(t, err)
			require.NotNil(t, r.MultipartForm)

			files := r.MultipartForm.File
			require.Equal(t, len(files), 1)

			filesList, found := files["uploadedfile"]
			require.True(t, found)
			require.Equal(t, len(filesList), 1)

			fileHeader := filesList[0]
			assert.Equal(t, fileHeader.Filename, "test.json")
			assert.NotZero(t, fileHeader.Size)

			w.WriteHeader(http.StatusOK)
	}))
}