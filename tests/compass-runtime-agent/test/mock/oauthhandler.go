package mock

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

const (
	headerContentType          = "Content-Type"
	contentTypeApplicationJson = "application/json;charset=UTF-8"

	accessTokenFormat = "%s:%s"
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

func (oh *oauthHandler) OAuthHandler(w http.ResponseWriter, r *http.Request) {
	oh.logger.Info("Handling Oauth request")

	headerAuthorization := r.Header.Get(AuthorizationHeader)
	oAuthToken := strings.TrimPrefix(headerAuthorization, "Bearer ")

	vars := mux.Vars(r)
	expectedClientId := vars["clientid"]
	expectedClientSecret := vars["clientsecret"]

	oh.logger.Infof("Handling OAuth request. Expected: clientID: %s, clientSecret: %s", expectedClientId, expectedClientSecret)

	expectedToken := oh.tokenFromCredentials(expectedClientId, expectedClientSecret)

	if oAuthToken != expectedToken {
		w.WriteHeader(http.StatusBadRequest)
		oh.logger.Errorf("Invalid token provided. Expected: %s, Actual: %s", expectedToken, oAuthToken)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (oh *oauthHandler) OAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	oh.logger.Info("Handling Oauth token request")

	vars := mux.Vars(r)
	expectedClientId := vars["clientid"]
	expectedClientSecret := vars["clientsecret"]

	oh.logger.Infof("Handling OAuth token request. Expected: clientID: %s, clientSecret: %s", expectedClientId, expectedClientSecret)

	if expectedClientId == "" || expectedClientSecret == "" {
		w.WriteHeader(http.StatusBadRequest)
		oh.logger.Error("Expected credentials not provided. ClientID, ClientSecret or both not provided")
		return
	}

	clientId, clientSecret, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		oh.logger.Errorf("Basic credentials (clientId and clientSecret) not provided. Authorization header: %s", r.Header.Get(AuthorizationHeader))
		return
	}

	if clientId != expectedClientId || clientSecret != expectedClientSecret {
		w.WriteHeader(http.StatusBadRequest)
		oh.logger.Errorf("Invalid credentials provided clientID: %s, clientSecret: %s", clientId, clientSecret)
		return
	}

	token := oh.tokenFromCredentials(clientId, clientSecret)

	oauthRes := oauthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "",
	}

	err := respondWithBody(w, http.StatusOK, oauthRes)
	if err != nil {
		oh.logger.Error(err.Error())
		return
	}
}

func (oh *oauthHandler) tokenFromCredentials(clientId, clientSecret string) string {
	tokenString := fmt.Sprintf(accessTokenFormat, clientId, clientSecret)
	return base64.StdEncoding.EncodeToString([]byte(tokenString))
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
