package mock

import (
	"net/http"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

const (
	HeaderCSRFToken = "X-csrf-token"
	cookieName      = "CSRF-Token"
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

	vars := mux.Vars(r)
	tokenToReturn := vars["token"]

	ch.logger.Infof("Responding with token %s", tokenToReturn)

	w.Header().Set(HeaderCSRFToken, tokenToReturn)
	http.SetCookie(w, &http.Cookie{
		Name:  cookieName,
		Value: tokenToReturn,
	})

	successResponse(w)
}

func (ch *csrfHandler) Target(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	expectedToken := vars["expectedToken"]

	ch.logger.Infof("Handling CSRF target request. Expected: header: %s, with value: %s, cookie: %s, with value: %s", HeaderCSRFToken, expectedToken, cookieName, expectedToken)

	headerToken := r.Header.Get(HeaderCSRFToken)
	if headerToken != expectedToken {
		ch.logger.Errorf("Invalid %s header: Expected: %s, Actual: %s", HeaderCSRFToken, expectedToken, headerToken)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	cookieToken, err := r.Cookie(cookieName)
	if err != nil {
		ch.logger.Errorf("Cookie %s not provided. Expected value: %s", cookieName, expectedToken)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if cookieToken.Value != expectedToken {
		ch.logger.Errorf("Invalid %s cookie: Expected: %s, Actual: %s", cookieName, cookieToken.Value, expectedToken)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ch.isFirstToken = false
	successResponse(w)
}
