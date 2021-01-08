package broker

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	osb "github.com/kubernetes-sigs/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
)

func TestOSBContextForNsScopedBroker(t *testing.T) {
	// GIVEN
	url := "http://ab-ns-for-stage.kyma-system.svc.cluster.local/stage/v2/catalog"

	sut := &OSBContextMiddleware{}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = mux.SetURLVars(req, map[string]string{"namespace": "stage"})
	req.Header.Set(osb.APIVersionHeader, osbAPIVersion)
	req.Header.Set(osb.OriginatingIdentityHeader, osb.PlatformKubernetes)
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

func TestOSBContextPreconditionError(t *testing.T) {
	// GIVEN
	url := "http://ab-ns-for-stage.kyma-system.svc.cluster.local/stage/v2/catalog"

	sut := &OSBContextMiddleware{}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = mux.SetURLVars(req, map[string]string{"namespace": "stage"})
	rw := httptest.NewRecorder()
	nextCalled := false

	// WHEN
	sut.ServeHTTP(rw, req, func(nextRw http.ResponseWriter, nextReq *http.Request) {
		nextCalled = true
	})
	// THEN
	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusPreconditionFailed, rw.Result().StatusCode)
}
