package middleware

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/monitoring/collector"
	"github.com/sirupsen/logrus"
)

type CodeMiddleware struct {
	metricsCollector collector.Collector
}

func NewCodeMiddleware(metricsCollector collector.Collector) *CodeMiddleware {
	return &CodeMiddleware{metricsCollector: metricsCollector}
}

func (dm *CodeMiddleware) Handle(next http.Handler) http.Handler {
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
			statusLabel := strconv.FormatInt(int64(writerWrapper.statusCode), 10)
			dm.metricsCollector.AddObservation(1, template, statusLabel, r.Method)
		}
	})
}
