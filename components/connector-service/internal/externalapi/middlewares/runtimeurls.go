package middlewares

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
)

type runtimeURLsMiddleware struct {
	gatewayHost                 string
	lookupEnabled               clientcontext.LookupEnabledType
	lookupConfigPath            string
	applicationContextExtractor clientcontext.ApplicationContextExtractor
	lookupService               LookupService
}

func NewRuntimeURLsMiddleware(gatewayHost, lookupConfigPath string, lookupEnabled clientcontext.LookupEnabledType, extractor clientcontext.ApplicationContextExtractor, lookupService LookupService) *runtimeURLsMiddleware {
	return &runtimeURLsMiddleware{
		gatewayHost:                 gatewayHost,
		lookupEnabled:               lookupEnabled,
		lookupConfigPath:            lookupConfigPath,
		applicationContextExtractor: extractor,
		lookupService:               lookupService,
	}
}

func (cc *runtimeURLsMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runtimeURLs := &clientcontext.APIHosts{
			MetadataHost: cc.gatewayHost,
			EventsHost:   cc.gatewayHost,
		}

		if cc.lookupEnabled {
			appCtx, appError := cc.applicationContextExtractor(r.Context())
			if appError != nil {
				httphelpers.RespondWithErrorAndLog(w, apperrors.Internal("Could not read Application Context. %s", appError))
				return
			}
			fetchedGatewayHost, appErr := cc.lookupService.Fetch(appCtx, cc.lookupConfigPath)

			if appErr != nil {
				httphelpers.RespondWithErrorAndLog(w, apperrors.Internal("Could not fetch gateway URL. %s", appErr))
			}
			runtimeURLs.EventsHost = fetchedGatewayHost
			runtimeURLs.MetadataHost = fetchedGatewayHost
		}

		reqWithCtx := r.WithContext(runtimeURLs.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
