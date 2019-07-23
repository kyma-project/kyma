package compass

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	tokenURLData = `{
	"url":"https://token.url/applications/info?token=abcd",
	"token":"abcd"
}`
)

func TestCompassConnector_EstablishConnection(t *testing.T) {

	file, err := ioutil.TempFile(".", "tokenFile")
	require.NoError(t, err)
	defer func() {
		err := os.Remove(file.Name())
		require.NoError(t, err)
	}()

	_, err = file.Write([]byte(tokenURLData))
	require.NoError(t, err)

	t.Run("should establish connection", func(t *testing.T) {
		// given
		connector := NewCompassConnector(file.Name())

		// when
		err = connector.EstablishConnection()

		// then
		require.NoError(t, err)
	})

	t.Run("should return error when failed to read file", func(t *testing.T) {
		// given
		connector := NewCompassConnector("non-existing")

		// when
		err = connector.EstablishConnection()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when invalid file content", func(t *testing.T) {
		// given
		invalidFile, err := ioutil.TempFile(".", "invalidFile")
		require.NoError(t, err)
		defer func() {
			err := os.Remove(invalidFile.Name())
			require.NoError(t, err)
		}()

		_, err = invalidFile.Write([]byte("invalid data"))
		require.NoError(t, err)

		connector := NewCompassConnector(invalidFile.Name())

		// when
		err = connector.EstablishConnection()

		// then
		require.Error(t, err)
	})

}
