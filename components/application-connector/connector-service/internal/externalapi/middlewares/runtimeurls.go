package middlewares

import (
	"net/http"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/clientcontext"
)

type runtimeURLsMiddleware struct {
	gatewayBaseURL              string
	lookupEnabled               clientcontext.LookupEnabledType
	lookupConfigPath            string
	applicationContextExtractor clientcontext.ApplicationContextExtractor
	lookupService               LookupService
}

func NewRuntimeURLsMiddleware(gatewayBaseURL, lookupConfigPath string, lookupEnabled clientcontext.LookupEnabledType, extractor clientcontext.ApplicationContextExtractor, lookupService LookupService) *runtimeURLsMiddleware {
	return &runtimeURLsMiddleware{
		gatewayBaseURL:              gatewayBaseURL,
		lookupEnabled:               lookupEnabled,
		lookupConfigPath:            lookupConfigPath,
		applicationContextExtractor: extractor,
		lookupService:               lookupService,
	}
}

func (cc *runtimeURLsMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runtimeURLs := &clientcontext.ApiURLs{
			MetadataBaseURL: cc.gatewayBaseURL,
			EventsBaseURL:   cc.gatewayBaseURL,
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
			runtimeURLs.EventsBaseURL = fetchedGatewayHost
			runtimeURLs.MetadataBaseURL = fetchedGatewayHost
		}

		reqWithCtx := r.WithContext(runtimeURLs.ExtendContext(r.Context()))

		handler.ServeHTTP(w, reqWithCtx)
	})
}
