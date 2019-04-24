package clientcontext

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClusterContextEnabledStrategy_ReadClusterContextFromRequest(t *testing.T) {

	t.Run("should read cluster context from request", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "/v1/applications/tokens", nil)
		require.NoError(t, err)

		req.Header.Set(TenantHeader, tenant)
		req.Header.Set(GroupHeader, group)

		strategy := NewClusterContextStrategy(true)

		// when
		clusterCtx := strategy.ReadClusterContextFromRequest(req)

		// then
		assert.Equal(t, tenant, clusterCtx.Tenant)
		assert.Equal(t, group, clusterCtx.Group)
	})

}

func TestClusterContextEnabledStrategy_IsValidContext(t *testing.T) {

	testCases := []struct {
		tenant string
		group  string
		valid  bool
	}{
		{tenant: tenant, group: group, valid: true},
		{tenant: tenant, group: "", valid: false},
		{tenant: "", group: group, valid: false},
		{tenant: "", group: "", valid: false},
	}

	t.Run("should validate context", func(t *testing.T) {
		strategy := NewClusterContextStrategy(true)

		for _, test := range testCases {
			valid := strategy.IsValidContext(ClusterContext{Tenant: test.tenant, Group: test.group})
			assert.Equal(t, test.valid, valid)
		}
	})
}

func TestClusterContextDisabledStrategy_IsValidContext(t *testing.T) {

	testCases := []struct {
		tenant string
		group  string
		valid  bool
	}{
		{tenant: tenant, group: group, valid: false},
		{tenant: tenant, group: "", valid: false},
		{tenant: "", group: group, valid: false},
		{tenant: "", group: "", valid: true},
	}

	t.Run("should validate context", func(t *testing.T) {
		strategy := NewClusterContextStrategy(false)

		for _, test := range testCases {
			valid := strategy.IsValidContext(ClusterContext{Tenant: test.tenant, Group: test.group})
			assert.Equal(t, test.valid, valid)
		}
	})
}

func TestClusterContextDisabledStrategy_ReadClusterContextFromRequest(t *testing.T) {

	t.Run("should return empty cluster context from request", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodPost, "/v1/applications/tokens", nil)
		require.NoError(t, err)

		req.Header.Set(TenantHeader, tenant)
		req.Header.Set(GroupHeader, group)

		strategy := NewClusterContextStrategy(false)

		// when
		clusterCtx := strategy.ReadClusterContextFromRequest(req)

		// then
		assert.Empty(t, clusterCtx)
	})

}
