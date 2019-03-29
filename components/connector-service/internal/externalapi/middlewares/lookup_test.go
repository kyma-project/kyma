package middlewares

import (
	"github.com/stretchr/testify/assert"
	"testing"
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

		config, _ := readConfigFromFile("testdata/config.json")

		assert.Equal(t, expectedConfig, config)
	})
}
