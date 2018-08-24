package broker

import (
	"net/http"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sirupsen/logrus"
)

// OSBContextMiddleware implements Handler interface
type OSBContextMiddleware struct {
	brokerFlavor brokerFlavorProvider
	log          *logrus.Entry
}

//go:generate mockery -name=brokerFlavorProvider -output=automock -outpkg=automock -case=underscore
type brokerFlavorProvider interface {
	IsClusterScoped() bool
	GetNsFromBrokerURL(url string) (string, error)
}

// NewOsbContextMiddleware created OsbContext middleware
func NewOsbContextMiddleware(brokerFlavorProvider brokerFlavorProvider, log *logrus.Entry) *OSBContextMiddleware {
	return &OSBContextMiddleware{
		brokerFlavor: brokerFlavorProvider,
		log:          log.WithField("service", "OSBContextMiddleware"),
	}
}

// ServeHTTP adds content of Open Service Broker Api headers to the requests
func (m *OSBContextMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	brokerNamespace := ""
	if m.brokerFlavor.IsClusterScoped() == false {
		var err error
		brokerNamespace, err = m.brokerFlavor.GetNsFromBrokerURL(r.Host)
		if err != nil {
			errMsg := "misconfiguration, broker is running as a namespace-scoped, but cannot extract namespace from request host"
			m.log.Error(errMsg, err)
			writeErrorResponse(rw, http.StatusInternalServerError, errMsg, "")
			return
		}
	}

	osbCtx := osbContext{
		APIVersion:          r.Header.Get(osb.APIVersionHeader),
		OriginatingIdentity: r.Header.Get(osb.OriginatingIdentityHeader),
		BrokerNamespace:     brokerNamespace,
		ClusterScopedBroker: m.brokerFlavor.IsClusterScoped(),
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
