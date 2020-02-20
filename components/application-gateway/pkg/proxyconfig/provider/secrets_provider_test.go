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

	requestParameters = `,"requestParameters":{"headers":{"header1":["h1_value"]},"queryParameters":{"query1":["q1_value"]}}`
	csrfConfig        = `,"csrfConfig":{"tokenUrl":"https://csrf/token"}`

	secretDataFormat = `{"destination":{"url":"https://my-application.com"%s%s}%s}`
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
				"MY_API_CREDENTIALS_TYPE": []byte("Oauth"),
				"MY_API_DESTINATION":      oauthConfigJson(requestParameters, csrfConfig),
			},
			assertFunc: func(t *testing.T, config proxyconfig.ProxyDestinationConfig) {
				credentials, ok := config.Credentials.(*proxyconfig.OauthConfig)
				require.True(t, ok)
				assert.Equal(t, "abcd-efgh", credentials.ClientId)
				assert.Equal(t, "hgfe-dcba", credentials.ClientSecret)
				assert.Equal(t, "https://token/token", credentials.TokenURL)
			},
		},
		{
			description: "basic auth credentials",
			secretData: map[string][]byte{
				"MY_API_CREDENTIALS_TYPE": []byte("BasicAuth"),
				"MY_API_DESTINATION":      basicAuthConfigJson(requestParameters, csrfConfig),
			},
			assertFunc: func(t *testing.T, config proxyconfig.ProxyDestinationConfig) {
				credentials, ok := config.Credentials.(*proxyconfig.BasicAuthConfig)
				require.True(t, ok)
				assert.Equal(t, "user", credentials.Username)
				assert.Equal(t, "pwd", credentials.Password)
			},
		},
		{
			description: "certificate credentials",
			secretData: map[string][]byte{
				"MY_API_CREDENTIALS_TYPE": []byte("Certificate"),
				"MY_API_DESTINATION":      certificateConfigJson(requestParameters, csrfConfig),
			},
			assertFunc: func(t *testing.T, config proxyconfig.ProxyDestinationConfig) {
				credentials, ok := config.Credentials.(*proxyconfig.CertificateConfig)
				require.True(t, ok)
				assert.Equal(t, "privKey", string(credentials.PrivateKey))
				assert.Equal(t, "cert", string(credentials.Certificate))
			},
		},
		{
			description: "no auth",
			secretData: map[string][]byte{
				"MY_API_CREDENTIALS_TYPE": []byte("noauth"),
				"MY_API_DESTINATION":      []byte(fmt.Sprintf(secretDataFormat, requestParameters, csrfConfig, "")),
			},
			assertFunc: func(t *testing.T, config proxyconfig.ProxyDestinationConfig) {
				_, ok := config.Credentials.(*proxyconfig.NoAuthConfig)
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
			assert.Equal(t, "https://my-application.com", proxyTargetConfig.Destination.URL)
			assert.Equal(t, &map[string][]string{"header1": {"h1_value"}}, proxyTargetConfig.Destination.RequestParameters.Headers)
			assert.Equal(t, &map[string][]string{"query1": {"q1_value"}}, proxyTargetConfig.Destination.RequestParameters.QueryParameters)
			testCase.assertFunc(t, proxyTargetConfig)
		})
	}

	for _, testCase := range []struct {
		description string
		secretData  map[string][]byte
		errorCode   int
	}{
		{
			description: "should return error if failed to unmarshal destination",
			secretData:  map[string][]byte{"MY_API_DESTINATION": []byte("not json")},
			errorCode:   apperrors.CodeInternal,
		},
		{
			description: "should return not found error if destination for API not found in the secret",
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

func oauthConfigJson(requestParams, csrf string) []byte {
	return []byte(fmt.Sprintf(secretDataFormat, requestParams, csrf, oauthJSONData))
}

func basicAuthConfigJson(requestParams, csrf string) []byte {
	return []byte(fmt.Sprintf(secretDataFormat, requestParams, csrf, basicAuthJSONData))
}

func certificateConfigJson(requestParams, csrf string) []byte {
	return []byte(fmt.Sprintf(secretDataFormat, requestParams, csrf, certificateAuthJSONData))
}
