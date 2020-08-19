package broker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker/automock"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/proxyconfig"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		fixInstanceID := fixBindRequest().InstanceID
		fixNamespace := fixBindRequest().Context["namespace"].(string)
		fixBindingID := fixBindRequest().BindingID
		appSvcID := internal.ApplicationServiceID("wrong-id")

		svc := broker.NewBindServiceV1(nil)

		// when
		resp, err := svc.GetCredentials(context.Background(), fixNamespace, appSvcID, fixBindingID, fixInstanceID, &fixApp)

		// then
		assert.EqualError(t, err, fmt.Sprintf("cannot get credentials to bind instance with ApplicationServiceID: %s, from Application: %s", appSvcID, fixApp.Name))
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
	var (
		fixApp               = fixApplicationV2()
		ctx                  = context.Background()
		fixReq               = fixBindRequest()
		fixNamespace         = fixReq.Context["namespace"].(string)
		fixBindingID         = fixReq.BindingID
		fixBindingSecretName = "sb-secret-name"
		fixAPICreds          = fixAPIPackageCredential()
	)

	appFinder := &automock.AppFinder{}
	defer appFinder.AssertExpectations(t)
	appFinder.On("FindOneByServiceID", fixApp.Services[0].ID).
		Return(&fixApp, nil).
		Once()

	apiCredsMock := &automock.APIPackageCredentialsGetter{}
	defer apiCredsMock.AssertExpectations(t)
	apiCredsMock.On("GetAPIPackageCredentials", ctx, fixApp.CompassMetadata.ApplicationID, string(fixApp.Services[0].ID), fixReq.InstanceID).
		Return(fixAPICreds, nil).
		Once()

	sbFetcherMock := &automock.ServiceBindingFetcher{}
	sbFetcherMock.On("GetServiceBindingSecretName", fixNamespace, fixBindingID).Return(fixBindingSecretName, nil)
	osbCtx := broker.NewOSBContext("not", "important", "")
	svc := broker.NewBindServiceV2(appFinder, apiCredsMock, "http://%s-gateway", sbFetcherMock)

	// when
	resp, err := svc.Bind(ctx, *osbCtx, fixReq)

	// then
	require.NoError(t, err)
	require.NotNil(t, resp.Credentials)

	assert.Equal(t, "http://system-gateway/secret/sb-secret-name/api/API_MOCK_PROMOTIONS_123", resp.Credentials["API_MOCK_PROMOTIONS_123_GATEWAY_URL"])
	assert.Equal(t, "http://promotions.target.io", resp.Credentials["API_MOCK_PROMOTIONS_123_TARGET_URL"])

	assert.Equal(t, "http://system-gateway/secret/sb-secret-name/api/API_MOCK_OCC_456", resp.Credentials["API_MOCK_OCC_456_GATEWAY_URL"])
	assert.Equal(t, "http://occ.target.io", resp.Credentials["API_MOCK_OCC_456_TARGET_URL"])

	assert.Equal(t, fixAPICreds.Type, resp.Credentials["CREDENTIALS_TYPE"])

	gotConfigJSON, err := json.Marshal(resp.Credentials["CONFIGURATION"])
	require.NoError(t, err)
	assert.JSONEq(t, `{
					"requestParameters": {
						"headers": {
						  "h- param-1": [
							"val1",
							"val2"
						  ]
						},
						"queryParameters": {
						  "q-param-1": [
							"val1",
							"val2"
						  ]
						}
					  },
					  "csrfConfig": {
						"tokenUrl": "http://token-url.io"
					  },
					  "credentials": {
						"username": "user",
						"password": "pass"
					  }
					}`, string(gotConfigJSON))
}

func TestBindServiceBindFailureV2(t *testing.T) {
	t.Run("On credentials get error", func(t *testing.T) {
		// given
		fixApp := fixApplication()
		fixInstanceID := fixBindRequest().InstanceID
		appSvcID := internal.ApplicationServiceID("wrong-id")

		svc := broker.NewBindServiceV2(nil, nil, "", nil)

		// when
		resp, err := svc.GetCredentials(context.Background(), "", appSvcID, fixInstanceID, "", &fixApp)

		// then
		assert.EqualError(t, err, fmt.Sprintf("cannot get credentials to bind instance with ApplicationServiceID: %s, from Application: %s", appSvcID, fixApp.Name))
		assert.Zero(t, resp)
	})

	t.Run("On unexpected req params", func(t *testing.T) {
		//given
		fixReq := fixBindRequest()
		fixReq.Parameters = map[string]interface{}{
			"some-key": "some-value",
		}

		svc := broker.NewBindServiceV2(nil, nil, "", nil)
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
		BindingID:  "binding-id",
		InstanceID: "instance-id",
		ServiceID:  "123",
		PlanID:     "plan-id",
		Context: map[string]interface{}{
			"namespace": "system",
		},
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
							ID:        "123",
							Name:      "api_mock_promotions",
							TargetURL: "http://promotions.target.io",
						},
					},
					{
						Type: "API",
						APIEntry: &internal.APIEntry{
							ID:        "456",
							Name:      "api_mock_occ",
							TargetURL: "http://occ.target.io",
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

func fixAPIPackageCredential() internal.APIPackageCredential {
	return internal.APIPackageCredential{
		ID:   "123",
		Type: proxyconfig.Basic,
		Config: proxyconfig.Configuration{
			RequestParameters: &authorization.RequestParameters{
				Headers: &map[string][]string{
					"h- param-1": {"val1", "val2"},
				},
				QueryParameters: &map[string][]string{
					"q-param-1": {"val1", "val2"},
				},
			},
			CSRFConfig: &proxyconfig.CSRFConfig{
				TokenURL: "http://token-url.io",
			},
			Credentials: proxyconfig.BasicAuthConfig{
				Username: "user",
				Password: "pass",
			},
		},
	}
}
