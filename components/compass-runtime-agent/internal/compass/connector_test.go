package compass

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestCompassConnector_EstablishConnection(t *testing.T) {

	t.Run("should establish connection", func(t *testing.T) {
		// given
		connector := NewCompassConnector(directorURL)

		// when
		connection, err := connector.EstablishConnection()

		// then
		require.NoError(t, err)
		assert.Equal(t, directorURL, connection.DirectorURL)
	})

	t.Run("should return error if director URL is empty", func(t *testing.T) {
		// given
		connector := NewCompassConnector("")

		// when
		_, err := connector.EstablishConnection()

		// then
		require.Error(t, err)
	})

}
