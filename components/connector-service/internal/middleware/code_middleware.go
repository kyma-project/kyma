package middleware

import (
	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/middleware/metrics"
	"github.com/sirupsen/logrus"
	"net/http"
)

type codeMiddleware struct {
	metricsCollector metrics.Collector
}

func NewCodeMiddleware(metricsCollector metrics.Collector) *codeMiddleware {
	return &codeMiddleware{metricsCollector: metricsCollector}
}

func (dm *codeMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writerWrapper := newResponseWriterWrapper(w)

		next.ServeHTTP(writerWrapper, r)

		route := mux.CurrentRoute(r)
		if route == nil {
			logrus.Warnf("No route matched '%s' for request tracking", r.RequestURI)
			return
		}

		template, err := route.GetPathTemplate()
		if err != nil {
			logrus.Errorf("Failed to get path template: %s", err.Error())
		} else {
			dm.metricsCollector.AddObservation(template, r.Method, float64(writerWrapper.statusCode))
		}
	})
}
