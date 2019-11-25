// +build acceptance

package k8s

import (
	"os"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
)

var AuthSuite *auth.TestSuite

func TestMain(m *testing.M) {
	AuthSuite = auth.New()

	code := m.Run()

	os.Exit(code)
}
