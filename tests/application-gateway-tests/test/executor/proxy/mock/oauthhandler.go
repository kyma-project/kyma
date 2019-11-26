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

func (oh *oauthHandler) OAuthTargetHandler(w http.ResponseWriter, r *http.Request) {
	err := oh.checkOauth(r)
	if err != nil {
		oh.logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
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

func (oh *oauthHandler) checkBasic(r *http.Request, expectedClientId, expectedClientSecret string) error {
	if expectedClientId == "" || expectedClientSecret == "" {
		return fmt.Errorf("Expected credentials not provided. ClientID, ClientSecret or both not provided")
	}

	headerAuthorization := r.Header.Get(AuthorizationHeader)

	encodedCredentials := strings.TrimPrefix(headerAuthorization, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return fmt.Errorf("Failed to decode credentials, %v", err)
	}

	credentials := strings.Split(string(decoded), ":")
	if len(credentials) < 2 {
		return fmt.Errorf("Decoded credentials are incomplete")
	}

	clientId := credentials[0]
	clientSecret := credentials[1]

	if clientId != expectedClientId || clientSecret != expectedClientSecret {
		return fmt.Errorf("Invalid credentials provided clientID: %s, clientSecret: %s", clientId, clientSecret)
	}

	return nil
}

func (oh *oauthHandler) checkQueryParams(r *http.Request, expectedParam, expectedParamValue string) error {
	paramValue := r.URL.Query().Get(expectedParam)
	if expectedParamValue != paramValue {
		return fmt.Errorf("Invalid query parameter value provided")
	}

	return nil
}

func (oh *oauthHandler) checkHeaders(r *http.Request, expectedHeader, expectedHeaderValue string) error {
	headerValue := r.Header.Get(expectedHeader)
	if expectedHeaderValue != headerValue {
		return errors.New("Invalid header value provided")
	}

	return nil
}

func (oh *oauthHandler) OAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	oh.logger.Info("Handling Oauth token request")

	vars := mux.Vars(r)
	expectedClientId := vars["clientid"]
	expectedClientSecret := vars["clientsecret"]

	oh.logger.Infof("Handling OAuth secured spec request. Expected: clientID: %s, clientSecret: %s", expectedClientId, expectedClientSecret)

	err := oh.checkBasic(r, expectedClientId, expectedClientSecret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		oh.logger.Error(err.Error())
		return
	}

	err = respondWithOk(w)
	if err != nil {
		oh.logger.Error(err.Error())
		return
	}
}

func (oh *oauthHandler) OAuthTokenQueryParamsHandler(w http.ResponseWriter, r *http.Request) {
	oh.logger.Info("Handling Oauth token request")

	vars := mux.Vars(r)
	expectedClientId := vars["clientid"]
	expectedClientSecret := vars["clientsecret"]
	expectedParam := vars["param"]
	expectedParamValue := vars["value"]

	oh.logger.Infof("Handling OAuth secured spec request. Expected: clientID: %s, clientSecret: %s, param: %s, with value: %s", expectedClientId, expectedClientSecret, expectedParam, expectedParamValue)

	err := oh.checkQueryParams(r, expectedParam, expectedParamValue)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		oh.logger.Error(err.Error())
		return
	}

	err = oh.checkBasic(r, expectedClientId, expectedClientSecret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		oh.logger.Error(err.Error())
		return
	}

	err = respondWithOk(w)
	if err != nil {
		oh.logger.Error(err.Error())
		return
	}
}

func (oh *oauthHandler) OAuthTokenHeadersHandler(w http.ResponseWriter, r *http.Request) {
	oh.logger.Info("Handling Oauth token request")

	vars := mux.Vars(r)
	expectedClientId := vars["clientid"]
	expectedClientSecret := vars["clientsecret"]
	expectedHeader := vars["header"]
	expectedHeaderValue := vars["value"]

	oh.logger.Infof("Handling OAuth secured spec request. Expected: clientID: %s, clientSecret: %s, header: %s, with value: %s", expectedClientId, expectedClientSecret, expectedHeader, expectedHeaderValue)

	err := oh.checkHeaders(r, expectedHeader, expectedHeaderValue)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		oh.logger.Error(err.Error())
		return
	}

	err = oh.checkBasic(r, expectedClientId, expectedClientSecret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		oh.logger.Error(err.Error())
		return
	}

	err = respondWithOk(w)
	if err != nil {
		oh.logger.Error(err.Error())
		return
	}
}

func respondWithOk(w http.ResponseWriter) error {
	oauthResponseOk := oauthResponse{
		AccessToken: bearerToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "",
	}
	return respondWithBody(w, http.StatusOK, oauthResponseOk)
}

func respondWithBody(w http.ResponseWriter, statusCode int, responseBody interface{}) error {
	var b bytes.Buffer

	err := json.NewEncoder(&b).Encode(responseBody)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to marshall body, %s", err))
	}

	respond(w, statusCode)
	_, err = w.Write(b.Bytes())
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to write bytes, %v", err))
	}
	return nil
}

func respond(w http.ResponseWriter, statusCode int) {
	w.Header().Set(headerContentType, contentTypeApplicationJson)
	w.WriteHeader(statusCode)
}
