package lookup

import (
	"bytes"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/graphql"
	"github.com/kyma-project/kyma/components/connector-service/internal/graphql/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestLookupService_Fetch(t *testing.T) {
	t.Run("Should get gatewayUrl from external service", func(t *testing.T) {
		//given
		service := &mocks.GraphQLService{}
		qlLookupService := NewGraphQLLookupService(service, "testdata/")
		appCtx := clientcontext.ClientContext{
			Group:  "exampleGroup",
			Tenant: "exampleTenant",
			ID:     "exampleApp",
		}

		expectedGatewayURL := "https://gateway.cool-cluster.cluster.extend.sap.cx"
		exampleResponse := `{
    "data": {
        "applications": [
            {
                "name": "example app",
                "account": {
                    "id": "boo"
                },
                "groups": [
                    {
                        "id": "bar",
                        "name": "cool-cluster",
                        "clusters": [
                            {
                                "id": "baz",
                                "name": "cool-cluster",
                                "endpoints": {
                                    "gateway": "https://gateway.cool-cluster.cluster.extend.sap.cx"
                                }
                            }
                        ]
                    }
                ]
            }
        ]
    }
}`
		body := ioutil.NopCloser(bytes.NewReader([]byte(exampleResponse)))
		response := &http.Response{Body: body}

		service.On("ReadConfig", mock.Anything).Return(graphql.Config{}, nil)
		service.On("SendRequest", mock.Anything, mock.Anything, mock.Anything).Return(response, nil)

		//when
		gatewayURL, e := qlLookupService.Fetch(appCtx)
		require.NoError(t, e)

		//then
		assert.Equal(t, expectedGatewayURL, gatewayURL)
	})
}
