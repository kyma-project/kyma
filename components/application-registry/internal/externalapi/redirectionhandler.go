package externalapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const pathVarFormat = "{%s}"

// NewRedirectionHandler creates handler for redirection that replaces path variables
func NewRedirectionHandler(destination string, status int) RedirectionHandler {
	return &redirectionHandler{
		destination: destination,
		status:      status,
	}
}

type redirectionHandler struct {
	destination string
	status      int
}

func (rh *redirectionHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := replaceVarsInPath(rh.destination, vars)

	http.Redirect(w, r, path, rh.status)
}

func replaceVarsInPath(path string, vars map[string]string) string {
	for key, val := range vars {
		path = strings.Replace(path, fmt.Sprintf(pathVarFormat, key), val, -1)
	}

	return path
}
