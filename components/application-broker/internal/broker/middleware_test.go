package broker

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/application-broker/internal/broker/automock"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOSBContextForNsScopedBroker(t *testing.T) {
	// GIVEN
	mockBrokerService := &automock.BrokerService{}
	defer mockBrokerService.AssertExpectations(t)
	url := "http://reb-ns-for-stage.kyma-system.svc.cluster.local/v2/catalog"

	mockBrokerService.On("GetNsFromBrokerURL", "reb-ns-for-stage.kyma-system.svc.cluster.local").Return("stage", nil)

	sut := NewOsbContextMiddleware(mockBrokerService, spy.NewLogDummy())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rw := httptest.NewRecorder()
	nextCalled := false

	// WHEN
	sut.ServeHTTP(rw, req, func(nextRw http.ResponseWriter, nextReq *http.Request) {
		nextCalled = true
		osbCtx, ex := osbContextFromContext(nextReq.Context())
		// THEN
		assert.True(t, ex)
		assert.Equal(t, "stage", osbCtx.BrokerNamespace)
	})
	// THEN
	assert.True(t, nextCalled)

}

func TestOsbContextReturnsErrorWhenCannotExtractNamespace(t *testing.T) {
	// GIVEN
	mockBrokerService := &automock.BrokerService{}
	defer mockBrokerService.AssertExpectations(t)

	mockBrokerService.On("GetNsFromBrokerURL", mock.Anything).Return("", errors.New("some error"))
	logSink := spy.NewLogSink()
	sut := NewOsbContextMiddleware(mockBrokerService, logSink.Logger)
	req := httptest.NewRequest(http.MethodGet, "https://core-reb.kyma-system.svc.cluster.local/v2/catalog", nil)
	rw := httptest.NewRecorder()
	nextCalled := false
	// WHEN
	sut.ServeHTTP(rw, req, func(nextRw http.ResponseWriter, nextReq *http.Request) {
		// THEN
		nextCalled = true
	})
	// THEN
	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
	logSink.AssertLogged(t, logrus.ErrorLevel, "misconfiguration, broker is running as a namespace-scoped, but cannot extract namespace from request")
}
