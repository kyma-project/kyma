package registry

import "net/http"

type RegistryTransport struct {
	http.RoundTripper
}

func (rt *RegistryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	return http.DefaultTransport.RoundTrip(req)
}
