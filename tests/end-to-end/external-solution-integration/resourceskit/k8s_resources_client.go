package resourceskit

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	model "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
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
	CreateDeployment(deployment *model.Deployment) (interface{}, error)
	CreateService(service *core.Service) (interface{}, error)
	DeleteService(name string, options *v1.DeleteOptions) error
	DeleteDeployment(name string, options *v1.DeleteOptions) error
	GetNamespace() string
}

type k8sResourcesClient struct {
	coreClient        *kubernetes.Clientset
	applicationClient *versioned.Clientset
	namespace         string
}

func NewK8sResourcesClient(config *restclient.Config, namespace string) (K8sResourcesClient, error) {
	return initClient(config, namespace)
}

func initClient(k8sConfig *restclient.Config, namespace string) (K8sResourcesClient, error) {
	coreClientSet, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	applicationClientSet, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return &k8sResourcesClient{
		coreClient:        coreClientSet,
		applicationClient: applicationClientSet,
		namespace:         namespace,
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

func (c *k8sResourcesClient) CreateDeployment(deployment *model.Deployment) (interface{}, error) {
	return c.coreClient.AppsV1().Deployments(c.namespace).Create(deployment)
}

func (c *k8sResourcesClient) CreateService(service *core.Service) (interface{}, error) {
	return c.coreClient.CoreV1().Services(c.namespace).Create(service)
}

func (c *k8sResourcesClient) DeleteDeployment(name string, options *v1.DeleteOptions) error {
	return c.coreClient.AppsV1().Deployments(c.namespace).Delete(name, options)
}

func (c *k8sResourcesClient) DeleteService(name string, options *v1.DeleteOptions) error {
	return c.coreClient.CoreV1().Services(c.namespace).Delete(name, options)
}

func (c *k8sResourcesClient) GetNamespace() string {
	return c.namespace
}
