package mock

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

const (
	bearerToken                = "1/mZ1edKKACtPAb7zGlwSzvs72PvhAbGmB8K1ZrGxpcNM"
	headerContentType          = "Content-Type"
	contentTypeApplicationJson = "application/json;charset=UTF-8"
)

type oauthHandler struct {
	logger *log.Entry
}

func NewOauthHandler() *oauthHandler {
	return &oauthHandler{
		logger: log.WithField("Handler", "OAuth"),
	}
}

type oauthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func (oh *oauthHandler) OAuthSpecHandler(w http.ResponseWriter, r *http.Request) {
	err := oh.checkOauth(r)
	if err != nil {
		oh.logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, "spec.json")
}

func (oh *oauthHandler) checkOauth(r *http.Request) error {
	headerAuthorization := r.Header.Get(AuthorizationHeader)
	oAuthToken := strings.TrimPrefix(headerAuthorization, "Bearer ")

	if oAuthToken != bearerToken {
		return errors.New("Invalid token provided")
	}
	return nil
}

func (oh *oauthHandler) OAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	oh.logger.Info("Handling Oauth token request")

	vars := mux.Vars(r)
	expectedClientId := vars["clientid"]
	expectedClientSecret := vars["clientsecret"]

	oh.logger.Infof("Handling OAuth secured spec request. Expected: clientID: %s, clientSecret: %s", expectedClientId, expectedClientSecret)

	if expectedClientId == "" || expectedClientSecret == "" {
		w.WriteHeader(http.StatusBadRequest)
		oh.logger.Error("Expected credentials not provided. ClientID, ClientSecret or both not provided")
		return
	}

	headerAuthorization := r.Header.Get(AuthorizationHeader)

	encodedCredentials := strings.TrimPrefix(headerAuthorization, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		oh.logger.Errorf("Failed to decode credentials, %s", err.Error())
		return
	}

	credentials := strings.Split(string(decoded), ":")
	if len(credentials) < 2 {
		w.WriteHeader(http.StatusBadRequest)
		oh.logger.Error("Decoded credentials are incomplete")
		return
	}

	clientId := credentials[0]
	clientSecret := credentials[1]

	if clientId != expectedClientId || clientSecret != expectedClientSecret {
		w.WriteHeader(http.StatusBadRequest)
		oh.logger.Errorf("Invalid credentials provided clientID: %s, clientSecret: %s", clientId, clientSecret)
		return
	}

	oauthRes := oauthResponse{
		AccessToken: bearerToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "",
	}

	err = respondWithBody(w, http.StatusOK, oauthRes)
	if err != nil {
		oh.logger.Error(err.Error())
		return
	}
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
