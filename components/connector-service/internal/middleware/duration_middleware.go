package middleware

import (
	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/middleware/metrics"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type durationMiddleware struct {
	metricsCollector metrics.Collector
}

func NewDurationMiddleware(metricsCollector metrics.Collector) *durationMiddleware {
	return &durationMiddleware{metricsCollector: metricsCollector}
}

func (dm *durationMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()

		next.ServeHTTP(w, r)

		route := mux.CurrentRoute(r)
		if route == nil {
			logrus.Warnf("No route matched '%s' for request tracking", r.RequestURI)
			return
		}

		template, err := route.GetPathTemplate()
		if err != nil {
			logrus.Errorf("Failed to get path template: %s", err.Error())
		} else {
			dm.metricsCollector.AddObservation(template, r.Method, time.Since(startedAt).Seconds())
		}
	})
}
