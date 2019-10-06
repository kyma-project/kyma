package route

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupHandlers(maxWorkers int, timeout time.Duration) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/v1/extract", NewExtractHandler(maxWorkers, timeout))
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}
