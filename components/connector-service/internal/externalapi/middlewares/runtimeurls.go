package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
)

const (
	BaseEventsHostHeader   = "EventsHost"
	BaseMetadataHostHeader = "MetadataHost"
)

type runtimeURLsMiddleware struct {
	defaultMetadataHost string
	defaultEventsHost   string
}

func NewRuntimeURLsMiddleware(defaultMetadataHost string, defaultEventsHost string) *runtimeURLsMiddleware {
	return &runtimeURLsMiddleware{
		defaultMetadataHost: defaultMetadataHost,
		defaultEventsHost:   defaultEventsHost,
	}
}

func (cc *runtimeURLsMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runtimeURLs := &clientcontext.APIHosts{
			MetadataHost: r.Header.Get(BaseMetadataHostHeader),
			EventsHost:   r.Header.Get(BaseEventsHostHeader),
		}

		if runtimeURLs.MetadataHost == "" {
			runtimeURLs.MetadataHost = cc.defaultMetadataHost
		}

		if runtimeURLs.EventsHost == "" {
			runtimeURLs.EventsHost = cc.defaultEventsHost
		}

		reqWithCtx := r.WithContext(runtimeURLs.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
