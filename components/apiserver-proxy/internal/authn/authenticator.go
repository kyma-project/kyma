package authn

import (
	"github.com/golang/glog"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"net/http"
)

type ProxyAuthenticator struct {
	authenticators []authenticator.Request
}

//AuthenticateRequest iterates over all registered authenticator and tries to authenticate given request. If all of them fail
func (p *ProxyAuthenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {

	for i, v := range p.authenticators {
		r, ok, err := v.AuthenticateRequest(req)
		if err != nil {
			glog.Warningf("Unable to authenticate the request due to an error: %v", err)
			if hasNext(i, p.authenticators) {
				continue
			}
			return r, ok, err
		}
	}
}

func NewProxyAuthenticator(list []authenticator.Request) ProxyAuthenticator {
	return ProxyAuthenticator{authenticators: list}
}
func (p *ProxyAuthenticator) Add(requestAuthenticator authenticator.Request) {
	p.authenticators = append(p.authenticators, requestAuthenticator)
}

func hasNext(index int, arr []authenticator.Request) bool {
	return index != (len(arr) - 1)
}
