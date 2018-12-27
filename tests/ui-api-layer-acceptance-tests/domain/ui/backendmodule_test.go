package ui

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/client"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BackendModule struct {
	Name string
}

type backendModuleQueryResponse struct {
	BackendModules []BackendModule
}

func TestBackendModule(t *testing.T) {
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}

	c, err := graphql.New()
	require.NoError(t, err)

	uiCli, _, err := client.NewUIClientWithConfig()
	require.NoError(t, err)

	t.Run("QueryBackendModules", func(t *testing.T) {
		moduleNames := []string{"foo", "bar"}

		err = createBackendModules(moduleNames, uiCli)
		require.NoError(t, err)

		resp, err := queryBackendModules(c)
		assert.NoError(t, err)

		assertBackendModules(t, moduleNames, resp.BackendModules)

		err = deleteBackendModules(moduleNames, uiCli)
		require.NoError(t, err)
	})
}

func createBackendModules(moduleNames []string, uiCli *versioned.Clientset) error {
	for _, moduleName := range moduleNames {
		resource := &v1alpha1.BackendModule{
			ObjectMeta: v1.ObjectMeta{
				Name: moduleName,
			},
		}
		_, err := uiCli.UiV1alpha1().BackendModules("").Create(resource)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteBackendModules(moduleNames []string, uiCli *versioned.Clientset) error {
	for _, moduleName := range moduleNames {
		err := uiCli.UiV1alpha1().BackendModules("").Delete(moduleName, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func queryBackendModules(c *graphql.Client) (backendModuleQueryResponse, error) {
	query := `
			query {
				backendModules {
					name
				}
			}	
		`
	req := graphql.NewRequest(query)

	var res backendModuleQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func assertBackendModules(t *testing.T, expectedNames []string, actual []BackendModule) {
	for _, v := range expectedNames {
		assert.Contains(t, actual, BackendModule{Name: v})
	}
}
