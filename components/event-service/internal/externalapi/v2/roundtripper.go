package v2

import "net/http"

//import "net/http"

type Roundtripper struct {
	before func(r *http.Request)
}

func (r Roundtripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if r.before != nil {
		r.before(request)
	}
	return http.DefaultTransport.RoundTrip(request)
}
