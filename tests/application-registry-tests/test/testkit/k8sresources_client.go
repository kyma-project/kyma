package testkit

import (
	"fmt"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-registry/pkg/apis/istio/v1alpha2"
	istioclient "github.com/kyma-project/kyma/components/application-registry/pkg/client/clientset/versioned"
	v1core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type K8sResourcesClient interface {
	GetService(name string, options v1.GetOptions) (*v1core.Service, error)
	GetSecret(name string, options v1.GetOptions) (*v1core.Secret, error)
	GetHandler(name string, options v1.GetOptions) (*v1alpha2.Handler, error)
	GetRule(name string, options v1.GetOptions) (*v1alpha2.Rule, error)
	GetInstance(name string, options v1.GetOptions) (*v1alpha2.Instance, error)
	GetApplicationServices(name string, options v1.GetOptions) (*v1alpha1.Application, error)
	CreateDummyApplication(namePrefix string, options v1.GetOptions, skipInstallation bool) (*v1alpha1.Application, error)
	DeleteApplication(name string, options *v1.DeleteOptions) error
}

type k8sResourcesClient struct {
	coreClient        *kubernetes.Clientset
	istioClient       *istioclient.Clientset
	applicationClient *versioned.Clientset
	namespace         string
	applicationName   string
}

func NewK8sInClusterResourcesClient(namespace string) (K8sResourcesClient, error) {
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

	istioClientset, err := istioclient.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return &k8sResourcesClient{
		coreClient:        coreClientset,
		istioClient:       istioClientset,
		applicationClient: applicationClientset,
		namespace:         namespace,
	}, nil
}

func (c *k8sResourcesClient) GetService(name string, options v1.GetOptions) (*v1core.Service, error) {
	return c.coreClient.CoreV1().Services(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetSecret(name string, options v1.GetOptions) (*v1core.Secret, error) {
	return c.coreClient.CoreV1().Secrets(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetHandler(name string, options v1.GetOptions) (*v1alpha2.Handler, error) {
	return c.istioClient.IstioV1alpha2().Handlers(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetRule(name string, options v1.GetOptions) (*v1alpha2.Rule, error) {
	return c.istioClient.IstioV1alpha2().Rules(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetInstance(name string, options v1.GetOptions) (*v1alpha2.Instance, error) {
	return c.istioClient.IstioV1alpha2().Instances(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetApplicationServices(name string, options v1.GetOptions) (*v1alpha1.Application, error) {
	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Get(name, options)
}

func (c *k8sResourcesClient) CreateDummyApplication(namePrefix string, options v1.GetOptions, skipInstallation bool) (*v1alpha1.Application, error) {
	spec := v1alpha1.ApplicationSpec{
		Services:         []v1alpha1.Service{},
		SkipInstallation: skipInstallation,
	}

	dummyAppName := addRandomPostfix(namePrefix)

	dummyApp := &v1alpha1.Application{
		TypeMeta:   v1.TypeMeta{Kind: "Application", APIVersion: v1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: dummyAppName},
		Spec:       spec,
	}

	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Create(dummyApp)
}

func addRandomPostfix(s string) string {
	return fmt.Sprintf(s+"-%s", rand.String(5))
}

func (c *k8sResourcesClient) DeleteApplication(name string, options *v1.DeleteOptions) error {
	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Delete(name, options)
}
