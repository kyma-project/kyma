package broker_test

import (
	"context"
	"testing"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker/automock"
)

const fieldNameGatewayURL = "GATEWAY_URL"

func TestBindServiceBindSuccess(t *testing.T) {
	// given
	reFinder := &automock.ReFinder{}
	defer reFinder.AssertExpectations(t)

	fixRE := fixRE()

	reFinder.On("FindOneByServiceID", fixRE.Services[0].ID).
		Return(&fixRE, nil).
		Once()

	osbCtx := broker.NewOSBContext("not", "important", "")
	svc := broker.NewBindService(reFinder)
	// when
	resp, err := svc.Bind(context.Background(), *osbCtx, fixBindRequest())

	// then
	require.NoError(t, err)
	require.NotNil(t, resp.Credentials)
	assert.Equal(t, "www.gate.com", resp.Credentials[fieldNameGatewayURL])

}

func TestBindServiceBindFailure(t *testing.T) {
	t.Run("On credentials get error", func(t *testing.T) {
		// given
		fixRE := fixRE()
		fixID := fixBindRequest().InstanceID

		svc := broker.NewBindService(nil)

		// when
		resp, err := svc.GetCredentials(internal.RemoteServiceID(fixID), &fixRE)

		// then
		assert.EqualErrorf(t, err, err.Error(), "cannot get credentials to bind instance wit RemoteServiceID: %s, from RemoteEnvironment: %s", fixID, fixRE.Name)
		assert.Zero(t, resp)
	})

	t.Run("On unexpected req params", func(t *testing.T) {
		//given
		fixReq := fixBindRequest()
		fixReq.Parameters = map[string]interface{}{
			"some-key": "some-value",
		}

		svc := broker.NewBindService(nil)
		osbCtx := broker.NewOSBContext("not", "important", "")

		// when
		resp, err := svc.Bind(context.Background(), *osbCtx, fixReq)

		// then
		assert.EqualError(t, err, "remote-environment-broker does not support configuration options for the service binding")
		assert.Zero(t, resp)
	})
}

func fixBindRequest() *osb.BindRequest {
	return &osb.BindRequest{
		InstanceID: "instance-id",
		ServiceID:  "123",
		PlanID:     "plan-id",
	}
}
func fixRE() internal.RemoteEnvironment {
	return internal.RemoteEnvironment{
		Name: "ec-prod",
		Services: []internal.Service{
			{
				ID: "123",
				APIEntry: &internal.APIEntry{
					GatewayURL:  "www.gate.com",
					AccessLabel: "free",
				},
			},
		},
	}
}
