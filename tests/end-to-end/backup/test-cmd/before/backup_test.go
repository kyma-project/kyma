package before

import (
	"testing"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/common"
)

func TestBeforeBackup(t *testing.T) {
	common.RunTest(t, common.TestBeforeBackup)
}
