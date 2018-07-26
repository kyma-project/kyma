package middleware

import (
	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type durationMiddleware struct {
	summaryVec *prometheus.SummaryVec
}

func NewDurationMiddleware(name string) (*durationMiddleware, apperrors.AppError) {
	summaryVec := newDurationSummaryVec(name)

	err := prometheus.Register(summaryVec)
	if err != nil {
		return nil, apperrors.Internal("Failed to create middleware %s: %s", name, err.Error())
	}

	return &durationMiddleware{summaryVec: summaryVec}, nil
}

func newDurationSummaryVec(name string) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       name,
			Help:       "Response time for each endpoint",
			Objectives: map[float64]float64{0.5: 0.05},
		},
		[]string{"endpoint", "method"},
	)
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
			dm.summaryVec.WithLabelValues(template, r.Method).Observe(time.Since(startedAt).Seconds())
		}
	})
}
