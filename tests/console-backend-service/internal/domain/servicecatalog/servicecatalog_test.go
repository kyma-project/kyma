// +build acceptance

package servicecatalog

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/setup"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/dex"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/module"
)

func TestMain(m *testing.M) {
	dex.ExitIfSCIEnabled()

	c, err := graphql.New()
	exitOnError(err, "while GraphQL client setup")

	module.SkipPluggableMainIfShould(c, ModuleName)

	scInstaller, err := setup.NewServiceCatalogConfigurer(TestNamespace)
	exitOnError(err, fmt.Sprintf("while initializing Service Catalog installer for module %s", ModuleName))

	err = scInstaller.Setup()
	if err != nil {
		cleanupErr := scInstaller.Cleanup()
		log.Printf("Error while cleanup after failed setup for %s: %s", ModuleName, cleanupErr.Error())
		exitOnError(err, fmt.Sprintf("while setup for module %s", ModuleName))
	}

	code := m.Run()

	cleanupErr := scInstaller.Cleanup()
	log.Printf("Error while cleanup for %s: %s", ModuleName, cleanupErr.Error())
	os.Exit(code)
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatal(errors.Wrap(err, context))
	}
}
