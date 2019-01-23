package httpcontext

import (
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
