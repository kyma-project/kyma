package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/histogram/mocks"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/latency"
)

func TestNewCollector(t *testing.T) {
	// given
	latency := new(mocks.BucketsProvider)
	latency.On("Buckets").Return(nil)

	// when
	collector := NewCollector(latency)

	// then
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.backendLatency)
	assert.NotNil(t, collector.backendLatency.MetricVec)
	assert.NotNil(t, collector.eventType)
	assert.NotNil(t, collector.eventType.MetricVec)
	latency.AssertExpectations(t)
}

//nolint: lll // prometheus tef follows
func TestCollector_MetricsMiddleware(t *testing.T) {
	router := mux.NewRouter()
	c := NewCollector(latency.BucketsProvider{})
	router.Use(c.MetricsMiddleware())
	router.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(6 * time.Millisecond)
		writer.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(router)
	defer srv.Close()
	http.Get(srv.URL + "/test") //nolint: errcheck // this call never fails as it is a testserver
	tef := `
	# HELP eventing_epp_requests_duration_seconds The duration of processing an incoming request (includes sending to the backend)
	# TYPE eventing_epp_requests_duration_seconds histogram
	eventing_epp_requests_duration_seconds_bucket{code="200",method="get",path="/test",le="0.005"} 0
	eventing_epp_requests_duration_seconds_bucket{code="200",method="get",path="/test",le="0.01"} 1
	eventing_epp_requests_duration_seconds_bucket{code="200",method="get",path="/test",le="0.02"} 1
	eventing_epp_requests_duration_seconds_bucket{code="200",method="get",path="/test",le="0.05"} 1
	eventing_epp_requests_duration_seconds_bucket{code="200",method="get",path="/test",le="0.1"} 1
	eventing_epp_requests_duration_seconds_bucket{code="200",method="get",path="/test",le="0.25"} 1
	eventing_epp_requests_duration_seconds_bucket{code="200",method="get",path="/test",le="0.5"} 1
	eventing_epp_requests_duration_seconds_bucket{code="200",method="get",path="/test",le="1"} 1
	eventing_epp_requests_duration_seconds_bucket{code="200",method="get",path="/test",le="2.5"} 1
	eventing_epp_requests_duration_seconds_bucket{code="200",method="get",path="/test",le="5"} 1
	eventing_epp_requests_duration_seconds_bucket{code="200",method="get",path="/test",le="+Inf"} 1
	eventing_epp_requests_duration_seconds_sum{code="200",method="get",path="/test"} 0.006829666
	eventing_epp_requests_duration_seconds_count{code="200",method="get",path="/test"} 1
	# HELP eventing_epp_requests_total The total number of requests
	# TYPE eventing_epp_requests_total counter
	eventing_epp_requests_total{code="200",method="get",path="/test"} 1
`
	if err := ignoreErr(testutil.CollectAndCompare(c, strings.NewReader(tef)), "eventing_epp_requests_duration_seconds_sum"); err != nil {
		t.Fatalf("%v", err)
	}
}

// Hack to filter out validation of the sum calculated by the metric.
func ignoreErr(err error, metric string) error {
	for _, line := range strings.Split(err.Error(), "\n") {
		if line == "--- metric output does not match expectation; want" || line == "+++ got:" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "+") ||
			strings.HasPrefix(strings.TrimSpace(line), "-") {
			if !(strings.HasPrefix(strings.TrimSpace(line), "+"+metric) ||
				strings.HasPrefix(strings.TrimSpace(line), "-"+metric)) {
				return err
			}
		}
	}
	return nil
}
