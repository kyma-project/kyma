package testkit

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type K8sResourcesClient interface {
	GetDeployment(name string, options v1.GetOptions) (interface{}, error)
	GetService(name string, options v1.GetOptions) (interface{}, error)
	GetIngress(name string, options v1.GetOptions) (interface{}, error)
	GetRole(name string, options v1.GetOptions) (interface{}, error)
	GetRoleBinding(name string, options v1.GetOptions) (interface{}, error)
}

type k8sResourcesClient struct {
	coreClient        *kubernetes.Clientset
	applicationClient *versioned.Clientset
	namespace         string
}

func NewK8sResourcesClient(namespace string) (K8sResourcesClient, error) {
	kubeconfig := os.Getenv("KUBECONFIG")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

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
