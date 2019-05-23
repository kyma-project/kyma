package resourceskit

import (
	"time"

	acv1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	verac "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	trv1 "github.com/kyma-project/kyma/components/connection-token-handler/pkg/apis/applicationconnector/v1alpha1"
	vertr "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type AppConnectorClient interface {
	CreateTokenRequest() (*trv1.TokenRequest, error)
	GetTokenRequest() (*trv1.TokenRequest, error)
	DeleteTokenRequest() error
	CreateDummyApplication(skipInstallation bool) (*acv1.Application, error)
	DeleteApplication() error
	GetApplication() (*acv1.Application, error)
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

func (c *appConnectorClient) CreateTokenRequest() (*trv1.TokenRequest, error) {
	tokenRequest := &trv1.TokenRequest{
		TypeMeta:   v1.TypeMeta{Kind: "Application", APIVersion: trv1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: consts.AppName, Namespace: c.namespace},
		Status: trv1.TokenRequestStatus{
			ExpireAfter: v1.NewTime(time.Now().Add(5 * time.Minute)),
		},
	}

	return c.tokenClient.ApplicationconnectorV1alpha1().TokenRequests(c.namespace).Create(tokenRequest)
}

func (c *appConnectorClient) GetTokenRequest() (*trv1.TokenRequest, error) {
	return c.tokenClient.ApplicationconnectorV1alpha1().TokenRequests(c.namespace).Get(consts.AppName, v1.GetOptions{})
}

func (c *appConnectorClient) DeleteTokenRequest() error {
	return c.tokenClient.ApplicationconnectorV1alpha1().TokenRequests(c.namespace).Delete(consts.AppName, &v1.DeleteOptions{})
}

func (c *appConnectorClient) CreateDummyApplication(skipInstallation bool) (*acv1.Application, error) {
	spec := acv1.ApplicationSpec{
		Services:         []acv1.Service{},
		AccessLabel:      consts.AccessLabel,
		SkipInstallation: skipInstallation,
	}

	dummyApp := &acv1.Application{
		TypeMeta:   v1.TypeMeta{Kind: "Application", APIVersion: acv1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: consts.AppName, Namespace: c.namespace},
		Spec:       spec,
	}

	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Create(dummyApp)
}

func (c *appConnectorClient) DeleteApplication() error {
	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Delete(consts.AppName, &v1.DeleteOptions{})
}

func (c *appConnectorClient) GetApplication() (*acv1.Application, error) {
	return c.applicationClient.ApplicationconnectorV1alpha1().Applications().Get(consts.AppName, v1.GetOptions{})
}
