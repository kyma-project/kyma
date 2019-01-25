package testkit

import (
	"github.com/kyma-project/kyma/components/connection-token-handler/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TokenRequestClient interface {
	CreateTokenRequest(appName string) (interface{}, error)
	GetTokenRequest(appName string) (interface{}, error)
}

type tokenRequestClient struct {
	tokenRequestClient *versioned.Clientset
	namespace string
}

func (t *tokenRequestClient) CreateTokenRequest(appName string) (interface{}, error) {
	tokenRequest := &v1alpha1.TokenRequest{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: t.namespace},
	}

	return t.tokenRequestClient.ApplicationconnectorV1alpha1().TokenRequests(t.namespace).Create(tokenRequest)
}