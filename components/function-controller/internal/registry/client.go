package registry

import (
	"github.com/heroku/docker-registry-client/registry"
)

func NewRegistryClient() (*registry.Registry, error) {
	url := "http://localhost:5001/"
	username := "" // anonymous
	password := "" // anonymous
	return registry.New(url, username, password)
}
