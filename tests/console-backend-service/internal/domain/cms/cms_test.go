// +build acceptance

package cms

import (
	"log"
	"os"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/dex"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/module"
	"github.com/pkg/errors"
)

func TestMain(m *testing.M) {
	dex.ExitIfSCIEnabled()

	c, err := graphql.New()
	exitOnError(err, "while GraphQL client setup")

	module.SkipPluggableMainIfShould(c, ModuleName)

	code := m.Run()
	os.Exit(code)
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatal(errors.Wrap(err, context))
	}
}
