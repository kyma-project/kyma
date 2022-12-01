package main

import (
	"context"
	"net/http"

	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/sirupsen/logrus"
)

func main() {
	named, err := reference.WithName("scratch")
	if err != nil {
		panic(err)
	}

	tr := &RegistryTransport{}
	repo, err := client.NewRepository(reference.TrimNamed(named), "http://localhost:5000/", tr)
	if err != nil {
		panic(err)
	}
	tags := repo.Tags(context.Background())
	tag, err := tags.Get(context.Background(), "1")
	// 1stTag:= tags.Get(context.Background(), "1")

	manifests, err := repo.Manifests(context.Background())
	manifests.Delete(context.Background(), tag.Digest)
	// m, err := manifests.Get(context.Background(), "", distribution.WithTag("1"), client.ReturnContentDigest(&tag.Digest))
	logrus.Infof("----------------------------------------------------------------- tags: %#v, err: %v", tag, err)
}

type RegistryTransport struct {
	http.RoundTripper
}

func (rt *RegistryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	return http.DefaultTransport.RoundTrip(req)
}
