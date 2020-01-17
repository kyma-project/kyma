package backuptest

import (
	"testing"

)

func TestBeforeBackup(t *testing.T) {
	testBackup(t, testBeforeBackup)
}

