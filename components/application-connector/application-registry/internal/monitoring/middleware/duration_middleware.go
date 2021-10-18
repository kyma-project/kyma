package middleware

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/application-connector/application-registry/internal/monitoring/collector"
	"github.com/sirupsen/logrus"
)

type durationMiddleware struct {
	metricsCollector collector.Collector
}

func NewDurationMiddleware(metricsCollector collector.Collector) *durationMiddleware {
	return &durationMiddleware{metricsCollector: metricsCollector}
}

func (dm *durationMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()

		next.ServeHTTP(w, r)

		route := mux.CurrentRoute(r)
		if route == nil {
			logrus.Warnf("No route matched '%s' for request tracking", r.RequestURI)
		}

		template, err := route.GetPathTemplate()
		if err != nil {
			logrus.Errorf("Getting path template failed, %s", err.Error())
		} else {
			dm.metricsCollector.AddObservation(time.Since(startedAt).Seconds(), template, r.Method)
		}
	})
}
