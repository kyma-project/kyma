package remoteenvironment_test

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/gateway"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteEnvironmentStatusSuccess(t *testing.T) {
	for tn, tc := range map[string]struct {
		givenStatus    gateway.Status
		expectedStatus gqlschema.RemoteEnvironmentStatus
	}{
		"serving": {
			givenStatus:    gateway.StatusServing,
			expectedStatus: gqlschema.RemoteEnvironmentStatusServing,
		},
		"not serving": {
			givenStatus:    gateway.StatusNotServing,
			expectedStatus: gqlschema.RemoteEnvironmentStatusNotServing,
		},
		"not configured": {
			givenStatus:    gateway.StatusNotConfigured,
			expectedStatus: gqlschema.RemoteEnvironmentStatusGatewayNotConfigured,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			statusGetterStub := automock.NewStatusGetter()
			statusGetterStub.On("GetStatus", "ec-prod").Return(tc.givenStatus, nil)
			resolver := remoteenvironment.NewRemoteEnvironmentResolver(nil, statusGetterStub)

			// when
			status, err := resolver.RemoteEnvironmentStatusField(context.Background(), &gqlschema.RemoteEnvironment{
				Name: "ec-prod",
			})

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}

func TestRemoteEnvironmentStatusFail(t *testing.T) {
	// given
	statusGetterStub := automock.NewStatusGetter()
	statusGetterStub.On("GetStatus", "ec-prod").Return(gateway.Status("fake"), nil)
	resolver := remoteenvironment.NewRemoteEnvironmentResolver(nil, statusGetterStub)

	// when
	_, err := resolver.RemoteEnvironmentStatusField(context.Background(), &gqlschema.RemoteEnvironment{
		Name: "ec-prod",
	})

	// then
	require.Error(t, err)
}

func TestConnectorServiceQuerySuccess(t *testing.T) {
	// given
	var (
		fixRemoteEnvName = "env-name"
		fixURL           = "http://some-url-with-token"
		fixGQLObj        = gqlschema.ConnectorService{
			Url: "http://some-url-with-token",
		}
	)

	serviceMock := automock.NewReSvc()
	defer serviceMock.AssertExpectations(t)
	serviceMock.On("GetConnectionUrl", fixRemoteEnvName).Return(fixURL, nil)

	resolver := remoteenvironment.NewRemoteEnvironmentResolver(serviceMock, nil)

	// when
	gotUrlObj, err := resolver.ConnectorServiceQuery(context.Background(), fixRemoteEnvName)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLObj, gotUrlObj)
}

func TestConnectorServiceQueryFail(t *testing.T) {
	// given
	var (
		fixRemoteEnvName = ""
		fixErr           = errors.New("something went wrong")
	)

	serviceMock := automock.NewReSvc()
	defer serviceMock.AssertExpectations(t)
	serviceMock.On("GetConnectionUrl", fixRemoteEnvName).Return("", fixErr)

	resolver := remoteenvironment.NewRemoteEnvironmentResolver(serviceMock, nil)

	// when
	gotUrlObj, err := resolver.ConnectorServiceQuery(context.Background(), fixRemoteEnvName)

	// then
	require.Error(t, err)
	assert.Zero(t, gotUrlObj)
	assert.Equal(t, "while getting Connection Url: something went wrong", err.Error())
}
