package registry

import (
	"context"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
)

func TestNewRegistryClient(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		opts    *RegistryClientOptions
		want    *registryClient
		wantErr bool
	}{
		{
			name: "valid options",
			opts: &RegistryClientOptions{
				Username: "test-user",
				Password: "test-password",
				URL:      "docker-registry.kyma-system.svc.cluster.local:5000",
			},
			want: &registryClient{
				ctx:       ctx,
				username:  "test-user",
				password:  "test-password",
				url:       validURL(t, "http://docker-registry.kyma-system.svc.cluster.local:5000"),
				transport: validTransport(t, "test-user", "test-password", validURL(t, "http://docker-registry.kyma-system.svc.cluster.local:5000")),
			},
		},
		{
			name: "valid options with URL scheme",
			opts: &RegistryClientOptions{
				Username: "test-user",
				Password: "test-password",
				URL:      "http://docker-registry.kyma-system.svc.cluster.local:5000",
			},
			want: &registryClient{
				ctx:       ctx,
				username:  "test-user",
				password:  "test-password",
				url:       validURL(t, "http://docker-registry.kyma-system.svc.cluster.local:5000"),
				transport: validTransport(t, "test-user", "test-password", validURL(t, "http://docker-registry.kyma-system.svc.cluster.local:5000")),
			},
		},
		{
			name: "Empty username",
			opts: &RegistryClientOptions{
				Username: "",
				Password: "test-password",
				URL:      "docker-registry.kyma-system.svc.cluster.local:5000",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRegistryClient(ctx, tt.opts, logr.Discard())
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRegistryClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if we have error and we expect it, we don't need to continue checking.
			if err != nil {
				return
			}
			if !validateRegistryClient(t, got.(*registryClient), tt.want) {
				t.Errorf("NewRegistryClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func validateRegistryClient(t *testing.T, got, want *registryClient) bool {
	// we skip registryClient.regClient because it's created directly in the upstream library and it's different every time.
	if want.username != got.username ||
		want.password != got.password ||
		!reflect.DeepEqual(want.url, got.url) ||
		!reflect.DeepEqual(want.transport, got.transport) {
		return false
	}
	return true
}

func validURL(t *testing.T, s string) *url.URL {
	u, err := url.Parse(s)
	require.NoError(t, err)

	return u
}

func validTransport(t *testing.T, u, p string, l *url.URL) http.RoundTripper {
	tr, err := registryAuthTransport(u, p, l)
	require.NoError(t, err)

	return tr
}
