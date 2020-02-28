package provider

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/proxyconfig"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"

	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const (
	secretName = "secret"

	oauthJSONData           = `,"credentials":{"clientId":"abcd-efgh", "clientSecret":"hgfe-dcba", "tokenUrl":"https://token/token"}`
	basicAuthJSONData       = `,"credentials":{"username":"user","password":"pwd"}`
	certificateAuthJSONData = `,"credentials":{"privateKey":"cHJpdktleQ==","certificate":"Y2VydA=="}`

	secretDataFormat = `{"requestParameters":{"headers":{"header1":["h1_value"]},"queryParameters":{"query1":["q1_value"]}},"csrfConfig":{"tokenUrl":"https://csrf/token"}%s}`
)

func Test_GetDestination(t *testing.T) {

	for _, testCase := range []struct {
		description string
		secretData  map[string][]byte
		assertFunc  func(t *testing.T, config proxyconfig.ProxyDestinationConfig)
	}{
		{
			description: "oauth credentials",
			secretData: map[string][]byte{
				"MY_API_TARGET_URL": []byte("https://my-application.com"),
				"CREDENTIALS_TYPE":  []byte("Oauth"),
				"CONFIGURATION":     oauthConfigJson(),
			},
			assertFunc: func(t *testing.T, config proxyconfig.ProxyDestinationConfig) {
				credentials, ok := config.Configuration.Credentials.(*proxyconfig.OauthConfig)
				require.True(t, ok)
				assert.Equal(t, "abcd-efgh", credentials.ClientId)
				assert.Equal(t, "hgfe-dcba", credentials.ClientSecret)
				assert.Equal(t, "https://token/token", credentials.TokenURL)
			},
		},
		{
			description: "basic auth credentials",
			secretData: map[string][]byte{
				"MY_API_TARGET_URL": []byte("https://my-application.com"),
				"CREDENTIALS_TYPE":  []byte("BasicAuth"),
				"CONFIGURATION":     basicAuthConfigJson(),
			},
			assertFunc: func(t *testing.T, config proxyconfig.ProxyDestinationConfig) {
				credentials, ok := config.Configuration.Credentials.(*proxyconfig.BasicAuthConfig)
				require.True(t, ok)
				assert.Equal(t, "user", credentials.Username)
				assert.Equal(t, "pwd", credentials.Password)
			},
		},
		{
			description: "certificate credentials",
			secretData: map[string][]byte{
				"MY_API_TARGET_URL": []byte("https://my-application.com"),
				"CREDENTIALS_TYPE":  []byte("Certificate"),
				"CONFIGURATION":     certificateConfigJson(),
			},
			assertFunc: func(t *testing.T, config proxyconfig.ProxyDestinationConfig) {
				credentials, ok := config.Configuration.Credentials.(*proxyconfig.CertificateConfig)
				require.True(t, ok)
				assert.Equal(t, "privKey", string(credentials.PrivateKey))
				assert.Equal(t, "cert", string(credentials.Certificate))
			},
		},
		{
			description: "no auth",
			secretData: map[string][]byte{
				"MY_API_TARGET_URL": []byte("https://my-application.com"),
				"CREDENTIALS_TYPE":  []byte("noauth"),
				"CONFIGURATION":     []byte(fmt.Sprintf(secretDataFormat, "")),
			},
			assertFunc: func(t *testing.T, config proxyconfig.ProxyDestinationConfig) {
				_, ok := config.Configuration.Credentials.(*proxyconfig.NoAuthConfig)
				require.True(t, ok)
			},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			k8sClient := fake.NewSimpleClientset(
				&v1.Secret{
					ObjectMeta: v12.ObjectMeta{Name: secretName},
					Data:       testCase.secretData,
				})

			repository := NewSecretsProxyTargetConfigProvider(k8sClient.CoreV1().Secrets(""))

			// when
			proxyTargetConfig, err := repository.GetDestinationConfig(secretName, "my_api")
			require.NoError(t, err)

			// then
			assert.Equal(t, "https://my-application.com", proxyTargetConfig.TargetURL)
			assert.Equal(t, &map[string][]string{"header1": {"h1_value"}}, proxyTargetConfig.Configuration.RequestParameters.Headers)
			assert.Equal(t, &map[string][]string{"query1": {"q1_value"}}, proxyTargetConfig.Configuration.RequestParameters.QueryParameters)
			testCase.assertFunc(t, proxyTargetConfig)
		})
	}

	for _, testCase := range []struct {
		description string
		secretData  map[string][]byte
		errorCode   int
	}{
		{
			description: "should return error if failed to unmarshal configuration",
			secretData: map[string][]byte{
				"MY_API_TARGET_URL": []byte("https://my-application.com"),
				"CONFIGURATION":     []byte("not json"),
			},
			errorCode: apperrors.CodeInternal,
		},
		{
			description: "should return not found error if Target URL for API not found in the secret",
			secretData:  map[string][]byte{},
			errorCode:   apperrors.CodeNotFound,
		},
		{
			description: "should return error if secret data is nil",
			secretData:  nil,
			errorCode:   apperrors.CodeWrongInput,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			k8sClient := fake.NewSimpleClientset(
				&v1.Secret{
					ObjectMeta: v12.ObjectMeta{Name: secretName},
					Data:       testCase.secretData,
				})

			repository := NewSecretsProxyTargetConfigProvider(k8sClient.CoreV1().Secrets(""))

			// when
			_, err := repository.GetDestinationConfig(secretName, "my_api")

			// then
			require.Error(t, err)
			assert.Equal(t, testCase.errorCode, err.Code())
		})
	}

	t.Run("should return not found error if secret not found", func(t *testing.T) {
		// given
		k8sClient := fake.NewSimpleClientset()

		repository := NewSecretsProxyTargetConfigProvider(k8sClient.CoreV1().Secrets(""))

		// when
		_, err := repository.GetDestinationConfig(secretName, "my_api")

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})
}

func oauthConfigJson() []byte {
	return []byte(fmt.Sprintf(secretDataFormat, oauthJSONData))
}

func basicAuthConfigJson() []byte {
	return []byte(fmt.Sprintf(secretDataFormat, basicAuthJSONData))
}

func certificateConfigJson() []byte {
	return []byte(fmt.Sprintf(secretDataFormat, certificateAuthJSONData))
}
