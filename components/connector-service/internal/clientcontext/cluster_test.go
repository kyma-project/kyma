package clientcontext

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterContext_IsEmpty(t *testing.T) {

	testCases := []struct {
		tenant    string
		group     string
		runtimeID string
		result    bool
	}{
		{tenant: tenant, group: group, runtimeID: runtimeID, result: false},
		{tenant: tenant, group: "", runtimeID: runtimeID, result: true},
		{tenant: "", group: group, runtimeID: runtimeID, result: true},
		{tenant: "", group: "", runtimeID: runtimeID, result: true},
		{tenant: tenant, group: group, runtimeID: "", result: true},
		{tenant: tenant, group: "", runtimeID: "", result: true},
		{tenant: "", group: group, runtimeID: "", result: true},
		{tenant: "", group: "", runtimeID: "", result: true},
	}

	t.Run("should check if empty", func(t *testing.T) {
		for _, test := range testCases {
			cc := ClientContext{
				ClusterContext: ClusterContext{
					Tenant: test.tenant,
					Group:  test.group,
				},
				ID: test.runtimeID,
			}

			assert.Equal(t, test.result, cc.IsEmpty())
		}
	})
}

func TestClusterContext_ExtendContext(t *testing.T) {

	t.Run("should extend context with cluster context", func(t *testing.T) {
		// given
		clusterContext := ClientContext{
			ClusterContext: ClusterContext{
				Group:  "group",
				Tenant: "tenant"},
			ID: "runtimeID",
		}

		// when
		extended := clusterContext.ExtendContext(context.Background())

		// then
		assert.Equal(t, clusterContext, extended.Value(ClientContextKey))
	})
}
