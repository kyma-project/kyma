package kube_config

import (
	"net/http"
	"regexp"
)

const MimeTypeYaml = "application/x-yaml"

type Endpoints struct {
	kubeConfig *KubeConfig
}

func NewEndpoints(c *KubeConfig) *Endpoints {
	return &Endpoints{c}
}

var bearerPattern = regexp.MustCompile("[Bb]earer ")

func (e *Endpoints) GetKubeConfig(w http.ResponseWriter, req *http.Request) {

	w.Header().Add("Content-Type", MimeTypeYaml)

	authorization := req.Header.Get("Authorization")
	if authorization == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing Authorization header with Bearer token."))
		return
	}

	if !bearerPattern.MatchString(authorization) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid value of the Authorization header. Bearer token is required."))
		return
	}

	token := bearerPattern.ReplaceAllString(authorization, "")

	e.kubeConfig.Generate(w, token)
}
