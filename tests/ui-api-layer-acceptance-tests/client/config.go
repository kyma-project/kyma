package client

import (
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewRestClientConfig(kubeconfigPath string) (*rest.Config, error) {
	var config *rest.Config
	var err error
	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}
	return config, nil
}

func NewRestClientConfigFromEnv() (*rest.Config, error) {
	kubeConfigPath := os.Getenv("KUBE_CONFIG")
	return NewRestClientConfig(kubeConfigPath)
}
