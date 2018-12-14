package nsbroker_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"
	"github.com/stretchr/testify/assert"
)

func TestHttpCheckerHappyPath(t *testing.T) {
	// given
	afterCh := make(chan time.Time)
	numberOfCalls := 0
	responses := make(chan int, 2)
	responses <- http.StatusInternalServerError
	responses <- http.StatusOK
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numberOfCalls = numberOfCalls + 1
		w.WriteHeader(<-responses)
	}))
	checker := nsbroker.NewHTTPChecker(func(d time.Duration) <-chan time.Time {
		return afterCh
	})

	// when
	checker.WaitUntilIsAvailable(ts.URL, 100*time.Millisecond)

	// then
	assert.Equal(t, 2, numberOfCalls)
}
