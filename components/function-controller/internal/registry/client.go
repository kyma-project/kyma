package registry

import (
	"github.com/heroku/docker-registry-client/registry"
)

func NewRegistryClient() (*registry.Registry, error) {
	url := "http://localhost:5000/"
	username := "" // anonymous
	password := "" // anonymous
	return registry.New(url, username, password)
}
