// +build acceptance

package servicecatalogaddons

import (
	"log"
	"os"
	"testing"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/domain/shared/setup"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/graphql"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/module"
	"github.com/pkg/errors"
)

func TestMain(m *testing.M) {
	dex.ExitIfSCIEnabled()

	c, err := graphql.New()
	exitOnError(err, "while GraphQL client setup")

	module.SkipPluggableMainIfShould(c, ModuleName)

	scInstaller, err := setup.NewServiceCatalogInstaller(TestNamespace)
	exitOnError(err, "while initializing Service Catalog installer")

	err = scInstaller.Setup()
	if err != nil {
		scInstaller.Cleanup()
		exitOnError(err, "while setup")
	}

	code := m.Run()

	scInstaller.Cleanup()
	os.Exit(code)
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatal(errors.Wrap(err, context))
	}
}
