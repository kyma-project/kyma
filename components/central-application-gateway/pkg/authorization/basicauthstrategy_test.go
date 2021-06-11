package authorization

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAuthStrategy(t *testing.T) {

	t.Run("should add Authorization header", func(t *testing.T) {
		// given
		basicAuthStrategy := newBasicAuthStrategy("username", "password")

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = basicAuthStrategy.AddAuthorization(request, nil)

		// then
		require.NoError(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Equal(t, "Basic dXNlcm5hbWU6cGFzc3dvcmQ=", authHeader)
	})
}
