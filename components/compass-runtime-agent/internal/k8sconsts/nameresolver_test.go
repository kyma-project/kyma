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
			application:                 "short_application",
			id:                          "c687e68a-9038-4f38-845b-9c61592e59e6",
			credentialsSecretName:       "short_application-c687e68a-9038-4f38-845b-9c61592e59e6",
			requestParametersSecretName: "params-short_application-c687e68a-9038-4f38-845b-9c61592e59e6",
		},
		{
			application:                 "max_application_aaaaaaaaaa",
			id:                          "c687e68a-9038-4f38-845b-9c61592e59e6",
			credentialsSecretName:       "max_application_aaaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6",
			requestParametersSecretName: "params-max_application_aaa-c687e68a-9038-4f38-845b-9c61592e59e6",
		},
		{
			application:                 "toolong_application_aaaaaxxxx",
			id:                          "c687e68a-9038-4f38-845b-9c61592e59e6",
			credentialsSecretName:       "toolong_application_aaaaax-c687e68a-9038-4f38-845b-9c61592e59e6",
			requestParametersSecretName: "params-toolong_application-c687e68a-9038-4f38-845b-9c61592e59e6",
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
