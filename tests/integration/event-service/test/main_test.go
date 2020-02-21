package test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	istioSidecarStartWaitTime      = 60 * time.Second
	istioSidecarStartCheckInterval = 2 * time.Second
)

var (
	config TestConfig
)

func TestMain(m *testing.M) {
	var err error

	config, err = ReadConfig()
	if err != nil {
		logrus.Errorf("Failed to read configuration: %s", err.Error())
		os.Exit(1)
	}

	err = waitForIstioSidecar()
	if err != nil {
		logrus.Errorf("Error waiting for Istio sidecar to start: %s", err.Error())
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func waitForIstioSidecar() error {
	return WaitForFunction(istioSidecarStartCheckInterval, istioSidecarStartWaitTime, func() bool {
		url := config.EventServiceUrl + "/v1/health"

		response, err := http.Get(url)
		if err != nil {
			logrus.Warnf("Failed to access health endpoint while waiting for sidecar to start. Retrying in %s: %s",
				istioSidecarStartCheckInterval.String(), err.Error())
			return false
		}
		defer func() {
			err := response.Body.Close()
			if err != nil {
				logrus.Warnf("Failed to close response body: %s", err.Error())
			}
		}()

		if response.StatusCode != http.StatusOK {
			logrus.Warnf("Received invalid response while waiting for sidecar to start. Retrying in %s: Status %s",
				istioSidecarStartCheckInterval.String(), response.Status)
			return false
		}

		logrus.Info("Successfully accessed health endpoint")
		return true
	})

}
