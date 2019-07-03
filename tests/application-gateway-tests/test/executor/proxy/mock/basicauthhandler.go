package mock

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	AuthorizationHeader = "Authorization"
)

type basicAuthHandler struct {
	logger *log.Entry
}

func NewBasicAuthHandler() *basicAuthHandler {
	return &basicAuthHandler{
		logger: log.WithField("Handler", "BasicAuth"),
	}
}

func (bah *basicAuthHandler) BasicAuth(w http.ResponseWriter, r *http.Request) {
	statusCode, e := bah.checkBasicAuth(r)
	if e != nil {
		bah.logger.Error(e.Error())
		w.WriteHeader(statusCode)
		return
	}
	w.WriteHeader(statusCode)
}

func (bah *basicAuthHandler) BasicAuthSpec(w http.ResponseWriter, r *http.Request) {
	statusCode, e := bah.checkBasicAuth(r)
	if e != nil {
		bah.logger.Error(e.Error())
		w.WriteHeader(statusCode)
		return
	}
	w.WriteHeader(statusCode)
	http.ServeFile(w, r, "spec.json")
}

func (bah *basicAuthHandler) checkBasicAuth(r *http.Request) (statusCode int, err error) {
	vars := mux.Vars(r)
	expectedUserName := vars["username"]
	expectedPassword := vars["password"]

	bah.logger.Infof("Handling BasicAuth request. Expected: username: %s, password: %s", expectedUserName, expectedPassword)

	if expectedPassword == "" || expectedUserName == "" {
		return http.StatusBadRequest, errors.New("Expected credentials not provided")
	}

	authorizationHeader := r.Header.Get(AuthorizationHeader)

	encodedCredentials := strings.TrimPrefix(authorizationHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return http.StatusBadRequest, errors.New("Failed to decode credentials")
	}

	credentials := strings.Split(string(decoded), ":")
	userName := credentials[0]
	password := credentials[1]

	if userName != expectedUserName || password != expectedPassword {
		return http.StatusBadRequest, errors.New("Invalid credentials provided")
	}
	return http.StatusOK, nil
}
