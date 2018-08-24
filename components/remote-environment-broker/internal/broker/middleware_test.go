package broker

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/broker/automock"
	"github.com/kyma-project/kyma/components/remote-environment-broker/platform/logger/spy"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOSBContextForClusterScopedBroker(t *testing.T) {
	// GIVEN
	mockBrokerFlavor := &automock.BrokerFlavorProvider{}
	defer mockBrokerFlavor.AssertExpectations(t)

	mockBrokerFlavor.On("IsClusterScoped").Return(true)
	sut := NewOsbContextMiddleware(mockBrokerFlavor, spy.NewLogDummy())
	// WHEN
	req := httptest.NewRequest(http.MethodGet, "https://core-reb.kyma-system.svc.cluster.local/v2/catalog", nil)
	rw := httptest.NewRecorder()
	// THEN
	nextCalled := false
	sut.ServeHTTP(rw, req, func(nextRw http.ResponseWriter, nextReq *http.Request) {
		nextCalled = true
		osbCtx, ex := osbContextFromContext(nextReq.Context())
		assert.True(t, ex)
		assert.True(t, osbCtx.ClusterScopedBroker)
	})
	assert.True(t, nextCalled)
}

func TestOSBContextForNsScopedBroker(t *testing.T) {
	// GIVEN
	mockBrokerFlavor := &automock.BrokerFlavorProvider{}
	defer mockBrokerFlavor.AssertExpectations(t)
	url := "http://reb-ns-for-stage.kyma-system.svc.cluster.local/v2/catalog"

	mockBrokerFlavor.On("IsClusterScoped").Return(false)

	mockBrokerFlavor.On("GetNsFromBrokerURL", "reb-ns-for-stage.kyma-system.svc.cluster.local").Return("stage", nil)

	sut := NewOsbContextMiddleware(mockBrokerFlavor, spy.NewLogDummy())

	// WHEN
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rw := httptest.NewRecorder()
	// THEN
	nextCalled := false
	sut.ServeHTTP(rw, req, func(nextRw http.ResponseWriter, nextReq *http.Request) {
		nextCalled = true
		osbCtx, ex := osbContextFromContext(nextReq.Context())
		assert.True(t, ex)
		assert.False(t, osbCtx.ClusterScopedBroker)
		assert.Equal(t, "stage", osbCtx.BrokerNamespace)
	})
	assert.True(t, nextCalled)

}

func TestOsbContextHandleMisconfigurationError(t *testing.T) {
	// GIVEN
	mockBrokerFlavor := &automock.BrokerFlavorProvider{}
	defer mockBrokerFlavor.AssertExpectations(t)

	mockBrokerFlavor.On("IsClusterScoped").Return(false)
	mockBrokerFlavor.On("GetNsFromBrokerURL", mock.Anything).Return("", errors.New("some error"))
	logSink := spy.NewLogSink()
	sut := NewOsbContextMiddleware(mockBrokerFlavor, logSink.Logger)
	// WHEN
	req := httptest.NewRequest(http.MethodGet, "https://core-reb.kyma-system.svc.cluster.local/v2/catalog", nil)
	rw := httptest.NewRecorder()
	// THEN
	nextCalled := false
	sut.ServeHTTP(rw, req, func(nextRw http.ResponseWriter, nextReq *http.Request) {
		nextCalled = true
	})
	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
	logSink.AssertLogged(t, logrus.ErrorLevel, "misconfiguration, broker is running as a namespace-scoped, but cannot extract namespace from request host")
}
