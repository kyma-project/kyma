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
	logger            *log.Entry
	provideFirstToken bool
	expectFirstToken  bool
}

func NewCsrfHandler() *csrfHandler {
	return &csrfHandler{
		logger:            log.WithField("Handler", "CSRF"),
		provideFirstToken: true,
		expectFirstToken:  true,
	}
}

func (ch *csrfHandler) CsrfToken(w http.ResponseWriter, r *http.Request) {
	ch.logger.Infof("Handling CSRF request")

	var token string
	if ch.provideFirstToken {
		ch.logger.Infof("Providing correct token: %s", firstToken)
		token = firstToken
	} else {
		ch.logger.Infof("Providing incorrect token: %s", secondToken)
		token = secondToken
	}

	w.Header().Set(HeaderCSRFToken, token)
	http.SetCookie(w, nil)

	ch.provideFirstToken = false
	successResponse(w)
}

func (ch *csrfHandler) Target(w http.ResponseWriter, r *http.Request) {
	var expectedToken string
	if ch.expectFirstToken {
		expectedToken = firstToken
		ch.provideFirstToken = true
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

	ch.expectFirstToken = false
	successResponse(w)
}
