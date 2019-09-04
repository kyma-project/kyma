package mock

import (
	"net/http"

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
	err := bah.checkBasicAuth(r)
	if err != nil {
		bah.logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (bah *basicAuthHandler) checkBasicAuth(r *http.Request) error {
	vars := mux.Vars(r)
	expectedUserName := vars["username"]
	expectedPassword := vars["password"]

	bah.logger.Infof("Handling BasicAuth request. Expected: username: %s, password: %s", expectedUserName, expectedPassword)

	if expectedPassword == "" || expectedUserName == "" {
		return errors.New("Expected credentials not provided")
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		return errors.Errorf("Basic credentials not provided. Authorization header: %s", r.Header.Get(AuthorizationHeader))
	}

	if username != expectedUserName || password != expectedPassword {
		return errors.New("Invalid credentials provided")
	}

	return nil
}
