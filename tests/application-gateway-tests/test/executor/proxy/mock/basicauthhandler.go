package mock

import (
	"encoding/base64"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
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
	vars := mux.Vars(r)
	expectedUserName := vars["username"]
	expectedPassword := vars["password"]

	bah.logger.Infof("Handling BasicAuth request. Expected: username: %s, password: %s", expectedUserName, expectedPassword)

	if expectedPassword == "" || expectedUserName == "" {
		bah.logger.Errorf("Expected credentials not provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authorizationHeader := r.Header.Get(AuthorizationHeader)

	encodedCredentials := strings.TrimPrefix(authorizationHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		bah.logger.Errorf("Failed to decode credentials")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	credentials := strings.Split(string(decoded), ":")
	userName := credentials[0]
	password := credentials[1]

	if userName != expectedUserName || password != expectedPassword {
		bah.logger.Errorf("Invalid credentials provided")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	successResponse(w)
}
