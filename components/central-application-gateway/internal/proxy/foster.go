package proxy

import (
	"net/url"
	"strings"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
)

func ProxyRedirectRewriter(gw *url.URL, target *url.URL) func(*url.URL) {

	return nil
}

func ExtractGatewatURL(u *url.URL) (*url.URL, apperrors.AppError) {
	trimmed := strings.TrimPrefix(u.Path, "/")
	split := strings.SplitN(trimmed, "/", 3)

	if len(split) < 2 {
		return nil, apperrors.WrongInput("GW: path must contain Application and Service name: %v => %v", u.Path, split)
	}

	new := *u
	new.Path = "/" + strings.Join(split[:2], "/")

	return &new, nil
}
