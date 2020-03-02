package proxyconfig

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/proxyconfig"
	v12 "k8s.io/api/core/v1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type SecretsCreator struct {
	secretsClient v1.SecretInterface
	namespace     string
}

func NewSecretsCreator(namespace string, secretsClient v1.SecretInterface) *SecretsCreator {
	return &SecretsCreator{
		secretsClient: secretsClient,
		namespace:     namespace,
	}
}

func (sc *SecretsCreator) NewSecret(secretName, apiName string, proxyConfig proxyconfig.ProxyDestinationConfig) error {
	destinationConfig, err := json.Marshal(proxyConfig.Configuration)
	if err != nil {
		return fmt.Errorf("failed to marshal destination config: %s", err.Error())
	}

	credentialsType := getCredentialsType(proxyConfig)

	prefix := newPrefixFunc(strings.ToUpper(apiName))

	secret := &v12.Secret{
		ObjectMeta: k8sMeta.ObjectMeta{Name: secretName, Namespace: sc.namespace},
		Data: map[string][]byte{
			prefix("TARGET_URL"): []byte(proxyConfig.TargetURL),
			"CREDENTIALS_TYPE":   []byte(credentialsType),
			"CONFIGURATION":      destinationConfig,
		},
	}

	_, err = sc.secretsClient.Create(secret)
	if err != nil {
		return fmt.Errorf("failed to create secret: %s", err.Error())
	}

	return nil
}

func getCredentialsType(proxyConfig proxyconfig.ProxyDestinationConfig) string {
	var credentialsType string
	switch proxyConfig.Configuration.Credentials.(type) {
	case proxyconfig.BasicAuthConfig:
		credentialsType = "BasicAuth"
	case proxyconfig.OauthConfig:
		credentialsType = "OAuth"
	case proxyconfig.CertificateConfig:
		credentialsType = "Certificate"
	default:
		credentialsType = "NoAuth"
	}

	return credentialsType
}

func newPrefixFunc(prefix string) func(string) string {
	return func(key string) string {
		return fmt.Sprintf("%s_%s", prefix, key)
	}
}
