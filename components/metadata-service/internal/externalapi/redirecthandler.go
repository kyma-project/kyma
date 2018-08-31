package externalapi

import (
	"net/http"
	"github.com/gorilla/mux"
	"strings"
	"fmt"
)

const pathVarFormat = "{%s}"

// NewRedirectHandler creates handler for redirection that replaces path variables
func NewRedirectHandler(destination string, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		destination = replaceVarsInPath(destination, vars)

		http.Redirect(w, r, destination, status)
	})
}

func replaceVarsInPath(path string, vars map[string]string) string {
	for key, val := range vars {
		path = strings.Replace(path, fmt.Sprintf(pathVarFormat, key), val, -1)
	}

	return path
}