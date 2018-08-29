package broker

import (
	"net/http"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sirupsen/logrus"
)

// OSBContextMiddleware implements Handler interface
type OSBContextMiddleware struct {
	brokerMode brokerModeService
	log        logrus.FieldLogger
}

//go:generate mockery -name=brokerModeService -output=automock -outpkg=automock -case=underscore
type brokerModeService interface {
	IsClusterScoped() bool
	GetNsFromBrokerURL(url string) (string, error)
}

// NewOsbContextMiddleware created OsbContext middleware
func NewOsbContextMiddleware(brokerModeService brokerModeService, log logrus.FieldLogger) *OSBContextMiddleware {
	return &OSBContextMiddleware{
		brokerMode: brokerModeService,
		log:        log.WithField("service", "OSBContextMiddleware"),
	}
}

// ServeHTTP adds content of Open Service BrokerService Api headers to the requests
func (m *OSBContextMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	brokerNamespace := ""
	if m.brokerMode.IsClusterScoped() == false {
		var err error
		brokerNamespace, err = m.brokerMode.GetNsFromBrokerURL(r.Host)
		if err != nil {
			errMsg := "misconfiguration, broker is running as a namespace-scoped, but cannot extract namespace from request"
			m.log.Error(errMsg, err)
			writeErrorResponse(rw, http.StatusInternalServerError, "OSBContextMiddlewareError", errMsg)
			return
		}
	}

	osbCtx := osbContext{
		APIVersion:          r.Header.Get(osb.APIVersionHeader),
		OriginatingIdentity: r.Header.Get(osb.OriginatingIdentityHeader),
		BrokerNamespace:     brokerNamespace,
		ClusterScopedBroker: m.brokerMode.IsClusterScoped(),
	}

	r = r.WithContext(contextWithOSB(r.Context(), osbCtx))

	next(rw, r)
}

// RequireAsyncMiddleware asserts if request allows for asynchronous response
type RequireAsyncMiddleware struct{}

// ServeHTTP handling asynchronous HTTP requests in Open Service BrokerService Api
func (RequireAsyncMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.URL.Query().Get("accepts_incomplete") != "true" {
		// message and desc as defined in https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#response-2
		writeErrorResponse(rw, http.StatusUnprocessableEntity, "AsyncRequired", "This service plan requires client support for asynchronous service operations.")
		return
	}

	next(rw, r)
}
