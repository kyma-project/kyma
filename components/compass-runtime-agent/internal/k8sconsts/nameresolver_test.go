package k8sconsts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameResolver(t *testing.T) {
	testCases := []struct {
		application             string
		id                      string
		resourceName            string
		metadataUrl             string
		host                    string
		secretName              string
		requestParamsSecretName string
	}{
		{
			application:             "short_application",
			id:                      "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName:            "short_application-c687e68a-9038-4f38-845b-9c61592e59e6",
			metadataUrl:             "http://short_application-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:                    "short_application-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			secretName:              "short_application-c687e68a-9038-4f38-845b-9c61592e59e6",
			requestParamsSecretName: "params-short_application-c687e68a-9038-4f38-845b-9c61592e59e6",
		},
		{
			application:             "max_application_aaaaaaaaaa",
			id:                      "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName:            "max_application_aaaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6",
			metadataUrl:             "http://max_application_aaaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:                    "max_application_aaaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			secretName:              "max_application_aaaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6",
			requestParamsSecretName: "params-max_application_aaaaaaaaaa-c687e68a-9038-4f38-845b-9c615",
		},
		{
			application:             "toolong_application_aaaaaxxxx",
			id:                      "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName:            "toolong_application_aaaaax-c687e68a-9038-4f38-845b-9c61592e59e6",
			metadataUrl:             "http://toolong_application_aaaaax-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:                    "toolong_application_aaaaax-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			secretName:              "toolong_application_aaaaax-c687e68a-9038-4f38-845b-9c61592e59e6",
			requestParamsSecretName: "params-toolong_application_aaaaax-c687e68a-9038-4f38-845b-9c615",
		},
	}

	t.Run("should get resource name with truncated application name if needed", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver("namespace")

			result := resolver.GetResourceName(testCase.application, testCase.id)

			assert.Equal(t, testCase.resourceName, result)
		}
	})

	t.Run("should get gateway url with truncated application name if needed", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver("namespace")

			result := resolver.GetGatewayUrl(testCase.application, testCase.id)

			assert.Equal(t, testCase.metadataUrl, result)
		}
	})

	t.Run("should extract service ID from gateway host", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver("namespace")

			result := resolver.ExtractServiceId(testCase.application, testCase.host)

			assert.Equal(t, testCase.id, result)
		}
	})

	t.Run("should get credentials secret name", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver("namespace")

			result := resolver.GetCredentialsSecretName(testCase.application, testCase.id)

			assert.Equal(t, testCase.secretName, result)
		}
	})

	t.Run("should get request parameters secret name", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver("namespace")

			result := resolver.GetRequestParamsSecretName(testCase.application, testCase.id)

			assert.Equal(t, testCase.requestParamsSecretName, result)
		}
	})
}
