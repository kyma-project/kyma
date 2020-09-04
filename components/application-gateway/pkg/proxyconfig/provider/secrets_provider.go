package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/proxyconfig"

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
// Expected secret format:
//	{API_NAME}_TARGET_URL
//  CONFIGURATION - JSON containing credentials and additional parameters
//	CREDENTIALS_TYPE - Credentials type - BasicAuth, OAuth, Certificate (not supported in Director) or NoAuth
func (r *repository) GetDestinationConfig(secretName string, apiName string) (proxyconfig.ProxyDestinationConfig, apperrors.AppError) {
	secret, err := r.client.Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return proxyconfig.ProxyDestinationConfig{}, apperrors.NotFound("secret '%s' not found", secretName)
		}
		return proxyconfig.ProxyDestinationConfig{}, apperrors.Internal("failed to get '%s' secret, %s", secretName, err.Error())
	}

	if secret.Data == nil {
		return proxyconfig.ProxyDestinationConfig{}, apperrors.WrongInput("provided secret is empty")
	}

	apiNameKey := strings.ToUpper(apiName)

	prefixKey := newPrefixFunc(apiNameKey)

	targetURL, found := secret.Data[prefixKey("TARGET_URL")]
	if !found {
		return proxyconfig.ProxyDestinationConfig{}, apperrors.NotFound("target URL for API %s not found in the %s secret", apiName, secretName)
	}

	secretType := parseType(getStringVal(secret.Data, "CREDENTIALS_TYPE"))
	if secretType == proxyconfig.Undefined {
		r.log.Warnf("WARNING: secret %s type is undefined, it will not use any auth", secretName)
	}

	configuration := proxyconfig.Configuration{
		Credentials: r.credentialsType(secretType),
	}

	rawConfiguration, found := secret.Data["CONFIGURATION"]
	if !found {
		r.log.Warnf("WARNING: secret %s does not contain additional configuration", secretName)
	} else {
		err = json.Unmarshal(rawConfiguration, &configuration)
		if err != nil {
			return proxyconfig.ProxyDestinationConfig{}, apperrors.Internal("failed to unmarshal target config from %s secret: %s", secretName, err.Error())
		}
	}

	return proxyconfig.ProxyDestinationConfig{
		TargetURL:     string(targetURL),
		Configuration: configuration,
	}, nil
}

func newPrefixFunc(prefix string) func(string) string {
	return func(key string) string {
		return fmt.Sprintf("%s_%s", prefix, key)
	}
}

func (r *repository) credentialsType(secretType proxyconfig.AuthType) proxyconfig.Credentials {
	switch secretType {
	case proxyconfig.Basic:
		return &proxyconfig.BasicAuthConfig{}
	case proxyconfig.Oauth:
		return &proxyconfig.OauthConfig{}
	case proxyconfig.Certificate:
		return &proxyconfig.CertificateConfig{}
	default:
		return &proxyconfig.NoAuthConfig{}
	}
}

func parseType(stringType string) proxyconfig.AuthType {
	secretType := proxyconfig.AuthType(strings.ToLower(stringType))

	switch secretType {
	case proxyconfig.NoAuth, proxyconfig.Basic, proxyconfig.Oauth, proxyconfig.Certificate:
		return secretType
	default:
		return proxyconfig.Undefined
	}
}

func getStringVal(data map[string][]byte, key string) string {
	val, found := data[key]
	if !found {
		return ""
	}

	return string(val)
}
