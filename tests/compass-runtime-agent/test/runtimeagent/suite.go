package runtimeagent

import (
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"path/filepath"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	restclient "k8s.io/client-go/rest"
)

const (
	defaultCheckInterval           = 2 * time.Second
	appGatewayHealthCheckTimeout   = 15 * time.Second
	accessServiceConnectionTimeout = 60 * time.Second
	apiServerAccessTimeout         = 60 * time.Second
	dnsWaitTime                    = 30 * time.Second
)

type TestSuite struct {
	k8sClient *kubernetes.Clientset
}

func NewTestSuite(config testkit.TestConfig) (*TestSuite, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		logrus.Info("Failed to read in cluster config, trying with local config")
		home := homedir.HomeDir()
		k8sConfPath := filepath.Join(home, ".kube", "config")
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", k8sConfPath)
		if err != nil {
			return nil, err
		}
	}

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return &TestSuite{
		k8sClient: k8sClient,
	}, nil
}

// WaitForAccessToAPIServer waits for access to API Server which might be delayed by initialization of Istio sidecar
func (ts *TestSuite) WaitForAccessToAPIServer() error {
	err := testkit.WaitForFunction(defaultCheckInterval, apiServerAccessTimeout, func() bool {
		logrus.Info("Trying to access API Server...")
		_, err := ts.k8sClient.ServerVersion()
		if err != nil {
			logrus.Errorf("Failed to access API Server: %s. Retrying in %s", err.Error(), defaultCheckInterval.String())
			return false
		}

		return true
	})

	return err
}
