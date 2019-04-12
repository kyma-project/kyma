package testkit

import (
	acv1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	verac "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	trv1 "github.com/kyma-project/kyma/components/connection-token-handler/pkg/apis/applicationconnector/v1alpha1"
	vertr "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"time"
)

type AppConnectorClient interface {
	CreateTokenRequest(appName string) (*trv1.TokenRequest, error)
	GetTokenRequest(appName string, options v1.GetOptions) (*trv1.TokenRequest, error)
	DeleteTokenRequest(appName string, options *v1.DeleteOptions) error
	CreateDummyApplication(name string, accessLabel string, skipInstallation bool) (*acv1.Application, error)
	DeleteApplication(name string, options *v1.DeleteOptions) error
	GetApplication(name string, options v1.GetOptions) (*acv1.Application, error)
}

type appConnectorClient struct {
	tokenClient       *vertr.Clientset
	applicationClient *verac.Clientset
	namespace         string
}

func NewAppConnectorClient(config *rest.Config, namespace string) (AppConnectorClient, error) {
	tokenClientSet, err := vertr.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	applicationClientSet, err := verac.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &appConnectorClient{tokenClient: tokenClientSet, applicationClient: applicationClientSet, namespace: namespace}, nil
}

func (c *appConnectorClient) CreateTokenRequest(appName string) (*trv1.TokenRequest, error) {
	tokenRequest := &trv1.TokenRequest{
		TypeMeta:   v1.TypeMeta{Kind: "Application", APIVersion: trv1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: appName, Namespace: c.namespace},
		Status: trv1.TokenRequestStatus{
			ExpireAfter: v1.NewTime(time.Now().Add(5 * time.Minute)),
		},
	}

	return c.tokenClient.ApplicationconnectorV1alpha1().TokenRequests(c.namespace).Create(tokenRequest)
}

func (c *appConnectorClient) GetTokenRequest(appName string, options v1.GetOptions) (*trv1.TokenRequest, error) {
	return c.tokenClient.ApplicationconnectorV1alpha1().TokenRequests(c.namespace).Get(appName, options)
}

func (c *appConnectorClient) DeleteTokenRequest(appName string, options *v1.DeleteOptions) error {
	return c.tokenClient.ApplicationconnectorV1alpha1().TokenRequests(c.namespace).Delete(appName, options)
}

func (c *appConnectorClient) CreateDummyApplication(name string, accessLabel string, skipInstallation bool) (*acv1.Application, error) {
	spec := acv1.ApplicationSpec{
		Services:         []acv1.Service{},
		AccessLabel:      accessLabel,
		SkipInstallation: skipInstallation,
	}

	dummyApp := &acv1.Application{
		TypeMeta:   v1.TypeMeta{Kind: "Application", APIVersion: acv1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: name, Namespace: c.namespace},
		Spec:       spec,
	}

	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Create(dummyApp)
}

func (c *appConnectorClient) DeleteApplication(name string, options *v1.DeleteOptions) error {
	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Delete(name, options)
}

func (c *appConnectorClient) GetApplication(name string, options v1.GetOptions) (*acv1.Application, error) {
	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Get(name, options)
}
