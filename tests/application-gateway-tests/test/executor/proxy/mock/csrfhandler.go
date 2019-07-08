package mock

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	correctToken    = "correctToken"
	incorrectToken  = "incorrectToken"
	HeaderCSRFToken = "X-csrf-token"
)

type csrfHandler struct {
	logger                  *log.Entry
	willProvideCorrectToken bool
}

func NewCsrfHandler() *csrfHandler {
	return &csrfHandler{
		logger:                  log.WithField("Handler", "CSRF"),
		willProvideCorrectToken: false,
	}
}

func (ch *csrfHandler) CsrfToken(w http.ResponseWriter, r *http.Request) {
	ch.logger.Infof("Handling CSRF request")

	var token string
	if ch.willProvideCorrectToken {
		ch.logger.Infof("Providing correct token: %s", correctToken)
		token = correctToken
	} else {
		ch.logger.Infof("Providing incorrect token: %s", incorrectToken)
		token = incorrectToken
	}

	w.Header().Set(HeaderCSRFToken, token)
	http.SetCookie(w, nil)

	ch.willProvideCorrectToken = true
	successResponse(w)
}

func (ch *csrfHandler) Target(w http.ResponseWriter, r *http.Request) {
	ch.logger.Infof("Handling request. Expected: header: %s, with value: %s", HeaderCSRFToken, correctToken)

	token := r.Header.Get(HeaderCSRFToken)
	if token != correctToken {
		ch.logger.Errorf("Invalid CSRF token: %s", token)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	successResponse(w)
}
