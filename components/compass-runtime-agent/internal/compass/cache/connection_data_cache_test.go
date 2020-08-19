package cache

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectionDataCache(t *testing.T) {

	connectorURL := "https://connector.kyma"
	directorURL := "https://director.kyma"
	cert := tls.Certificate{}

	assertionsSubscriber := func(data ConnectionData) error {
		assert.Equal(t, cert, data.Certificate)
		assert.Equal(t, directorURL, data.DirectorURL)
		assert.Equal(t, connectorURL, data.ConnectorURL)
		return nil
	}

	t.Run("should notify on UpdateConnectionData", func(t *testing.T) {
		// given
		cache := NewConnectionDataCache()
		cache.AddSubscriber(assertionsSubscriber)

		// when
		cache.UpdateConnectionData(cert, directorURL, connectorURL)
	})

	t.Run("should notify on UpdateURLs", func(t *testing.T) {
		// given
		cache := NewConnectionDataCache()
		cache.UpdateConnectionData(cert, "", "")

		cache.AddSubscriber(assertionsSubscriber)

		// when
		cache.UpdateURLs(directorURL, connectorURL)
	})
}
