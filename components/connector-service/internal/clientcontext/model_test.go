package clientcontext

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterContext_IsEmpty(t *testing.T) {

	testCases := []struct {
		tenant string
		group  string
		result bool
	}{
		{"tenant", "group", false},
		{"tenant", "", true},
		{"", "group", true},
		{"", "", true},
	}

	t.Run("should check if empty", func(t *testing.T) {
		for _, test := range testCases {
			cc := ClusterContext{
				Tenant: test.tenant,
				Group:  test.group,
			}

			assert.Equal(t, test.result, cc.IsEmpty())
		}
	})
}

func TestApplicationContext_IsEmpty(t *testing.T) {

	testCases := []struct {
		tenant      string
		group       string
		application string
		result      bool
	}{
		{"tenant", "group", "app", false},
		{"tenant", "group", "", true},
		{"tenant", "", "application", true},
		{"", "group", "application", true},
		{"", "", "", true},
	}

	t.Run("should check if empty", func(t *testing.T) {
		for _, test := range testCases {
			appCtx := ApplicationContext{
				Application: test.application,
				ClusterContext: ClusterContext{
					Tenant: test.tenant,
					Group:  test.group,
				},
			}

			assert.Equal(t, test.result, appCtx.IsEmpty())
		}
	})
}

func TestApplicationContext_FillPlaceholders(t *testing.T) {

	appCtx := ApplicationContext{
		Application: "application",
		ClusterContext: ClusterContext{
			Tenant: "tenant",
			Group: "group",
		},
	}

	t.Run("should fill placeholders with values", func(t *testing.T) {
		// given
		expectedInfoURL := "https://test.cluster.cx/v1/tenant/" + appCtx.ClusterContext.Tenant +
			"/group/" + appCtx.ClusterContext.Group +
			"/application/" + appCtx.Application +
			"/applications/management/info"

		baseInfoURL := "https://test.cluster.cx/v1/tenant/" + TenantPlaceholder +
			"/group/" + GroupPlaceholder +
			"/application/" + ApplicationPlaceholder +
			"/applications/management/info"

		// when
		filledInfoURL := appCtx.FillPlaceholders(baseInfoURL)

		// then
		assert.Equal(t, expectedInfoURL, filledInfoURL)
	})

	t.Run("should leave the format intact if there are no placeholders", func(t *testing.T) {
		// given
		expectedInfoURL := "https://test.cluster.cx/v1/applications/management/info"
		baseInfoURL := expectedInfoURL

		// when
		filledInfoURL := appCtx.FillPlaceholders(baseInfoURL)

		// then
		assert.Equal(t, expectedInfoURL, filledInfoURL)
	})
}

func TestClusterContext_FillPlaceholders(t *testing.T) {

	clsCtx := ClusterContext{
		Tenant: "tenant",
		Group: "group",
	}

	t.Run("should fill placeholders with values", func(t *testing.T) {
		// given
		expectedInfoURL := "https://test.cluster.cx/v1/tenant/" + clsCtx.Tenant +
			"/group/" + clsCtx.Group +
			"/applications/management/info"

		baseInfoURL := "https://test.cluster.cx/v1/tenant/" + TenantPlaceholder +
			"/group/" + GroupPlaceholder +
			"/applications/management/info"

		// when
		filledInfoURL := clsCtx.FillPlaceholders(baseInfoURL)

		// then
		assert.Equal(t, expectedInfoURL, filledInfoURL)
	})

	t.Run("should leave the format intact if there are no placeholders", func(t *testing.T) {
		// given
		expectedInfoURL := "https://test.cluster.cx/v1/applications/management/info"
		baseInfoURL := expectedInfoURL

		// when
		filledInfoURL := clsCtx.FillPlaceholders(baseInfoURL)

		// then
		assert.Equal(t, expectedInfoURL, filledInfoURL)
	})
}

func TestApplicationContext_ExtendContext(t *testing.T) {

	t.Run("should extend context with application context", func(t *testing.T) {
		// given
		appContext := ApplicationContext{
			Application: "app",
		}

		// when
		extended := appContext.ExtendContext(context.Background())

		// then
		assert.Equal(t, appContext, extended.Value(ApplicationContextKey))
	})
}

func TestClusterContext_ExtendContext(t *testing.T) {

	t.Run("should extend context with cluster context", func(t *testing.T) {
		// given
		clusterContext := ClusterContext{
			Group:  "group",
			Tenant: "tenant",
		}

		// when
		extended := clusterContext.ExtendContext(context.Background())

		// then
		assert.Equal(t, clusterContext, extended.Value(ClusterContextKey))
	})
}
