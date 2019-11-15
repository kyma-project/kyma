// +build acceptance

package cms

import (
	"log"
	"os"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/module"
	"github.com/pkg/errors"
)

var AuthSuite *auth.TestSuite

func TestMain(m *testing.M) {

	c, err := graphql.New()
	exitOnError(err, "while GraphQL client setup")

	module.SkipPluggableMainIfShould(c, ModuleName)

	AuthSuite = auth.New()

	code := m.Run()
	os.Exit(code)
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatal(errors.Wrap(err, context))
	}
}
