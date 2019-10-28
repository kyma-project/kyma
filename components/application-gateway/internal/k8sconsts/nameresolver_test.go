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
		gatewayUrl   string
		host         string
	}{
		{
			application:  "short_application",
			id:           "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName: "short_application-c687e68a-9038-4f38-845b-9c61592e59e6",
			gatewayUrl:   "http://short_application-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:         "short_application-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
		{
			application:  "max_application_aaaaaaaaaa",
			id:           "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName: "max_application_aaaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6",
			gatewayUrl:   "http://max_application_aaaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:         "max_application_aaaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
		{
			application:  "toolong_application_aaaaaxxxx",
			id:           "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName: "toolong_application_aaaaax-c687e68a-9038-4f38-845b-9c61592e59e6",
			gatewayUrl:   "http://toolong_application_aaaaax-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:         "toolong_application_aaaaax-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
	}

	t.Run("should get resource name with truncated application name if needed", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver(testCase.application)

			result := resolver.GetResourceName(testCase.id)

			assert.Equal(t, testCase.resourceName, result)
		}
	})

	t.Run("should extract service ID from the access service host name", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver(testCase.application)

			result := resolver.ExtractServiceId(testCase.host)

			assert.Equal(t, testCase.id, result)
		}
	})
}
