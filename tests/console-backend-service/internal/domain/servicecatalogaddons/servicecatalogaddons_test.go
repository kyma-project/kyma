// +build acceptance

package servicecatalogaddons

import (
	"log"
	"os"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/dex"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/setup"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/exit"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/module"
)

var AuthSuite *auth.TestSuite

func TestMain(m *testing.M) {
	dex.ExitIfSCIEnabled()

	c, err := graphql.New()
	exit.OnError(err, "while GraphQL client setup for module %s", ModuleName)

	module.SkipPluggableMainIfShould(c, ModuleName)

	scInstaller, err := setup.NewServiceCatalogConfigurer(TestNamespace, false)
	exit.OnError(err, "while initializing Service Catalog Configurer for module %s", ModuleName)

	err = scInstaller.Setup()
	if err != nil {
		cleanupErr := scInstaller.Cleanup()
		log.Printf("Error while cleanup after failed setup for %s: %s", ModuleName, cleanupErr.Error())
		exit.OnError(err, "while setup for module %s", ModuleName)
	}

	AuthSuite = auth.New()

	code := m.Run()

	cleanupErr := scInstaller.Cleanup()
	if cleanupErr != nil {
		log.Printf("Error while cleanup for %s: %s", ModuleName, cleanupErr.Error())
	}
	os.Exit(code)
}
