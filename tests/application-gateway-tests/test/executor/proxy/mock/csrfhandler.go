package mock

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	firstToken      = "firstToken"
	secondToken     = "secondToken"
	HeaderCSRFToken = "X-csrf-token"
)

type csrfHandler struct {
	logger       *log.Entry
	isFirstToken bool
}

func NewCsrfHandler() *csrfHandler {
	return &csrfHandler{
		logger:       log.WithField("Handler", "CSRF"),
		isFirstToken: true,
	}
}

func (ch *csrfHandler) CsrfToken(w http.ResponseWriter, r *http.Request) {
	ch.logger.Infof("Handling CSRF request")

	var token string
	if ch.isFirstToken {
		ch.logger.Infof("Providing first token: %s", firstToken)
		token = firstToken
	} else {
		ch.logger.Infof("Providing second token: %s", secondToken)
		token = secondToken
	}

	w.Header().Set(HeaderCSRFToken, token)
	http.SetCookie(w, nil)

	successResponse(w)
}

func (ch *csrfHandler) Target(w http.ResponseWriter, r *http.Request) {
	var expectedToken string
	if ch.isFirstToken {
		expectedToken = firstToken
	} else {
		expectedToken = secondToken
	}

	ch.logger.Infof("Handling request. Expected: header: %s, with value: %s", HeaderCSRFToken, expectedToken)

	token := r.Header.Get(HeaderCSRFToken)
	if token != expectedToken {
		ch.logger.Errorf("Invalid CSRF token: %s", token)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ch.isFirstToken = false
	successResponse(w)
}
