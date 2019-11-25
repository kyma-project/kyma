// +build acceptance

package ui

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/waiter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	readyTimeout = time.Minute
)

type BackendModule struct {
	Name string
}

type backendModuleQueryResponse struct {
	BackendModules []BackendModule
}

func TestBackendModule(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	uiCli, _, err := client.NewUIClientWithConfig()
	require.NoError(t, err)

	t.Run("QueryBackendModules", func(t *testing.T) {
		moduleNames := []string{"foo", "bar"}

		err = createBackendModules(moduleNames, uiCli)
		require.NoError(t, err)

		err = waiter.WaitAtMost(func() (bool, error) {
			resp, err := queryBackendModules(c)
			if err != nil {
				return false, err
			}

			return checkBackendModulesExist(moduleNames, resp.BackendModules)

		}, readyTimeout)
		assert.NoError(t, err)

		err = deleteBackendModules(moduleNames, uiCli)
		require.NoError(t, err)

		t.Log("Checking authorization directives...")
		as := auth.New()
		ops := &auth.OperationsInput{
			auth.List: {fixBackendModulesRequest()},
		}
		as.Run(t, ops)
	})
}

func createBackendModules(moduleNames []string, uiCli *versioned.Clientset) error {
	for _, moduleName := range moduleNames {
		resource := &v1alpha1.BackendModule{
			ObjectMeta: v1.ObjectMeta{
				Name: moduleName,
			},
		}
		_, err := uiCli.UiV1alpha1().BackendModules().Create(resource)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteBackendModules(moduleNames []string, uiCli *versioned.Clientset) error {
	for _, moduleName := range moduleNames {
		err := uiCli.UiV1alpha1().BackendModules().Delete(moduleName, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func fixBackendModulesRequest() *graphql.Request {
	query := `
			query {
				backendModules {
					name
				}
			}	
		`
	req := graphql.NewRequest(query)

	return req
}

func queryBackendModules(c *graphql.Client) (backendModuleQueryResponse, error) {
	req := fixBackendModulesRequest()

	var res backendModuleQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func checkBackendModulesExist(expectedNames []string, modules []BackendModule) (bool, error) {
	for _, name := range expectedNames {
		if !contains(modules, BackendModule{Name: name}) {
			return false, fmt.Errorf("BackendModule %s doesn't exist", name)
		}
	}

	return true, nil
}

func contains(modules []BackendModule, module BackendModule) bool {
	for _, v := range modules {
		if v.Name == module.Name {
			return true
		}
	}

	return false
}
