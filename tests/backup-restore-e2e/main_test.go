package main

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/backup-restore-e2e/framework"
)

// Entrypoint for testing
func TestMain(m *testing.M) {
	formatter := &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
	logrus.SetFormatter(formatter)
	logrus.Info("Starting tests")
	framework.MainEntry(m)

}
