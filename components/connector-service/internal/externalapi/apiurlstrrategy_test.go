package externalapi

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationUrlsStrategy_Generate(t *testing.T) {

	appRegistryHost := "gateway.kyma.local"
	eventsHost := "gateway.kyma.local"
	host := "connector-service.kyma.local"

	t.Run("should generate API URLs", func(t *testing.T) {
		// given
		infoURL := "https://management.kyma.local/info"

		urlsGenerator := NewApplicationApiUrlsStrategy(appRegistryHost, eventsHost, infoURL, host)

		contextReader := &mocks.ContextReader{}
		contextReader.On("GetApplication").Return(appName)

		// when
		api := urlsGenerator.Generate(contextReader)

		// then
		applicationAPI, ok := api.(applicationApi)
		require.True(t, ok)

		assert.Equal(t, fmt.Sprintf(MetadataURLFormat, appRegistryHost, appName), applicationAPI.MetadataURL)
		assert.Equal(t, fmt.Sprintf(EventsURLFormat, eventsHost, appName), applicationAPI.EventsURL)
		assert.Equal(t, infoURL, applicationAPI.InfoURL)
		assert.Equal(t, fmt.Sprintf(AppCertificatesURLFormat, host), applicationAPI.CertificatesURL)
	})

	t.Run("should construct InfoURL if not provided", func(t *testing.T) {
		// given
		urlsGenerator := NewApplicationApiUrlsStrategy(appRegistryHost, eventsHost, "", host)

		contextReader := &mocks.ContextReader{}
		contextReader.On("GetApplication").Return(appName)

		// when
		api := urlsGenerator.Generate(contextReader)

		// then
		applicationAPI, ok := api.(applicationApi)
		require.True(t, ok)

		assert.Equal(t, fmt.Sprintf(MetadataURLFormat, appRegistryHost, appName), applicationAPI.MetadataURL)
		assert.Equal(t, fmt.Sprintf(EventsURLFormat, eventsHost, appName), applicationAPI.EventsURL)
		assert.Equal(t, fmt.Sprintf(AppManagementInfoURLFormat, host), applicationAPI.InfoURL)
		assert.Equal(t, fmt.Sprintf(AppCertificatesURLFormat, host), applicationAPI.CertificatesURL)
	})

}

func TestRuntimeUrlsStrategy_Generate(t *testing.T) {

	host := "connector-service.kyma.local"

	t.Run("should generate runtimes URLs", func(t *testing.T) {
		// given
		contextReader := &mocks.ContextReader{}

		urlsGenerator := NewRuntimeApiUrlsStrategy(host)

		// when
		api := urlsGenerator.Generate(contextReader)

		// then
		runtimeAPI, ok := api.(runtimeApi)
		require.True(t, ok)

		assert.Equal(t, fmt.Sprintf(RuntimeManagementInfoURLFormat, host), runtimeAPI.InfoURL)
		assert.Equal(t, fmt.Sprintf(RuntimeCertificatesURLFormat, host), runtimeAPI.CertificatesURL)
	})
}
