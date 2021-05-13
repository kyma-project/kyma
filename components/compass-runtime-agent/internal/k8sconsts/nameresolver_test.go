package k8sconsts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameResolver(t *testing.T) {
	testCases := []struct {
		application                 string
		id                          string
		credentialsSecretName       string
		requestParametersSecretName string
	}{
		{
			application:                 "short_app",
			id:                          "c687e68a-9038-4f38-845b-9c61592e59e6",
			credentialsSecretName:       "short_app-credentials-c687e68a-9038-4f38-845b-9c61592e59e6",
			requestParametersSecretName: "short_app-params-c687e68a-9038-4f38-845b-9c61592e59e6",
		},
		{
			application:                 "app_12345678901",
			id:                          "c687e68a-9038-4f38-845b-9c61592e59e6",
			credentialsSecretName:       "app_12345678901-credential-c687e68a-9038-4f38-845b-9c61592e59e6",
			requestParametersSecretName: "app_12345678901-params-c687e68a-9038-4f38-845b-9c61592e59e6",
		},
		{
			application:                 "app_1234567890123456",
			id:                          "c687e68a-9038-4f38-845b-9c61592e59e6",
			credentialsSecretName:       "app_1234567890123456-crede-c687e68a-9038-4f38-845b-9c61592e59e6",
			requestParametersSecretName: "app_1234567890123456-param-c687e68a-9038-4f38-845b-9c61592e59e6",
		},
	}

	t.Run("should get credentials secret secret name with truncated application name if needed", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver()

			result := resolver.GetCredentialsSecretName(testCase.application, testCase.id)

			assert.Equal(t, testCase.credentialsSecretName, result)
		}
	})

	t.Run("should get request parameters secret name with truncated application name if needed", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver()

			result := resolver.GetRequestParametersSecretName(testCase.application, testCase.id)

			assert.Equal(t, testCase.requestParametersSecretName, result)
		}
	})
}
