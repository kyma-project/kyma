package backuptest

import "testing"

func TestAfterRestore(t *testing.T) {
	testBackup(t, testAfterRestore)
}