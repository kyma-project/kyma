package middleware

import (
	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"net/http"
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
	histogramVec *prometheus.HistogramVec
}

func NewCodeMiddleware(name string) (*codeMiddleware, apperrors.AppError) {
	histogramVec := newCodeHistogram(name)

	err := prometheus.Register(histogramVec)
	if err != nil {
		return nil, apperrors.Internal("Failed to create middleware %s: %s", name, err.Error())
	}

	return &codeMiddleware{histogramVec: histogramVec}, nil
}

func newCodeHistogram(name string) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    "Status codes returned by each endpoint",
			Buckets: []float64{200, 201, 400, 403, 404, 409, 500, 503},
		},
		[]string{"endpoint", "method"},
	)
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
			dm.histogramVec.WithLabelValues(template, r.Method).Observe(float64(writerWrapper.statusCode))
		}
	})
}
