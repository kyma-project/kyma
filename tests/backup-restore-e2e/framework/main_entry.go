package framework

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

// MainEntry for framework
func MainEntry(m *testing.M) {
	if err := setup(); err != nil {
		logrus.Errorf("fail to setup framework: %v", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := teardown(); err != nil {
		logrus.Errorf("fail to teardown framework: %v", err)
		os.Exit(1)
	}
	logrus.Info("Test finished")
	os.Exit(code)
}
