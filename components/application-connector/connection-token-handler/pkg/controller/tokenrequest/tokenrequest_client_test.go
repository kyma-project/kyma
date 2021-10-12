package tokenrequest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getServerMock(t *testing.T, appName, tenant, group, token string) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/applications/tokens" {
				checkHeaders(t, appName, r, tenant, group)

				w.Header().Add("Content-Type", "application/json")
				w.Write([]byte(fmt.Sprintf(
					`{"token": "%s", "url": "http://url.with.token?token=%s"}`, token, token)))
			}
		}),
	)
}

func checkHeaders(t *testing.T, appName string, r *http.Request, tenant string, group string) {
	assert.Equal(t, appName, r.Header.Get(applicationHeader))
	if tenant != emptyTenant && group != emptyGroup {
		assert.Equal(t, tenant, r.Header.Get(tenantHeader))
		assert.Equal(t, group, r.Header.Get(groupHeader))
	}
}

func TestTokenRequestClient_FetchToken(t *testing.T) {
	appName := "some-app"
	token := "some-long-token-value"

	t.Run("should return TokenDto with valid token", func(t *testing.T) {
		//given
		srvMock := getServerMock(t, appName, emptyTenant, emptyGroup, token)
		defer srvMock.Close()
		//when
		svcClient := NewConnectorServiceClient(srvMock.URL)
		tokenDto, err := svcClient.FetchToken(appName, emptyTenant, emptyGroup)
		//then
		assert.NoError(t, err)
		assert.NotNil(t, tokenDto)
		assert.Equal(t, token, tokenDto.Token)
	})

	t.Run("should return TokenDto with valid token when tenant and group provided", func(t *testing.T) {
		//given
		tenant := "some-tenant"
		group := "some-group"
		srvMock := getServerMock(t, appName, tenant, group, token)
		defer srvMock.Close()
		//when
		svcClient := NewConnectorServiceClient(srvMock.URL)
		tokenDto, err := svcClient.FetchToken(appName, tenant, group)
		//then
		assert.NoError(t, err)
		assert.NotNil(t, tokenDto)
		assert.Equal(t, token, tokenDto.Token)
	})

	t.Run("should return error when calling invalid URL", func(t *testing.T) {
		//given
		srvMock := getServerMock(t, appName, emptyTenant, emptyGroup, token)
		defer srvMock.Close()
		//when
		svcClient := NewConnectorServiceClient(srvMock.URL + "/some-text")
		_, err := svcClient.FetchToken(appName, emptyTenant, emptyGroup)
		//then
		assert.Error(t, err)
	})
}
