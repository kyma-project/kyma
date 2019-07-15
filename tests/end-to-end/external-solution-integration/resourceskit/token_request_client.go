package resourceskit

import (
	"time"

	connectionTokenHandlerApi "github.com/kyma-project/kyma/components/connection-token-handler/pkg/apis/applicationconnector/v1alpha1"
	connectionTokenHandlerClient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TokenRequestClient interface {
	CreateTokenRequest() (*connectionTokenHandlerApi.TokenRequest, error)
	GetTokenRequest() (*connectionTokenHandlerApi.TokenRequest, error)
	DeleteTokenRequest() error
}

type tokenRequestClient struct {
	tokenRequests connectionTokenHandlerClient.TokenRequestInterface
}

func NewTokenRequestClient(tokenRequests connectionTokenHandlerClient.TokenRequestInterface) TokenRequestClient {
	return &tokenRequestClient{tokenRequests: tokenRequests}
}

func (t *tokenRequestClient) CreateTokenRequest() (*connectionTokenHandlerApi.TokenRequest, error) {
	tokenRequest := &connectionTokenHandlerApi.TokenRequest{
		ObjectMeta: metav1.ObjectMeta{Name: consts.AppName},
		Status: connectionTokenHandlerApi.TokenRequestStatus{
			ExpireAfter: metav1.NewTime(time.Now().Add(5 * time.Minute)),
		},
	}

	return t.tokenRequests.Create(tokenRequest)
}

func (t *tokenRequestClient) GetTokenRequest() (*connectionTokenHandlerApi.TokenRequest, error) {
	return t.tokenRequests.Get(consts.AppName, metav1.GetOptions{})
}

func (t *tokenRequestClient) DeleteTokenRequest() error {
	return t.tokenRequests.Delete(consts.AppName, &metav1.DeleteOptions{})
}
