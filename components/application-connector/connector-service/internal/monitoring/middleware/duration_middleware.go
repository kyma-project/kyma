package middleware

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/monitoring/collector"
	"github.com/sirupsen/logrus"
)

type DurationMiddleware struct {
	metricsCollector collector.Collector
}

func NewDurationMiddleware(metricsCollector collector.Collector) *DurationMiddleware {
	return &DurationMiddleware{metricsCollector: metricsCollector}
}

func (dm *DurationMiddleware) Handle(next http.Handler) http.Handler {
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
			dm.metricsCollector.AddObservation(time.Since(startedAt).Seconds(), template, r.Method)
		}
	})
}
