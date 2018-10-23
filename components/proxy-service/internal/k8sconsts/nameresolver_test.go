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
		gatewayUrl   string
		host         string
	}{
		{
			remotEnv:     "short_remoteenv",
			id:           "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName: "re-short_remoteenv-c687e68a-9038-4f38-845b-9c61592e59e6",
			gatewayUrl:   "http://re-short_remoteenv-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:         "re-short_remoteenv-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
		{
			remotEnv:     "max_remoteenv_aaaaaaaaa",
			id:           "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName: "re-max_remoteenv_aaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6",
			gatewayUrl:   "http://re-max_remoteenv_aaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:         "re-max_remoteenv_aaaaaaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
		{
			remotEnv:     "toolong_remoteenv_aaaaaxxxx",
			id:           "c687e68a-9038-4f38-845b-9c61592e59e6",
			resourceName: "re-toolong_remoteenv_aaaaa-c687e68a-9038-4f38-845b-9c61592e59e6",
			gatewayUrl:   "http://re-toolong_remoteenv_aaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
			host:         "re-toolong_remoteenv_aaaaa-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local",
		},
	}

	t.Run("should get resource name with truncated remote environment name if needed", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver(testCase.remotEnv, "namespace")

			result := resolver.GetResourceName(testCase.id)

			assert.Equal(t, testCase.resourceName, result)
		}
	})
	

	t.Run("should extract service ID from the access service host name", func(t *testing.T) {
		for _, testCase := range testCases {
			resolver := NewNameResolver(testCase.remotEnv, "namespace")

			result := resolver.ExtractServiceId(testCase.host)

			assert.Equal(t, testCase.id, result)
		}
	})
}
