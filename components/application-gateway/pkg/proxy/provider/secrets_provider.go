package provider

import (
	"encoding/json"
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

// GetDestinationConfig fetches destination config from the secret of specified name
func (r *repository) GetDestinationConfig(name string) (proxy.ProxyDestinationConfig, apperrors.AppError) {
	secret, err := r.client.Get(name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return proxy.ProxyDestinationConfig{}, apperrors.NotFound("secret '%s' not found", name)
		}
		return proxy.ProxyDestinationConfig{}, apperrors.Internal("failed to get '%s' secret, %s", name, err.Error())
	}

	if secret.Data == nil {
		return proxy.ProxyDestinationConfig{}, apperrors.WrongInput("provided secret is empty")
	}

	// TODO: adjust keys to contract decided with Broker
	secretType := parseType(getStringVal(secret.Data, "type"))
	if secretType == proxy.Undefined {
		r.log.Warnf("WARNING: secret %s type is undefined, it will not use any auth", name)
	}

	proxyConfig := proxy.ProxyDestinationConfig{
		Destination: proxy.Destination{},
		Credentials: r.credentialsType(secretType),
	}

	err = json.Unmarshal(secret.Data["configuration"], &proxyConfig)
	if err != nil {
		return proxy.ProxyDestinationConfig{}, apperrors.Internal("failed to unmarshal target config from %s secret: %s", name, err.Error())
	}

	return proxyConfig, nil
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
