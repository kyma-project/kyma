package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
)

const (
	//Header names are not fully consistent with documentation, due to net/http library. It changes all header names to start with uppercase, with following lowercase letters
	BaseEventsHostHeader   = "Eventshost"
	BaseMetadataHostHeader = "Metadatahost"
)

type runtimeURLsMiddleware struct {
	gatewayHost string
	required    clientcontext.CtxRequiredType
}

func NewRuntimeURLsMiddleware(gatewayHost string, required clientcontext.CtxRequiredType) *runtimeURLsMiddleware {
	return &runtimeURLsMiddleware{
		gatewayHost: gatewayHost,
		required:    required,
	}
}

func (cc *runtimeURLsMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runtimeURLs := &clientcontext.APIHosts{
			MetadataHost: cc.gatewayHost,
			EventsHost:   cc.gatewayHost,
		}

		if metadataHosts, found := r.Header[BaseMetadataHostHeader]; found {
			runtimeURLs.MetadataHost = metadataHosts[0]
		} else if found == false && bool(cc.required) {
			httphelpers.RespondWithError(w, apperrors.BadRequest("Required headers not specified (%s).", BaseMetadataHostHeader))
			return
		}

		if eventsHosts, found := r.Header[BaseEventsHostHeader]; found {
			runtimeURLs.EventsHost = eventsHosts[0]
		} else if found == false && bool(cc.required) {
			httphelpers.RespondWithError(w, apperrors.BadRequest("Required headers not specified (%s).", BaseEventsHostHeader))
			return
		}

		reqWithCtx := r.WithContext(runtimeURLs.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
