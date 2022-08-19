package oauth

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOauthClient_GetAuthorizationToken(t *testing.T) {
	t.Run("Should return oauth token", func(t *testing.T) {
		//given
		credentials := credentials{
			clientID:       "12345",
			clientSecret:   "some dark and scary secret",
			tokensEndpoint: "http://hydra:4445",
		}

		token := Token{
			AccessToken: "12345",
			Expiration:  1234,
		}

		client := NewTestClient(func(req *http.Request) *http.Response {
			username, secret, ok := req.BasicAuth()

			if ok && username == credentials.clientID && secret == credentials.clientSecret {
				jsonToken, err := json.Marshal(&token)

				require.NoError(t, err)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader(jsonToken)),
				}
			}
			return &http.Response{
				StatusCode: http.StatusForbidden,
			}
		})

		oauthClient := NewOauthClient(client, credentials)

		//when
		responseToken, err := oauthClient.GetAuthorizationToken()
		require.NoError(t, err)
		token.Expiration += time.Now().Unix()

		//then
		assert.Equal(t, token.AccessToken, responseToken.AccessToken)
		assert.Equal(t, token.Expiration, responseToken.Expiration)
	})
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}
