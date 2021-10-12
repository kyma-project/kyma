package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphQLLookupService(t *testing.T) {
	t.Run("Should parse json file to config struct", func(t *testing.T) {
		//given
		headers := Headers{}
		headers["faros-xf-user"] = "connector-service"
		headers["faros-xf-groups"] = "kyma-admins"

		expectedConfig := LookUpConfig{
			URL:     "https://faros.test.graph.ql",
			Headers: headers,
		}

		config, _ := readConfig("testdata/config.json")

		assert.Equal(t, expectedConfig, config)
	})

	t.Run("Should get gatewayUrl from external service", func(t *testing.T) {
		//given
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
		//when
		gatewayURL := getGatewayUrl([]byte(exampleResponse))

		//then
		assert.Equal(t, expectedGatewayURL, gatewayURL.String())
	})
}
