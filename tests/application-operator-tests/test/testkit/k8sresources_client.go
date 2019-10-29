package testkit

import (
	"github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned/typed/networking.istio.io/v1alpha3"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type K8sResourcesClient interface {
	GetDeployment(name string, options v1.GetOptions) (interface{}, error)
	GetService(name string, options v1.GetOptions) (interface{}, error)
	GetVirtualService(name string, options v1.GetOptions) (interface{}, error)
	GetRole(name string, options v1.GetOptions) (interface{}, error)
	GetRoleBinding(name string, options v1.GetOptions) (interface{}, error)
	GetClusterRole(name string, options v1.GetOptions) (interface{}, error)
	GetClusterRoleBinding(name string, options v1.GetOptions) (interface{}, error)
	GetServiceAccount(name string, options v1.GetOptions) (interface{}, error)
	CreateDummyApplication(name string, accessLabel string, skipInstallation bool) (*v1alpha1.Application, error)
	DeleteApplication(name string, options *v1.DeleteOptions) error
	GetApplication(name string, options v1.GetOptions) (*v1alpha1.Application, error)
	ListPods(options v1.ListOptions) (*corev1.PodList, error)
	DeletePod(name string, options *v1.DeleteOptions) error
	GetLogs(podName string, options *corev1.PodLogOptions) *restclient.Request
}

type k8sResourcesClient struct {
	coreClient        *kubernetes.Clientset
	applicationClient *versioned.Clientset
	istioClient       *v1alpha3.NetworkingV1alpha3Client
	namespace         string
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

	applicationClientset, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	istioClientset, err := v1alpha3.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return &k8sResourcesClient{
		coreClient:        coreClientset,
		applicationClient: applicationClientset,
		istioClient:       istioClientset,
		namespace:         namespace,
	}, nil
}

func (c *k8sResourcesClient) GetDeployment(name string, options v1.GetOptions) (interface{}, error) {
	return c.coreClient.AppsV1().Deployments(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetVirtualService(name string, options v1.GetOptions) (interface{}, error) {
	return c.istioClient.VirtualServices(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetRole(name string, options v1.GetOptions) (interface{}, error) {
	return c.coreClient.RbacV1().Roles(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetRoleBinding(name string, options v1.GetOptions) (interface{}, error) {
	return c.coreClient.RbacV1().RoleBindings(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetClusterRole(name string, options v1.GetOptions) (interface{}, error) {
	return c.coreClient.RbacV1().ClusterRoles().Get(name, options)
}

func (c *k8sResourcesClient) GetClusterRoleBinding(name string, options v1.GetOptions) (interface{}, error) {
	return c.coreClient.RbacV1().ClusterRoleBindings().Get(name, options)
}

func (c *k8sResourcesClient) GetServiceAccount(name string, options v1.GetOptions) (interface{}, error) {
	return c.coreClient.CoreV1().ServiceAccounts(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetService(name string, options v1.GetOptions) (interface{}, error) {
	return c.coreClient.CoreV1().Services(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) CreateDummyApplication(name string, accessLabel string, skipInstallation bool) (*v1alpha1.Application, error) {
	spec := v1alpha1.ApplicationSpec{
		Services:         []v1alpha1.Service{},
		AccessLabel:      accessLabel,
		SkipInstallation: skipInstallation,
	}

	dummyApp := &v1alpha1.Application{
		TypeMeta:   v1.TypeMeta{Kind: "Application", APIVersion: v1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: name, Namespace: c.namespace},
		Spec:       spec,
	}

	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Create(dummyApp)
}

func (c *k8sResourcesClient) DeleteApplication(name string, options *v1.DeleteOptions) error {
	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Delete(name, options)
}

func (c *k8sResourcesClient) GetApplication(name string, options v1.GetOptions) (*v1alpha1.Application, error) {
	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Get(name, options)
}

func (c *k8sResourcesClient) ListPods(options v1.ListOptions) (*corev1.PodList, error) {
	return c.coreClient.CoreV1().Pods(c.namespace).List(options)
}

func (c *k8sResourcesClient) DeletePod(name string, options *v1.DeleteOptions) error {
	return c.coreClient.CoreV1().Pods(c.namespace).Delete(name, options)
}

func (c *k8sResourcesClient) GetLogs(podName string, options *corev1.PodLogOptions) *restclient.Request {
	return c.coreClient.CoreV1().Pods(c.namespace).GetLogs(podName, options)
}
