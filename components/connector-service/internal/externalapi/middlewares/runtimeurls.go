package middlewares

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"net/http"
)

const (
	//Header names are not fully consistent with documentation, due to net/http library. It changes all header names to start with uppercase, with following lowercase letters
	BaseEventsHostHeader   = "Eventshost"
	BaseMetadataHostHeader = "Metadatahost"
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
			MetadataHost: cc.defaultMetadataHost,
			EventsHost:   cc.defaultEventsHost,
		}

		metadataHosts, found := r.Header[BaseMetadataHostHeader]
		if found {
			runtimeURLs.MetadataHost = metadataHosts[0]
		}

		eventsHost, found := r.Header[BaseEventsHostHeader]
		if found {
			runtimeURLs.EventsHost = eventsHost[0]
		}

		reqWithCtx := r.WithContext(runtimeURLs.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
