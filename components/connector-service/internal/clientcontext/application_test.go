package clientcontext

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplicationContext_IsEmpty(t *testing.T) {

	testCases := []struct {
		tenant      string
		group       string
		application string
		result      bool
	}{
		{"tenant", "group", "app", false},
		{"tenant", "group", "", true},
		{"tenant", "", "application", false},
		{"", "group", "application", false},
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
			Group:  "group",
		},
	}

	t.Run("should fill placeholders with values", func(t *testing.T) {
		// given
		expectedInfoURL := "https://test.cluster.cx/tenant/" + appCtx.ClusterContext.Tenant +
			"/group/" + appCtx.ClusterContext.Group +
			"/application/" + appCtx.Application +
			"/v1/applications/management/info"

		baseInfoURL := "https://test.cluster.cx/tenant/" + TenantPlaceholder +
			"/group/" + GroupPlaceholder +
			"/application/" + ApplicationPlaceholder +
			"/v1/applications/management/info"

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
