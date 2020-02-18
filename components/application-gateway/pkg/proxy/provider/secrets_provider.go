package provider

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/proxy"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type repository struct {
	client v12.SecretInterface
	log    *logrus.Entry
}

// NewSecretsProxyTargetConfigProvider creates a new secrets repository
func NewSecretsProxyTargetConfigProvider(client v12.SecretInterface) *repository {
	return &repository{
		client: client,
		log:    logrus.WithField("Component", "SecretsClient"),
	}
}

// TODO: adjust keys to contract decided with Broker
// Expected secret format:
//	{API_NAME}_GATEWAY_URL - Gateway URL with proper path
//	{API_NAME}_CREDENTIALS_TYPE - Credentials type - BasicAuth, OAuth, Certificate (not supported in Director) or NoAuth
//	{API_NAME}_DESTINATION

// GetDestinationConfig fetches destination config from the secret of specified name
func (r *repository) GetDestinationConfig(secretName string, apiName string) (proxy.ProxyDestinationConfig, apperrors.AppError) {
	secret, err := r.client.Get(secretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return proxy.ProxyDestinationConfig{}, apperrors.NotFound("secret '%s' not found", secretName)
		}
		return proxy.ProxyDestinationConfig{}, apperrors.Internal("failed to get '%s' secret, %s", secretName, err.Error())
	}

	if secret.Data == nil {
		return proxy.ProxyDestinationConfig{}, apperrors.WrongInput("provided secret is empty")
	}

	apiNameKey := strings.ToUpper(apiName)

	prefixKey := newPrefixFunc(apiNameKey)

	secretType := parseType(getStringVal(secret.Data, prefixKey("CREDENTIALS_TYPE")))
	if secretType == proxy.Undefined {
		r.log.Warnf("WARNING: secret %s type is undefined, it will not use any auth", secretName)
	}

	proxyConfig := proxy.ProxyDestinationConfig{
		Destination: proxy.Destination{},
		Credentials: r.credentialsType(secretType),
	}

	destinationData, found := secret.Data[prefixKey("DESTINATION")]
	if !found {
		return proxy.ProxyDestinationConfig{}, apperrors.NotFound("API %s destination configuration not found in %s secret", apiName, secretName)
	}

	err = json.Unmarshal(destinationData, &proxyConfig)
	if err != nil {
		return proxy.ProxyDestinationConfig{}, apperrors.Internal("failed to unmarshal target config from %s secret: %s", secretName, err.Error())
	}

	return proxyConfig, nil
}

func newPrefixFunc(prefix string) func(string) string {
	return func(key string) string {
		return fmt.Sprintf("%s_%s", prefix, key)
	}
}

func (r *repository) credentialsType(secretType proxy.AuthType) proxy.Credentials {
	switch secretType {
	case proxy.Basic:
		return &proxy.BasicAuthConfig{}
	case proxy.Oauth:
		return &proxy.OauthConfig{}
	case proxy.Certificate:
		return &proxy.CertificateConfig{}
	default:
		return &proxy.NoAuthConfig{}
	}
}

func parseType(stringType string) proxy.AuthType {
	secretType := proxy.AuthType(strings.ToLower(stringType))

	switch secretType {
	case proxy.NoAuth, proxy.Basic, proxy.Oauth, proxy.Certificate:
		return secretType
	default:
		return proxy.Undefined
	}
}

func getStringVal(data map[string][]byte, key string) string {
	val, found := data[key]
	if !found {
		return ""
	}

	return string(val)
}
