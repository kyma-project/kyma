package mock

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	firstToken      = "firstToken"
	secondToken     = "secondToken"
	HeaderCSRFToken = "X-csrf-token"
	cookieName      = "cookieToken"
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
	http.SetCookie(w, &http.Cookie{
		Name:  cookieName,
		Value: token,
	})

	successResponse(w)
}

func (ch *csrfHandler) Target(w http.ResponseWriter, r *http.Request) {
	var expectedToken string
	if ch.isFirstToken {
		expectedToken = firstToken
	} else {
		expectedToken = secondToken
	}

	ch.logger.Infof("Handling request. Expected: header: %s, with value: %s, cookie: %s, with value: %s", HeaderCSRFToken, expectedToken, cookieName, expectedToken)

	headerToken := r.Header.Get(HeaderCSRFToken)
	if headerToken != expectedToken {
		ch.logger.Errorf("Invalid header: %s with CSRF token value: %s", HeaderCSRFToken, headerToken)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	cookieToken, err := r.Cookie(cookieName)
	if err != nil {
		ch.logger.Errorf("No cookie: %s", cookieName)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if cookieToken.Value != expectedToken {
		ch.logger.Errorf("Invalid cookie: %s with CSRF token value: %s", cookieName, cookieToken.Value)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ch.isFirstToken = false
	successResponse(w)
}
