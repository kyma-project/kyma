package module

import (
	"log"
	"os"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/graphql"
)

type BackendModule struct {
	Name string
}

type backendModuleQueryResponse struct {
	BackendModules []BackendModule
}

func IsEnabled(moduleName string, c *graphql.Client) (bool, error) {
	env := os.Getenv("MODULE_PLUGGABILITY")
	if env == "" || env == "false" {
		return true, nil
	}

	log.Println("Module pluggability enabled. Querying BackendModule custom resources...")
	response, err := queryBackendModules(c)
	if err != nil {
		return false, err
	}

	for _, module := range response.BackendModules {
		if module.Name == moduleName {
			return true, nil
		}
	}

	return false, nil
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
