package testkit

import (
	"github.com/kyma-project/kyma/components/metadata-service/pkg/apis/istio/v1alpha2"
	istioclient "github.com/kyma-project/kyma/components/metadata-service/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type K8sResourcesClient interface {
	GetService(name string, options v1.GetOptions) (*v1core.Service, error)
	GetSecret(name string, options v1.GetOptions) (*v1core.Secret, error)
	GetDenier(name string, options v1.GetOptions) (*v1alpha2.Denier, error)
	GetRule(name string, options v1.GetOptions) (*v1alpha2.Rule, error)
	GetChecknothing(name string, options v1.GetOptions) (*v1alpha2.Checknothing, error)
	GetRemoteEnvironmentServices(name string, options v1.GetOptions) (*v1alpha1.RemoteEnvironment, error)
	CreateDummyRemoteEnvironment(name string, options v1.GetOptions) (*v1alpha1.RemoteEnvironment, error)
	DeleteRemoteEnvironment(name string, options *v1.DeleteOptions) error
}

type k8sResourcesClient struct {
	coreClient              *kubernetes.Clientset
	istioClient             *istioclient.Clientset
	remoteEnvironmentClient *versioned.Clientset
	namespace               string
	remoteEnvironmentName   string
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

	remoteEnvironmentClientset, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	istioClientset, err := istioclient.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return &k8sResourcesClient{
		coreClient:              coreClientset,
		istioClient:             istioClientset,
		remoteEnvironmentClient: remoteEnvironmentClientset,
		namespace:               namespace,
	}, nil
}

func (c *k8sResourcesClient) GetService(name string, options v1.GetOptions) (*v1core.Service, error) {
	return c.coreClient.CoreV1().Services(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetSecret(name string, options v1.GetOptions) (*v1core.Secret, error) {
	return c.coreClient.CoreV1().Secrets(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetDenier(name string, options v1.GetOptions) (*v1alpha2.Denier, error) {
	return c.istioClient.IstioV1alpha2().Deniers(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetRule(name string, options v1.GetOptions) (*v1alpha2.Rule, error) {
	return c.istioClient.IstioV1alpha2().Rules(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetChecknothing(name string, options v1.GetOptions) (*v1alpha2.Checknothing, error) {
	return c.istioClient.IstioV1alpha2().Checknothings(c.namespace).Get(name, options)
}

func (c *k8sResourcesClient) GetRemoteEnvironmentServices(name string, options v1.GetOptions) (*v1alpha1.RemoteEnvironment, error) {
	return c.remoteEnvironmentClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Get(name, options)
}

func (c *k8sResourcesClient) CreateDummyRemoteEnvironment(name string, options v1.GetOptions) (*v1alpha1.RemoteEnvironment, error) {
	dummyRe := &v1alpha1.RemoteEnvironment{
		TypeMeta:   v1.TypeMeta{Kind: "RemoteEnvironment", APIVersion: v1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: name},
		Spec: v1alpha1.RemoteEnvironmentSpec{
			Services: []v1alpha1.Service{},
		},
	}

	return c.remoteEnvironmentClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Create(dummyRe)
}

func (c *k8sResourcesClient) DeleteRemoteEnvironment(name string, options *v1.DeleteOptions) error {
	return c.remoteEnvironmentClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Delete(name, options)
}
