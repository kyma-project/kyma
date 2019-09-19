package compassconnection

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const (
	directorURL = "https://compass-geteway.kyma.local/director/graphql"
)

func TestCompassConnector_EstablishConnection(t *testing.T) {

	t.Run("should establish connection", func(t *testing.T) {
		// given
		connector := NewCompassConnector()

		// when
		connection, err := EstablishConnection()

		// then
		require.NoError(t, err)
		assert.Equal(t, directorURL, connection.DirectorURL)
	})

	t.Run("should return error if director URL is empty", func(t *testing.T) {
		// given
		connector := NewCompassConnector("")

		// when
		_, err := EstablishConnection()

		// then
		require.Error(t, err)
	})

}
