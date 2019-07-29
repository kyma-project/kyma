package kyma

import (
	resourcesServiceMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/mocks"
	appMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications/mocks"
	syncMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/sync/mocks"
	"testing"
)

func TestConverter(t *testing.T) {

	t.Run("should apply Create operation", func(t *testing.T) {
		// given
		//reconciler sync.Reconciler, applicationRepository applications.Manager, converter applications.Converter, resourcesService apiresources.Service
		reconcilerMock := &syncMocks.Reconciler{}
		applicationsManagerMock := &appMocks.Manager{}
		converterMock := &appMocks.Converter{}
		resourcesServiceMocks := &resourcesServiceMocks.Service{}

		reconcilerMock.On("")
		applicationsManagerMock.On("")
		converterMock.On("")
		resourcesServiceMocks.On("")

		// when

		// then

	})

	t.Run("should apply Update operation", func(t *testing.T) {

	})

	t.Run("should apply Delete operation", func(t *testing.T) {

	})
}
