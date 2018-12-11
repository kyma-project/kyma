package k8sconsts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameResolver(t *testing.T) {
	testCases := []struct {
		remotEnv     string
		id           string
		resourceName string
		metadataUrl  string
		host         string
	}{
		{
			remotEnv:     "short_application",
			id:           "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName: "re-short_application-c687e68a-9038-4f38-845b-9c61592e59e6",
			metadataUrl:  "http://re-short_application-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:         "re-short_application-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
		{
			remotEnv:     "max_application_aaaaaaaaa",
			id:           "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName: "re-max_application_aaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6",
			metadataUrl:  "http://re-max_application_aaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:         "re-max_application_aaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
		{
			remotEnv:     "toolong_application_aaaaaxxxx",
			id:           "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName: "re-toolong_application_aaa-c687e68a-9038-4f38-845b-9c61592e59e6",
			metadataUrl:  "http://re-toolong_application_aaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:         "re-toolong_application_aaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
	}

	t.Run("should get resource name with truncated application name if needed", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver("namespace")

			result := resolver.GetResourceName(testCase.remotEnv, testCase.id)

			assert.Equal(t, testCase.resourceName, result)
		}
	})

	t.Run("should get gateway url with truncated application name if needed", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver("namespace")

			result := resolver.GetGatewayUrl(testCase.remotEnv, testCase.id)

			assert.Equal(t, testCase.metadataUrl, result)
		}
	})

	t.Run("should extract service ID from gateway host", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver("namespace")

			result := resolver.ExtractServiceId(testCase.remotEnv, testCase.host)

			assert.Equal(t, testCase.id, result)
		}
	})
}
