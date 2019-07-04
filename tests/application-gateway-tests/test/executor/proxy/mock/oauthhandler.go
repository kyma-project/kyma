package mock

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	bearerToken                = "1/mZ1edKKACtPAb7zGlwSzvs72PvhAbGmB8K1ZrGxpcNM"
	headerContentType          = "Content-Type"
	contentTypeApplicationJson = "application/json;charset=UTF-8"
)

type oauthHandler struct {
	logger *log.Entry
}

func NewOauthHandler() *basicAuthHandler {
	return &basicAuthHandler{
		logger: log.WithField("Handler", "OAuth"),
	}
}

type oauthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func (oh *oauthHandler) OAuthTokenHandler(logger *log.Entry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("Handling Oauth request")

		oauthRes := oauthResponse{
			AccessToken: bearerToken,
			TokenType:   "Bearer",
			ExpiresIn:   3600,
			Scope:       "",
		}

		respondWithBody(w, 200, oauthRes)
	})
}

func respondWithBody(w http.ResponseWriter, statusCode int, responseBody interface{}) error {
	var b bytes.Buffer

	err := json.NewEncoder(&b).Encode(responseBody)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to marshall body, %s", err))
	}

	respond(w, statusCode)
	w.Write(b.Bytes())
	return nil
}

func respond(w http.ResponseWriter, statusCode int) {
	w.Header().Set(headerContentType, contentTypeApplicationJson)
	w.WriteHeader(statusCode)
}
