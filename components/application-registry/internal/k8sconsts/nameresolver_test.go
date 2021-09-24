package k8sconsts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	namespace         = "namespace"
	centralGatewayUrl = "http://central-application-gateway.kyma-system:8080"
)

func TestNameResolver(t *testing.T) {
	testCases := []struct {
		application        string
		id                 string
		serviceDisplayName string
		resourceName       string
		metadataUrl        string
		centralGatewayUrl  string
		host               string
	}{
		{
			application:        "short_application",
			id:                 "c687e68a-9038-4f38-845b-9c61592e59e6",
			serviceDisplayName: "some service",
			resourceName:       "short_application-c687e68a-9038-4f38-845b-9c61592e59e6",
			metadataUrl:        "http://short_application-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			centralGatewayUrl:  "http://central-application-gateway.kyma-system:8080/short_application/some-service",
			host:               "short_application-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
		{
			application:        "max_application_aaaaaaaaaa",
			id:                 "c687e68a-9038-4f38-845b-9c61592e59e6",
			serviceDisplayName: "some service",
			resourceName:       "max_application_aaaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6",
			metadataUrl:        "http://max_application_aaaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			centralGatewayUrl:  "http://central-application-gateway.kyma-system:8080/max_application_aaaaaaaaaa/some-service",
			host:               "max_application_aaaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
		{
			application:        "toolong_application_aaaaaxxxx",
			id:                 "c687e68a-9038-4f38-845b-9c61592e59e6",
			serviceDisplayName: "some service",
			resourceName:       "toolong_application_aaaaax-c687e68a-9038-4f38-845b-9c61592e59e6",
			metadataUrl:        "http://toolong_application_aaaaax-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			centralGatewayUrl:  "http://central-application-gateway.kyma-system:8080/toolong_application_aaaaaxxxx/some-service",
			host:               "toolong_application_aaaaax-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
	}

	t.Run("should get resource name with truncated application name if needed", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver(namespace, centralGatewayUrl)

			result := resolver.GetResourceName(testCase.application, testCase.id)

			assert.Equal(t, testCase.resourceName, result)
		}
	})

	t.Run("should get gateway url with truncated application name if needed", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver(namespace, centralGatewayUrl)

			result := resolver.GetGatewayUrl(testCase.application, testCase.id)

			assert.Equal(t, testCase.metadataUrl, result)
		}
	})

	t.Run("should get central gateway url", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver(namespace, centralGatewayUrl)

			result := resolver.GetCentralGatewayUrl(testCase.application, testCase.serviceDisplayName)

			assert.Equal(t, testCase.centralGatewayUrl, result)
		}
	})

	t.Run("should extract service ID from gateway host", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver(namespace, centralGatewayUrl)

			result := resolver.ExtractServiceId(testCase.application, testCase.host)

			assert.Equal(t, testCase.id, result)
		}
	})
}
