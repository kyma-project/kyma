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
