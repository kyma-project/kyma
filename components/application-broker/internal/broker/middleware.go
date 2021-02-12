package broker

import (
	"net/http"

	"github.com/gorilla/mux"
	osb "github.com/kubernetes-sigs/go-open-service-broker-client/v2"
)

// OSBContextMiddleware implements Handler interface, creates an osbContext instance from HTTP data and stores in the request context.
type OSBContextMiddleware struct {
}

// ServeHTTP adds content of Open Service Broker Api headers to the requests
func (m *OSBContextMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	brokerNamespace := mux.Vars(r)["namespace"]

	osbCtx := osbContext{
		APIVersion:          r.Header.Get(osb.APIVersionHeader),
		OriginatingIdentity: r.Header.Get(osb.OriginatingIdentityHeader),
		BrokerNamespace:     brokerNamespace,
	}

	if err := osbCtx.validateAPIVersion(); err != nil {
		writeErrorResponse(rw, http.StatusPreconditionFailed, err.Error(), "Requests requires the 'X-Broker-API-Version' header specified")
		return
	}
	if err := osbCtx.validateOriginatingIdentity(); err != nil {
		writeErrorResponse(rw, http.StatusPreconditionFailed, err.Error(), "Requests requires the 'X-Broker-API-Originating-Identity' header specified")
		return
	}

	r = r.WithContext(contextWithOSB(r.Context(), osbCtx))

	next(rw, r)
}

// RequireAsyncMiddleware asserts if request allows for asynchronous response
type RequireAsyncMiddleware struct{}

// ServeHTTP handling asynchronous HTTP requests in Open Service Broker Api
func (RequireAsyncMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.URL.Query().Get("accepts_incomplete") != "true" {
		// message and desc as defined in https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#response-2
		writeErrorResponse(rw, http.StatusUnprocessableEntity, "AsyncRequired", "This service plan requires client support for asynchronous service operations.")
		return
	}

	next(rw, r)
}
