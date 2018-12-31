package k8sconsts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameResolver(t *testing.T) {
	testCases := []struct {
		application  string
		id           string
		resourceName string
		metadataUrl  string
		host         string
	}{
		{
			application:  "short_application",
			id:           "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName: "app-short_application-c687e68a-9038-4f38-845b-9c61592e59e6",
			metadataUrl:  "http://app-short_application-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:         "app-short_application-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
		{
			application:  "max_application_aaaaaaaaa",
			id:           "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName: "app-max_application_aaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6",
			metadataUrl:  "http://app-max_application_aaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:         "app-max_application_aaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
		{
			application:  "toolong_application_aaaaaxxxx",
			id:           "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName: "app-toolong_application_aa-c687e68a-9038-4f38-845b-9c61592e59e6",
			metadataUrl:  "http://app-toolong_application_aa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:         "app-toolong_application_aa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
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
}
