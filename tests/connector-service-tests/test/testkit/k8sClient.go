package testkit

import (
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type K8sResourcesClient interface {
	CreateDummyRemoteEnvironment(name string, accessLabel string) (*v1alpha1.RemoteEnvironment, error)
	DeleteRemoteEnvironment(name string, options *v1.DeleteOptions) error
}

type k8sResourcesClient struct {
	coreClient              *kubernetes.Clientset
	remoteEnvironmentClient *versioned.Clientset
	namespace               string
}

func NewK8sResourcesClient(namespace string) (K8sResourcesClient, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return initClient(k8sConfig, namespace)
}

func initClient(k8sConfig *restclient.Config, namespace string) (K8sResourcesClient, error) {
	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	remoteEnvironmentClientset, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return &k8sResourcesClient{
		coreClient:              coreClientset,
		remoteEnvironmentClient: remoteEnvironmentClientset,
		namespace:               namespace,
	}, nil
}

func (c *k8sResourcesClient) CreateDummyRemoteEnvironment(name string, accessLabel string) (*v1alpha1.RemoteEnvironment, error) {
	dummyRe := &v1alpha1.RemoteEnvironment{
		TypeMeta:   v1.TypeMeta{Kind: "RemoteEnvironment", APIVersion: v1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: name, Namespace: c.namespace},
		Spec: v1alpha1.RemoteEnvironmentSpec{
			Services:    []v1alpha1.Service{},
			AccessLabel: accessLabel,
		},
	}

	return c.remoteEnvironmentClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Create(dummyRe)
}

func (c *k8sResourcesClient) DeleteRemoteEnvironment(name string, options *v1.DeleteOptions) error {
	return c.remoteEnvironmentClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Delete(name, options)
}
