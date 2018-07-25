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
	summaryVec *prometheus.SummaryVec
}

func NewCodeMiddleware(name, namespace string) (*codeMiddleware, apperrors.AppError) {
	summaryVec := newCodeSummaryVec(name, namespace)

	err := prometheus.Register(summaryVec)
	if err != nil {
		return nil, apperrors.Internal("Failed to create middleware %s: %s", name, err.Error())
	}

	return &codeMiddleware{summaryVec: summaryVec}, nil
}

func newCodeSummaryVec(name, namespace string) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Name:      name,
		},
		[]string{"endpoint", "code"},
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
			dm.summaryVec.WithLabelValues(template, r.Method).Observe(float64(writerWrapper.statusCode))
		}
	})
}
