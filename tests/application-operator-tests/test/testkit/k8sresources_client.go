package testkit

import (
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type K8sResourcesClient interface {
	GetDeployment(name string, options v1.GetOptions) (interface{}, error)
	GetService(name string, options v1.GetOptions) (interface{}, error)
	GetIngress(name string, options v1.GetOptions) (interface{}, error)
	GetRole(name string, options v1.GetOptions) (interface{}, error)
	GetRoleBinding(name string, options v1.GetOptions) (interface{}, error)
	CreateDummyRemoteEnvironment(name string, accessLabel string, skipInstallation bool) (*v1alpha1.RemoteEnvironment, error)
	DeleteRemoteEnvironment(name string, options *v1.DeleteOptions) error
	GetRemoteEnvironment(name string, options v1.GetOptions) (*v1alpha1.RemoteEnvironment, error)
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

func (c *k8sResourcesClient) GetDeployment(name string, options v1.GetOptions) (interface{}, error) {
	return c.coreClient.AppsV1().Deployments(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetIngress(name string, options v1.GetOptions) (interface{}, error) {
	return c.coreClient.ExtensionsV1beta1().Ingresses(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetRole(name string, options v1.GetOptions) (interface{}, error) {
	return c.coreClient.RbacV1().Roles(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetRoleBinding(name string, options v1.GetOptions) (interface{}, error) {
	return c.coreClient.RbacV1().RoleBindings(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetService(name string, options v1.GetOptions) (interface{}, error) {
	return c.coreClient.CoreV1().Services(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) CreateDummyRemoteEnvironment(name string, accessLabel string, skipInstallation bool) (*v1alpha1.RemoteEnvironment, error) {
	spec := v1alpha1.RemoteEnvironmentSpec{
		Services:    []v1alpha1.Service{},
		AccessLabel: accessLabel,
	}

	if skipInstallation {
		spec.SkipInstallation = true
	}

	dummyRe := &v1alpha1.RemoteEnvironment{
		TypeMeta:   v1.TypeMeta{Kind: "RemoteEnvironment", APIVersion: v1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: name, Namespace: c.namespace},
		Spec:       spec,
	}

	return c.remoteEnvironmentClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Create(dummyRe)
}

func (c *k8sResourcesClient) DeleteRemoteEnvironment(name string, options *v1.DeleteOptions) error {
	return c.remoteEnvironmentClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Delete(name, options)
}

func (c *k8sResourcesClient) GetRemoteEnvironment(name string, options v1.GetOptions) (*v1alpha1.RemoteEnvironment, error) {
	return c.remoteEnvironmentClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Get(name, options)
}
