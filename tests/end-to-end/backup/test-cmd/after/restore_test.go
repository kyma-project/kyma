package after

import (
	"testing"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/common"
)

func TestAfterRestore(t *testing.T) {
	common.RunTest(t, common.TestAfterRestore)
}
