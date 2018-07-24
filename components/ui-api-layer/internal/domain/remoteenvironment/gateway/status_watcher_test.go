package gateway_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/gateway"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/gateway/automock"
)

func TestGatewayStatusWatcher_GetStatusNotServing(t *testing.T) {
	// GIVEN
	gtwLister := automock.NewGatewayServiceLister()
	gtwLister.ReturnOnGetGatewayServices([]gateway.ServiceData{
		{Host: "ec-prod.production.svc.cluster.local:8080", RemoteEnvironmentName: "ec-prod"},
	})
	// the host is not existing, set http timeout to very short period to not wait too much
	svc := gateway.NewStatusWatcher(gtwLister, time.Millisecond)
	stopCh := make(chan struct{})

	// WHEN
	svc.Refresh(stopCh)

	// THEN
	assert.Equal(t, svc.GetStatus("ec-prod"), gateway.StatusNotServing)
}

func TestGatewayStatusWatcher_GetStatusServing(t *testing.T) {
	// GIVEN
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	gtwLister := automock.NewGatewayServiceLister()
	gtwLister.ReturnOnGetGatewayServices([]gateway.ServiceData{
		{Host: u.Host, RemoteEnvironmentName: "ec-prod"},
	})

	svc := gateway.NewStatusWatcher(gtwLister, 200*time.Millisecond)
	stopCh := make(chan struct{})

	// WHEN
	svc.Refresh(stopCh)

	// THEN
	assert.Equal(t, svc.GetStatus("ec-prod"), gateway.StatusServing)
}

func TestGatewayStatusWatcher_GetStatusNotConfigured(t *testing.T) {
	// GIVEN
	gtwLister := automock.NewGatewayServiceLister()
	gtwLister.ReturnOnGetGatewayServices([]gateway.ServiceData{
		{Host: "ec-prod.production.svc.cluster.local:8080", RemoteEnvironmentName: "ec-prod"},
	})
	// the host is not existing, set http timeout to very short period to not wait too much
	svc := gateway.NewStatusWatcher(gtwLister, time.Millisecond)
	stopCh := make(chan struct{})

	// WHEN
	svc.Refresh(stopCh)

	// THEN
	assert.Equal(t, svc.GetStatus("not-existing"), gateway.StatusNotConfigured)
}
