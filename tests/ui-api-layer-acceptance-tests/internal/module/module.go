package module

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func SkipPluggableTestIfShould(t *testing.T, c *graphql.Client, moduleName string) {
	checkIfTestShouldBeSkipped(true, c, moduleName, func(err error) {
		require.NoError(t, err)
	}, func(skipMessage string) {
		t.Skip(skipMessage)
	})
}

func SkipPluggableMainIfShould(c *graphql.Client, moduleName string) {
	checkIfTestShouldBeSkipped(true, c, moduleName, func(err error) {
		finalErr := errors.Wrapf(err, "while checking if module %s is enabled", moduleName)
		log.Fatal(finalErr)
	}, func(skipMessage string) {
		log.Println(skipMessage)
		os.Exit(0)
	})
}

func checkIfTestShouldBeSkipped(pluggable bool, c *graphql.Client, moduleName string, onError func(error), onSkip func(string)) {
	if !pluggable {
		return
	}

	enabled, err := IsEnabled(moduleName, c)
	if err != nil {
		onError(err)
	}

	if !enabled {
		onSkip(fmt.Sprintf("Module %s is disabled", moduleName))
	}
}
