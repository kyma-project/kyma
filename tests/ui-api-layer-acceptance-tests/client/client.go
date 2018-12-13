package client

import (
	scClientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	idpClientset "github.com/kyma-project/kyma/components/idppreset/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func NewClientWithConfig() (*v1.CoreV1Client, *rest.Config, error) {
	k8sConfig, err := NewRestClientConfigFromEnv()
	if err != nil {
		return nil, nil, errors.Wrap(err, "while creating new client with config")
	}

	k8sCli, err := v1.NewForConfig(k8sConfig)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while creating new client with config")
	}

	return k8sCli, k8sConfig, nil
}

func NewServiceCatalogClientWithConfig() (*scClientset.Clientset, *rest.Config, error) {
	k8sConfig, err := NewRestClientConfigFromEnv()
	if err != nil {
		return nil, nil, errors.Wrap(err, "while creating new client with config")
	}

	scCli, err := scClientset.NewForConfig(k8sConfig)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while creating new client with config")
	}

	return scCli, k8sConfig, nil
}

func NewIDPPresetClientWithConfig() (*idpClientset.Clientset, *rest.Config, error) {
	k8sConfig, err := NewRestClientConfigFromEnv()
	if err != nil {
		return nil, nil, errors.Wrap(err, "while creating new client with config")
	}

	idpCli, err := idpClientset.NewForConfig(k8sConfig)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while creating new client with config")
	}

	return idpCli, k8sConfig, nil
}
