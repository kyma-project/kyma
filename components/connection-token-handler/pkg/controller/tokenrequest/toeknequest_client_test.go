package tokenrequest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getServerMock(reName, token string) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/remoteenvironments/"+reName+"/tokens" {
				w.Header().Add("Content-Type", "application/json")
				w.Write([]byte(fmt.Sprintf(
					`{"token": "%s", "url": "http://url.with.token?token=%s"}`, token, token)))
			}
		}),
	)
}
func TestTokenRequestClient_FetchToken(t *testing.T) {
	reName := "some-re"
	token := "some-long-token-value"

	t.Run("should return TokenDto with valid token", func(t *testing.T) {
		srvMock := getServerMock(reName, token)
		defer srvMock.Close()

		svcClient := NewConnectorServiceClient(srvMock.URL)
		tokenDto, err := svcClient.FetchToken(reName)

		assert.NoError(t, err)
		assert.NotNil(t, tokenDto)
		assert.Equal(t, token, tokenDto.Token)
	})

	t.Run("should return error when calling invalid URL", func(t *testing.T) {
		srvMock := getServerMock(reName, token)
		defer srvMock.Close()

		svcClient := NewConnectorServiceClient(srvMock.URL + "/some-text")
		_, err := svcClient.FetchToken(reName)

		assert.Error(t, err)
	})
}
