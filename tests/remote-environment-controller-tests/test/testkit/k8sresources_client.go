package testkit

import (
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	v1extensions "k8s.io/api/extensions/v1beta1"
	v1rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type K8sResourcesClient interface {
	GetDeployment(name string, options v1.GetOptions) (*v1apps.Deployment, error)
	GetService(name string, options v1.GetOptions) (*v1core.Service, error)
	GetIngress(name string, options v1.GetOptions) (*v1extensions.Ingress, error)
	GetRole(name string, options v1.GetOptions) (*v1rbac.Role, error)
	GetRoleBinding(name string, options v1.GetOptions) (*v1rbac.RoleBinding, error)
	CreateDummyRemoteEnvironment(name string) (*v1alpha1.RemoteEnvironment, error)
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

func (c *k8sResourcesClient) GetDeployment(name string, options v1.GetOptions) (*v1apps.Deployment, error) {
	return c.coreClient.AppsV1().Deployments(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetIngress(name string, options v1.GetOptions) (*v1extensions.Ingress, error) {
	return c.coreClient.ExtensionsV1beta1().Ingresses(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetRole(name string, options v1.GetOptions) (*v1rbac.Role, error) {
	return c.coreClient.RbacV1().Roles(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetRoleBinding(name string, options v1.GetOptions) (*v1rbac.RoleBinding, error) {
	return c.coreClient.RbacV1().RoleBindings(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetService(name string, options v1.GetOptions) (*v1core.Service, error) {
	return c.coreClient.CoreV1().Services(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) CreateDummyRemoteEnvironment(name string) (*v1alpha1.RemoteEnvironment, error) {
	dummyRe := &v1alpha1.RemoteEnvironment{
		TypeMeta:   v1.TypeMeta{Kind: "RemoteEnvironment", APIVersion: v1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: name, Namespace: c.namespace},
		Spec: v1alpha1.RemoteEnvironmentSpec{
			Services: []v1alpha1.Service{},
		},
	}

	return c.remoteEnvironmentClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Create(dummyRe)
}

func (c *k8sResourcesClient) DeleteRemoteEnvironment(name string, options *v1.DeleteOptions) error {
	return c.remoteEnvironmentClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Delete(name, options)
}

func (c *k8sResourcesClient) GetRemoteEnvironment(name string, options v1.GetOptions) (*v1alpha1.RemoteEnvironment, error) {
	return c.remoteEnvironmentClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Get(name, options)
}
