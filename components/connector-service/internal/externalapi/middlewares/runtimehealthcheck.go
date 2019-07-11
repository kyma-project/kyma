package middlewares

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/externalapi/middlewares/runtimeregistry"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"net/http"
)

type runtimeHealthCheckMiddleware struct {
	connectorClientExtractor clientcontext.ConnectorClientExtractor
	runtimeRegistryService   runtimeregistry.RuntimeRegistryService
}

func NewRuntimeHealthCheckMiddleware(connectorClientExtractor clientcontext.ConnectorClientExtractor, runtimeRegistryService runtimeregistry.RuntimeRegistryService) *runtimeHealthCheckMiddleware {
	return &runtimeHealthCheckMiddleware{
		connectorClientExtractor: connectorClientExtractor,
		runtimeRegistryService:   runtimeRegistryService,
	}
}

func (cc *runtimeHealthCheckMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextService, err := cc.connectorClientExtractor(r.Context())
		if err != nil {
			httphelpers.RespondWithErrorAndLog(w, apperrors.BadRequest("Invalid certificate: %s", err.Error()))
			return
		}

		runtimeID := contextService.GetSubject().CommonName

		if runtimeID == clientcontext.RuntimeDefaultCommonName {
			handler.ServeHTTP(w, r)
			return
		}

		state := runtimeregistry.RuntimeState{Identifier: runtimeID, State: runtimeregistry.EstablishedState}

		e := cc.runtimeRegistryService.ReportState(state)

		if e != nil {
			httphelpers.RespondWithErrorAndLog(w, apperrors.Internal("Unable to report runtime state: %s", e.Error()))
			return
		}

		handler.ServeHTTP(w, r)
	})
}
