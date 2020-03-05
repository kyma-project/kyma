package broker_test

import (
	"context"
	"fmt"
	"testing"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker/automock"
)

func TestBindServiceBindSuccess(t *testing.T) {
	// given
	appFinder := &automock.AppFinder{}
	defer appFinder.AssertExpectations(t)

	fixApp := fixApplication()

	appFinder.On("FindOneByServiceID", fixApp.Services[0].ID).
		Return(&fixApp, nil).
		Once()

	osbCtx := broker.NewOSBContext("not", "important", "")
	svc := broker.NewBindServiceV1(appFinder)
	// when
	resp, err := svc.Bind(context.Background(), *osbCtx, fixBindRequest())

	// then
	require.NoError(t, err)
	require.NotNil(t, resp.Credentials)
	assert.Equal(t, "www.gate.com", resp.Credentials["GATEWAY_URL"])

}

func TestBindServiceBindFailure(t *testing.T) {
	t.Run("On credentials get error", func(t *testing.T) {
		// given
		fixApp := fixApplication()
		fixID := fixBindRequest().InstanceID

		svc := broker.NewBindServiceV1(nil)

		// when
		resp, err := svc.GetCredentials(internal.ApplicationServiceID(fixID), &fixApp)

		// then
		assert.EqualError(t, err, fmt.Sprintf("cannot get credentials to bind instance with ApplicationServiceID: %s, from Application: %s", fixID, fixApp.Name))
		assert.Zero(t, resp)
	})

	t.Run("On unexpected req params", func(t *testing.T) {
		//given
		fixReq := fixBindRequest()
		fixReq.Parameters = map[string]interface{}{
			"some-key": "some-value",
		}

		svc := broker.NewBindServiceV1(nil)
		osbCtx := broker.NewOSBContext("not", "important", "")

		// when
		resp, err := svc.Bind(context.Background(), *osbCtx, fixReq)

		// then
		assert.EqualError(t, err, "application-broker does not support configuration options for the service binding")
		assert.Zero(t, resp)
	})
}

func TestBindServiceBindSuccessV2(t *testing.T) {
	// given
	appFinder := &automock.AppFinder{}
	defer appFinder.AssertExpectations(t)

	fixApp := fixApplicationV2()
	appFinder.On("FindOneByServiceID", fixApp.Services[0].ID).
		Return(&fixApp, nil).
		Once()

	osbCtx := broker.NewOSBContext("not", "important", "")
	svc := broker.NewBindServiceV2(appFinder)

	// when
	resp, err := svc.Bind(context.Background(), *osbCtx, fixBindRequest())

	// then
	require.NoError(t, err)
	require.NotNil(t, resp.Credentials)

	assert.Equal(t, "http://promotions.gateway.io", resp.Credentials["API_MOCK_PROMOTIONS_GATEWAY_URL"])
	assert.Equal(t, "http://promotions.target.io", resp.Credentials["API_MOCK_PROMOTIONS_TARGET_URL"])

	assert.Equal(t, "http://occ.gateway.io", resp.Credentials["API_MOCK_OCC_GATEWAY_URL"])
	assert.Equal(t, "http://occ.target.io", resp.Credentials["API_MOCK_OCC_TARGET_URL"])
}

func TestBindServiceBindFailureV2(t *testing.T) {
	t.Run("On credentials get error", func(t *testing.T) {
		// given
		fixApp := fixApplication()
		fixID := "some-wrong-id"

		svc := broker.NewBindServiceV2(nil)

		// when
		resp, err := svc.GetCredentials(internal.ApplicationServiceID(fixID), &fixApp)

		// then
		assert.EqualError(t, err, fmt.Sprintf("cannot get credentials to bind instance with ApplicationServiceID: %s, from Application: %s", fixID, fixApp.Name))
		assert.Zero(t, resp)
	})

	t.Run("On unexpected req params", func(t *testing.T) {
		//given
		fixReq := fixBindRequest()
		fixReq.Parameters = map[string]interface{}{
			"some-key": "some-value",
		}

		svc := broker.NewBindServiceV2(nil)
		osbCtx := broker.NewOSBContext("not", "important", "")

		// when
		resp, err := svc.Bind(context.Background(), *osbCtx, fixReq)

		// then
		assert.EqualError(t, err, "application-broker does not support configuration options for the service binding")
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

func fixApplication() internal.Application {
	return internal.Application{
		Name: "ec-prod",
		Services: []internal.Service{
			{
				ID: "123",
				Entries: []internal.Entry{
					{
						Type: "API",
						APIEntry: &internal.APIEntry{
							GatewayURL: "www.gate.com",
						},
					},
					{
						Type: "Events",
					},
				},
			},
		},
	}
}

func fixApplicationV2() internal.Application {
	return internal.Application{
		Name: "ec-prod",
		Services: []internal.Service{
			{
				ID: "plan-id",
				Entries: []internal.Entry{
					{
						Type: "API",
						APIEntry: &internal.APIEntry{
							Name:       "api_mock_promotions",
							TargetURL:  "http://promotions.target.io",
							GatewayURL: "http://promotions.gateway.io",
						},
					},
					{
						Type: "API",
						APIEntry: &internal.APIEntry{
							Name:       "api_mock_occ",
							TargetURL:  "http://occ.target.io",
							GatewayURL: "http://occ.gateway.io",
						},
					},
					{
						Type: "Events",
					},
				},
			},
		},
	}
}
