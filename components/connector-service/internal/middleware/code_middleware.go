package middleware

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"github.com/kyma-project/kyma/components/connector-service/internal/middleware/metrics"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{w, 0}
}

func (lrw *responseWriterWrapper) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

type codeMiddleware struct {
	metricsService metrics.Collector
}

func NewCodeMiddleware(metricsService metrics.Collector) (*codeMiddleware) {
	return &codeMiddleware{metricsService: metricsService}
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
			dm.metricsService.AddObservation(template, r.Method, float64(writerWrapper.statusCode))
		}
	})
}
