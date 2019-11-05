package testkit

import (
	"time"

	"github.com/kyma-project/kyma/components/connection-token-handler/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
)

const namespace = "default"

type K8sResourcesClient interface {
	GetTokenRequest(name string, options v1.GetOptions) (*v1alpha1.TokenRequest, error)
	CreateTokenRequest(name, group, tenant string) (*v1alpha1.TokenRequest, error)
	DeleteTokenRequest(name string, options *v1.DeleteOptions) error
}

type k8sResourcesClient struct {
	tokenRequestsClient *versioned.Clientset
}

func NewK8sResourcesClient() (K8sResourcesClient, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return initClient(k8sConfig)
}

func initClient(k8sConfig *restclient.Config) (K8sResourcesClient, error) {
	tokenRequestClientSet, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}
	return &k8sResourcesClient{
		tokenRequestsClient: tokenRequestClientSet,
	}, nil
}

func (c *k8sResourcesClient) CreateTokenRequest(name, group, tenant string) (*v1alpha1.TokenRequest, error) {
	tokenRequest := &v1alpha1.TokenRequest{
		TypeMeta:   v1.TypeMeta{Kind: "TokenRequest", APIVersion: v1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: name},
		Context:    v1alpha1.ClusterContext{Group: group, Tenant: tenant},
		//TODO CR should be created without passing status
		Status: v1alpha1.TokenRequestStatus{ExpireAfter: v1.Date(2999, time.December, 12, 12, 12, 12, 12, time.Local)},
	}

	return c.tokenRequestsClient.ApplicationconnectorV1alpha1().TokenRequests(namespace).Create(tokenRequest)
}

func (c *k8sResourcesClient) DeleteTokenRequest(name string, options *v1.DeleteOptions) error {
	return c.tokenRequestsClient.ApplicationconnectorV1alpha1().TokenRequests(namespace).Delete(name, options)
}

func (c *k8sResourcesClient) GetTokenRequest(name string, options v1.GetOptions) (*v1alpha1.TokenRequest, error) {
	return c.tokenRequestsClient.ApplicationconnectorV1alpha1().TokenRequests(namespace).Get(name, options)
}
