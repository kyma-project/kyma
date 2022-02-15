package testkit

import (
	"context"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"

	istio "istio.io/client-go/pkg/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
)

type K8sResourcesClient interface {
	GetDeployment(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error)
	GetService(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error)
	GetVirtualService(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error)
	GetRole(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error)
	GetRoleBinding(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error)
	GetClusterRole(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error)
	GetClusterRoleBinding(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error)
	GetServiceAccount(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error)
	CreateDummyApplication(ctx context.Context, name string, accessLabel string, skipInstallation bool) (*v1alpha1.Application, error)
	DeleteApplication(ctx context.Context, name string, options metav1.DeleteOptions) error
	GetApplication(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.Application, error)
	ListPods(ctx context.Context, options metav1.ListOptions) (*corev1.PodList, error)
	DeletePod(ctx context.Context, name string, options metav1.DeleteOptions) error
	GetLogs(ctx context.Context, podName string, options *corev1.PodLogOptions) *restclient.Request
	CreateNamespace(ctx context.Context, namespace *corev1.Namespace) (*corev1.Namespace, error)
	DeleteNamespace(ctx context.Context) error
	CreateServiceInstance(ctx context.Context, serviceInstance *v1beta1.ServiceInstance) (*v1beta1.ServiceInstance, error)
	DeleteServiceInstance(ctx context.Context, name string) error
}

type k8sResourcesClient struct {
	coreClient            *kubernetes.Clientset
	applicationClient     *versioned.Clientset
	serviceInstanceClient clientset.Interface
	istioClient           *istio.Clientset
	namespace             string
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

	serviceInstanceClient, err := clientset.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	istioClientset, err := istio.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return &k8sResourcesClient{
		coreClient:            coreClientset,
		applicationClient:     applicationClientset,
		serviceInstanceClient: serviceInstanceClient,
		istioClient:           istioClientset,
		namespace:             namespace,
	}, nil
}

func (c *k8sResourcesClient) GetDeployment(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error) {
	return c.coreClient.AppsV1().Deployments(c.namespace).Get(ctx, name, options)
}

func (c *k8sResourcesClient) GetVirtualService(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error) {
	return c.istioClient.NetworkingV1alpha3().VirtualServices(c.namespace).Get(ctx, name, options)
}

func (c *k8sResourcesClient) GetRole(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error) {
	return c.coreClient.RbacV1().Roles(c.namespace).Get(ctx, name, options)
}

func (c *k8sResourcesClient) GetRoleBinding(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error) {
	return c.coreClient.RbacV1().RoleBindings(c.namespace).Get(ctx, name, options)
}

func (c *k8sResourcesClient) GetClusterRole(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error) {
	return c.coreClient.RbacV1().ClusterRoles().Get(ctx, name, options)
}

func (c *k8sResourcesClient) GetClusterRoleBinding(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error) {
	return c.coreClient.RbacV1().ClusterRoleBindings().Get(ctx, name, options)
}

func (c *k8sResourcesClient) GetServiceAccount(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error) {
	return c.coreClient.CoreV1().ServiceAccounts(c.namespace).Get(ctx, name, options)
}

func (c *k8sResourcesClient) GetService(ctx context.Context, name string, options metav1.GetOptions) (interface{}, error) {
	return c.coreClient.CoreV1().Services(c.namespace).Get(ctx, name, options)
}

func (c *k8sResourcesClient) CreateDummyApplication(ctx context.Context, name string, accessLabel string, skipInstallation bool) (*v1alpha1.Application, error) {
	spec := v1alpha1.ApplicationSpec{
		Services:         []v1alpha1.Service{},
		AccessLabel:      accessLabel,
		SkipInstallation: skipInstallation,
	}

	dummyApp := &v1alpha1.Application{
		TypeMeta:   metav1.TypeMeta{Kind: "Application", APIVersion: v1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: c.namespace},
		Spec:       spec,
	}

	options := metav1.CreateOptions{}

	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Create(ctx, dummyApp, options)
}

func (c *k8sResourcesClient) DeleteApplication(ctx context.Context, name string, options metav1.DeleteOptions) error {
	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Delete(ctx, name, options)
}

func (c *k8sResourcesClient) GetApplication(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.Application, error) {
	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Get(ctx, name, options)
}

func (c *k8sResourcesClient) ListPods(ctx context.Context, options metav1.ListOptions) (*corev1.PodList, error) {
	return c.coreClient.CoreV1().Pods(c.namespace).List(ctx, options)
}

func (c *k8sResourcesClient) DeletePod(ctx context.Context, name string, options metav1.DeleteOptions) error {
	return c.coreClient.CoreV1().Pods(c.namespace).Delete(ctx, name, options)
}

func (c *k8sResourcesClient) GetLogs(ctx context.Context, podName string, options *corev1.PodLogOptions) *restclient.Request {
	return c.coreClient.CoreV1().Pods(c.namespace).GetLogs(podName, options)
}

func (c *k8sResourcesClient) CreateNamespace(ctx context.Context, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	return c.coreClient.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
}

func (c *k8sResourcesClient) DeleteNamespace(ctx context.Context) error {
	return c.coreClient.CoreV1().Namespaces().Delete(ctx, c.namespace, metav1.DeleteOptions{})
}

func (c *k8sResourcesClient) CreateServiceInstance(ctx context.Context, serviceInstance *v1beta1.ServiceInstance) (*v1beta1.ServiceInstance, error) {
	return c.serviceInstanceClient.ServicecatalogV1beta1().ServiceInstances(c.namespace).Create(ctx, serviceInstance, metav1.CreateOptions{})
}

func (c *k8sResourcesClient) DeleteServiceInstance(ctx context.Context, name string) error {
	return c.serviceInstanceClient.ServicecatalogV1beta1().ServiceInstances(c.namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
